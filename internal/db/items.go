package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/baiirun/dotworld-tasks/internal/model"
)

// CreateItem inserts a new item into the database.
func (db *DB) CreateItem(item *model.Item) error {
	if !item.Type.IsValid() {
		return fmt.Errorf("invalid item type: %s", item.Type)
	}
	if !item.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", item.Status)
	}

	_, err := db.Exec(`
		INSERT INTO items (id, project, type, title, description, status, priority, parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Project, item.Type, item.Title, item.Description,
		item.Status, item.Priority, item.ParentID, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// GetItem retrieves an item by ID.
func (db *DB) GetItem(id string) (*model.Item, error) {
	row := db.QueryRow(`
		SELECT id, project, type, title, description, status, priority, parent_id, created_at, updated_at
		FROM items WHERE id = ?`, id)

	item := &model.Item{}
	var parentID sql.NullString
	err := row.Scan(
		&item.ID, &item.Project, &item.Type, &item.Title, &item.Description,
		&item.Status, &item.Priority, &parentID, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if parentID.Valid {
		item.ParentID = &parentID.String
	}
	return item, nil
}

// UpdateStatus changes an item's status.
func (db *DB) UpdateStatus(id string, status model.Status) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s", status)
	}

	result, err := db.Exec(`
		UPDATE items SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found: %s", id)
	}
	return nil
}

// AppendDescription appends text to an item's description.
func (db *DB) AppendDescription(id string, text string) error {
	result, err := db.Exec(`
		UPDATE items
		SET description = COALESCE(description, '') || ? || char(10) || ?,
		    updated_at = ?
		WHERE id = ?`,
		"\n", text, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to append description: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found: %s", id)
	}
	return nil
}
