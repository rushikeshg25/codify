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

	// In the cluster but not in the database: orphaned, tear it down.
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

	// Claims to be running but has no Deployment: the create was lost. Mark it
	// failed rather than leaving the user on a spinner forever. Re-provisioning
	// automatically is deliberately not done here — it would resurrect
	// codegrounds a user had just stopped.
	for id, status := range known {
		if status != StatusRunning {
			continue
		}
		if _, ok := live[id]; !ok {
			log.Printf("Reconciler: codeground %s is marked running but has no deployment", id)
			setStatus(db, id, StatusFailed, "deployment is missing")
		}
	}

	return nil
}

// liveCodegrounds lists the codegrounds that actually exist in the cluster,
// reconstructed from the labels every resource carries.
func liveCodegrounds(ctx context.Context, k8s *K8s) (map[string]Codeground, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	deployments, err := k8s.clientset.AppsV1().Deployments(k8s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: managedSelector,
	})
	if err != nil {
		return nil, err
	}

	live := make(map[string]Codeground, len(deployments.Items))
	for _, d := range deployments.Items {
		id := d.Labels["codify/codeground"]
		if id == "" {
			continue
		}
		userID, err := strconv.Atoi(d.Labels["codify/user"])
		if err != nil {
			log.Printf("Reconciler: deployment %s has an unreadable user label, skipping", d.Name)
			continue
		}
		live[id] = Codeground{Id: id, UserId: userID}
	}
	return live, nil
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
