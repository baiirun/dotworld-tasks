package db

import (
	"fmt"

	"github.com/baiirun/prog/internal/model"
)

// AddLog adds a log entry to an item.
func (db *DB) AddLog(itemID, message string) error {
	_, err := db.Exec(`
		INSERT INTO logs (item_id, message) VALUES (?, ?)`,
		itemID, message)
	if err != nil {
		return fmt.Errorf("failed to add log: %w", err)
	}
	return nil
}

// GetLogs retrieves all logs for an item, ordered by creation time.
func (db *DB) GetLogs(itemID string) ([]model.Log, error) {
	rows, err := db.Query(`
		SELECT id, item_id, message, created_at
		FROM logs WHERE item_id = ? ORDER BY created_at ASC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var logs []model.Log
	for rows.Next() {
		var log model.Log
		if err := rows.Scan(&log.ID, &log.ItemID, &log.Message, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
