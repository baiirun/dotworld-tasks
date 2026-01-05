package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/baiirun/prog/internal/model"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Should create parent directories
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("failed to get default path: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}

	if !contains(path, ".prog/prog.db") {
		t.Errorf("expected path to contain .prog/prog.db, got %q", path)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr))
}

func TestCreateItem(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		Priority:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Verify it was created
	got, err := db.GetItem(item.ID)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if got.Title != item.Title {
		t.Errorf("title = %q, want %q", got.Title, item.Title)
	}
	if got.Project != item.Project {
		t.Errorf("project = %q, want %q", got.Project, item.Project)
	}
}

func TestCreateItem_InvalidType(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:      "ts-123456",
		Project: "test",
		Type:    model.ItemType("invalid"),
		Title:   "Test",
		Status:  model.StatusOpen,
	}

	err := db.CreateItem(item)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestCreateItem_InvalidStatus(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:      "ts-123456",
		Project: "test",
		Type:    model.ItemTypeTask,
		Title:   "Test",
		Status:  model.Status("invalid"),
	}

	err := db.CreateItem(item)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestGetItem_NotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.GetItem("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestUpdateStatus(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(item)

	if err := db.UpdateStatus(item.ID, model.StatusInProgress); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	got, _ := db.GetItem(item.ID)
	if got.Status != model.StatusInProgress {
		t.Errorf("status = %q, want %q", got.Status, model.StatusInProgress)
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	db := setupTestDB(t)

	err := db.UpdateStatus("nonexistent", model.StatusDone)
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	db := setupTestDB(t)

	err := db.UpdateStatus("ts-123456", model.Status("invalid"))
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestAppendDescription(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:          model.GenerateID(model.ItemTypeTask),
		Project:     "test",
		Type:        model.ItemTypeTask,
		Title:       "Test",
		Description: "Initial",
		Status:      model.StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.CreateItem(item)

	if err := db.AppendDescription(item.ID, "Appended text"); err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	got, _ := db.GetItem(item.ID)
	if got.Description == "Initial" {
		t.Error("description was not appended")
	}
}

func TestSetParent(t *testing.T) {
	db := setupTestDB(t)

	epic := &model.Item{
		ID:        model.GenerateID(model.ItemTypeEpic),
		Project:   "test",
		Type:      model.ItemTypeEpic,
		Title:     "Test Epic",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(epic)

	task := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test Task",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(task)

	if err := db.SetParent(task.ID, epic.ID); err != nil {
		t.Fatalf("failed to set parent: %v", err)
	}

	got, _ := db.GetItem(task.ID)
	if got.ParentID == nil {
		t.Fatal("expected parent ID to be set")
	}
	if *got.ParentID != epic.ID {
		t.Errorf("parent = %q, want %q", *got.ParentID, epic.ID)
	}
}

func TestSetParent_NotEpic(t *testing.T) {
	db := setupTestDB(t)

	task1 := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Task 1",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(task1)

	task2 := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Task 2",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(task2)

	err := db.SetParent(task2.ID, task1.ID)
	if err == nil {
		t.Error("expected error when parent is not an epic")
	}
}

func TestSetDescription(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:          model.GenerateID(model.ItemTypeTask),
		Project:     "test",
		Type:        model.ItemTypeTask,
		Title:       "Test",
		Description: "Original description",
		Status:      model.StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.CreateItem(item)

	if err := db.SetDescription(item.ID, "New description"); err != nil {
		t.Fatalf("failed to set description: %v", err)
	}

	got, _ := db.GetItem(item.ID)
	if got.Description != "New description" {
		t.Errorf("description = %q, want %q", got.Description, "New description")
	}
}

func TestSetDescription_EmptyToContent(t *testing.T) {
	db := setupTestDB(t)

	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(item)

	if err := db.SetDescription(item.ID, "Added description"); err != nil {
		t.Fatalf("failed to set description: %v", err)
	}

	got, _ := db.GetItem(item.ID)
	if got.Description != "Added description" {
		t.Errorf("description = %q, want %q", got.Description, "Added description")
	}
}

func TestSetDescription_NotFound(t *testing.T) {
	db := setupTestDB(t)

	err := db.SetDescription("nonexistent", "text")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestSetParent_NotFound(t *testing.T) {
	db := setupTestDB(t)

	epic := &model.Item{
		ID:        model.GenerateID(model.ItemTypeEpic),
		Project:   "test",
		Type:      model.ItemTypeEpic,
		Title:     "Epic",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(epic)

	// Nonexistent task
	err := db.SetParent("nonexistent", epic.ID)
	if err == nil {
		t.Error("expected error for nonexistent task")
	}

	// Nonexistent parent
	task := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Task",
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.CreateItem(task)

	err = db.SetParent(task.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent parent")
	}
}
