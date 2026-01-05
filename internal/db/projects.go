package db

import (
	"fmt"
	"time"
)

// EnsureProject creates a project if it doesn't exist.
// This is idempotent - calling it multiple times with the same name is safe.
func (db *DB) EnsureProject(name string) error {
	_, err := db.Exec(`
		INSERT INTO projects (name, created_at, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO NOTHING`,
		name, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to ensure project: %w", err)
	}
	return nil
}
