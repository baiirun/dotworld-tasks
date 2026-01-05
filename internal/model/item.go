package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateID returns a new ID with a type-specific prefix and 6 hex chars.
//
// Prefixes by item type:
//   - task: "ts-" (e.g., ts-a1b2c3)
//   - epic: "ep-" (e.g., ep-a1b2c3)
func GenerateID(itemType ItemType) string {
	prefix := "ts-"
	if itemType == ItemTypeEpic {
		prefix = "ep-"
	}
	b := make([]byte, 3)
	rand.Read(b)
	return prefix + hex.EncodeToString(b)
}

type ItemType string

const (
	ItemTypeTask ItemType = "task"
	ItemTypeEpic ItemType = "epic"
)

func (t ItemType) IsValid() bool {
	return t == ItemTypeTask || t == ItemTypeEpic
}

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

func (s Status) IsValid() bool {
	return s == StatusOpen || s == StatusInProgress || s == StatusBlocked || s == StatusDone
}

type Item struct {
	ID          string
	Project     string
	Type        ItemType
	Title       string
	Description string
	Status      Status
	Priority    int
	ParentID    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Log struct {
	ID        int64
	ItemID    string
	Message   string
	CreatedAt time.Time
}

type Dep struct {
	ItemID    string
	DependsOn string
}
