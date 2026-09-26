package main

import (
	"context"
	"database/sql"
	"log"
	"time"
)

/*
The reaper stops idle codegrounds from running forever.

A codeground lived until somebody explicitly deleted it, so a tab closed and
forgotten kept its full CPU and memory reservation indefinitely. At any real
usage level that is the dominant cost of running the platform, and none of it
is doing anything for anyone.

Idleness is measured by last_active_at, which the browser refreshes while the
codeground is open. A reaped codeground is marked STOPPED rather than deleted:
the row survives, so the user can see what happened and start it again.
*/
func StartReaper(ctx context.Context, db *sql.DB, k8s *K8s) {
	interval := durationOr("REAP_INTERVAL", 5*time.Minute)
	idleAfter := durationOr("CODEGROUND_IDLE_TTL", 30*time.Minute)
	log.Printf("Reaper starting, interval %s, idle after %s", interval, idleAfter)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := reapOnce(db, k8s, idleAfter); err != nil {
					log.Printf("Reap failed: %v", err)
				}
			}
		}
	}()
}

func reapOnce(db *sql.DB, k8s *K8s, idleAfter time.Duration) error {
	rows, err := db.Query(
		`SELECT id, user_id FROM codegrounds
		 WHERE status = ? AND last_active_at < (NOW() - INTERVAL ? SECOND)`,
		StatusRunning, int(idleAfter.Seconds()),
	)
	if err != nil {
		return err
	}

	// Collect before tearing down: holding the rows open across a sequence of
	// slow Kubernetes calls would pin a connection for the whole sweep.
	var idle []Codeground
	for rows.Next() {
		var cg Codeground
		if err := rows.Scan(&cg.Id, &cg.UserId); err != nil {
			rows.Close()
			return err
		}
		idle = append(idle, cg)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	for _, cg := range idle {
		log.Printf("Reaper: codeground %s idle for over %s, stopping", cg.Id, idleAfter)
		cg.K8s = k8s
		if err := cg.DeleteCodeground(); err != nil {
			log.Printf("Reaper: failed to stop idle codeground %s: %v", cg.Id, err)
			continue
		}
		setStatus(db, cg.Id, StatusStopped, "stopped automatically after inactivity")
	}

	return nil
}
