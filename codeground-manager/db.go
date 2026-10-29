package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
The manager needs its own view of the database.

Fire-and-forget queue messages cannot express "what should exist right now",
which is what both the reconciler and the reaper have to answer. Reading the
codegrounds table directly gives the manager desired state to compare reality
against, and somewhere to write back what actually happened.
*/
func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true",
			envOr("BLUEPRINT_DB_USERNAME", "root"),
			os.Getenv("BLUEPRINT_DB_PASSWORD"),
			envOr("BLUEPRINT_DB_HOST", "localhost"),
			envOr("BLUEPRINT_DB_PORT", "3306"),
			envOr("BLUEPRINT_DB_DATABASE", "codify"),
		)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// Status values mirror the enum in init.sql.
const (
	StatusProvisioning = "PROVISIONING"
	StatusRunning      = "RUNNING"
	StatusStopped      = "STOPPED"
	StatusFailed       = "FAILED"
)

// setStatus records where a codeground got to. A failure to write status is
// logged rather than propagated: it must never turn a successful provision
// into a reported failure.
func setStatus(db *sql.DB, codegroundID, status, detail string) {
	if db == nil {
		return
	}
	if _, err := db.Exec(
		"UPDATE codegrounds SET status = ?, status_detail = ? WHERE id = ?",
		status, truncateDetail(detail), codegroundID,
	); err != nil {
		log.Printf("Failed to record status %s for codeground %s: %v", status, codegroundID, err)
	}
}

// truncateDetail keeps the message inside the column width.
func truncateDetail(detail string) string {
	const max = 255
	if len(detail) > max {
		return detail[:max]
	}
	return detail
}
