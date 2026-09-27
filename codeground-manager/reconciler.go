package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
The reconciler is the backstop for everything the queue cannot guarantee.

Queue messages are instructions, not state: if one is lost, or the manager dies
halfway through a create, or a delete is published but the rows and resources
disagree, nothing notices. Resources named after a codeground that no longer
exists are invisible — nothing lists them, nothing bills them, and they keep
holding a node's memory until somebody goes looking by hand.

So rather than trying to make every individual message perfectly reliable, this
loop periodically compares what exists in the cluster against what the database
says should exist, and settles the difference. That makes every other path
allowed to fail: a lost create surfaces as FAILED, and a half-finished teardown
gets finished here.
*/
func StartReconciler(ctx context.Context, db *sql.DB, k8s *K8s) {
	interval := durationOr("RECONCILE_INTERVAL", 2*time.Minute)
	log.Printf("Reconciler starting, interval %s", interval)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := reconcileOnce(ctx, db, k8s); err != nil {
					log.Printf("Reconcile failed: %v", err)
				}
			}
		}
	}()
}

func reconcileOnce(ctx context.Context, db *sql.DB, k8s *K8s) error {
	live, err := liveCodegrounds(ctx, k8s)
	if err != nil {
		return err
	}

	known, err := knownCodegrounds(db)
	if err != nil {
		return err
	}

	// In the cluster but not in the database: orphaned, tear it down. This
	// includes volumes whose Deployment is already gone, which would otherwise
	// go on costing storage with nothing left referencing them.
	for id, cg := range live {
		if _, ok := known[id]; ok {
			continue
		}
		log.Printf("Reconciler: codeground %s has no database row, tearing down", id)
		cg.K8s = k8s
		if err := cg.DeleteCodeground(); err != nil {
			log.Printf("Reconciler: failed to tear down orphan %s: %v", id, err)
		}
	}

	// In the database but drifted from what the cluster is actually doing.
	for id, status := range known {
		cg, ok := live[id]

		if !ok || !cg.HasDeployment {
			// The create was lost. Mark it failed rather than leaving the user
			// on a spinner forever. Re-provisioning automatically is
			// deliberately not done: the volume may still exist and the user
			// can press start, which reattaches it.
			if status == StatusRunning {
				log.Printf("Reconciler: codeground %s is marked running but has no deployment", id)
				setStatus(db, id, StatusFailed, "deployment is missing; start it again to recreate it")
			}
			continue
		}

		switch {
		case status == StatusRunning && cg.Replicas == 0:
			// Scaled down outside the manager, or a stop whose status write
			// did not land.
			log.Printf("Reconciler: codeground %s is scaled to zero, marking stopped", id)
			setStatus(db, id, StatusStopped, "found stopped")

		case status == StatusStopped && cg.Replicas > 0:
			// Still burning CPU and memory while the database believes it is
			// stopped, so nothing else will ever shut it down.
			log.Printf("Reconciler: codeground %s is marked stopped but still running, scaling down", id)
			cg.K8s = k8s
			if err := cg.StopCodeground(); err != nil {
				log.Printf("Reconciler: failed to scale down %s: %v", id, err)
			}
		}
	}

	return nil
}

// liveCodeground is what the cluster currently holds for one codeground.
type liveCodeground struct {
	Codeground
	// HasDeployment distinguishes a stopped codeground, which still has a
	// Deployment at zero replicas, from one whose Deployment is gone and only
	// a volume remains.
	HasDeployment bool
	Replicas      int32
}

/*
liveCodegrounds lists what the cluster actually holds, reconstructed from the
labels every resource carries.

Volumes are listed as well as Deployments. A codeground whose Deployment was
deleted but whose volume survived would otherwise be invisible here, and an
orphaned volume goes on costing storage indefinitely with nothing referencing
it.
*/
func liveCodegrounds(ctx context.Context, k8s *K8s) (map[string]liveCodeground, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	live := make(map[string]liveCodeground)

	deployments, err := k8s.clientset.AppsV1().Deployments(k8s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: managedSelector,
	})
	if err != nil {
		return nil, err
	}
	for _, d := range deployments.Items {
		cg, ok := codegroundFromLabels(d.Labels, d.Name)
		if !ok {
			continue
		}
		var replicas int32
		if d.Spec.Replicas != nil {
			replicas = *d.Spec.Replicas
		}
		live[cg.Id] = liveCodeground{Codeground: cg, HasDeployment: true, Replicas: replicas}
	}

	pvcs, err := k8s.clientset.CoreV1().PersistentVolumeClaims(k8s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: managedSelector,
	})
	if err != nil {
		return nil, err
	}
	for _, p := range pvcs.Items {
		cg, ok := codegroundFromLabels(p.Labels, p.Name)
		if !ok {
			continue
		}
		if _, exists := live[cg.Id]; !exists {
			live[cg.Id] = liveCodeground{Codeground: cg}
		}
	}

	return live, nil
}

// codegroundFromLabels rebuilds a codeground's identity from the labels the
// manager stamps on everything it creates.
func codegroundFromLabels(labels map[string]string, resourceName string) (Codeground, bool) {
	id := labels["codify/codeground"]
	if id == "" {
		return Codeground{}, false
	}
	userID, err := strconv.Atoi(labels["codify/user"])
	if err != nil {
		log.Printf("Reconciler: %s has an unreadable user label, skipping", resourceName)
		return Codeground{}, false
	}
	return Codeground{Id: id, UserId: userID}, true
}

// knownCodegrounds returns the codegrounds the database says should exist,
// mapped to their current status.
func knownCodegrounds(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT id, status FROM codegrounds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	known := make(map[string]string)
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			return nil, err
		}
		known[id] = status
	}
	return known, rows.Err()
}

// durationOr reads a Go duration from the environment, falling back when unset
// or unparseable.
func durationOr(key string, fallback time.Duration) time.Duration {
	if raw := os.Getenv(key); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			return d
		}
		log.Printf("Invalid %s, using %s", key, fallback)
	}
	return fallback
}
