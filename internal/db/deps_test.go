package db

import (
	"testing"
	"time"

	"github.com/baiirun/prog/internal/model"
)

func createTestItem(t *testing.T, db *DB, title string) *model.Item {
	t.Helper()
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     title,
		Status:    model.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	return item
}

func TestAddDep(t *testing.T) {
	db := setupTestDB(t)

	task1 := createTestItem(t, db, "Task 1")
	task2 := createTestItem(t, db, "Task 2")

	if err := db.AddDep(task2.ID, task1.ID); err != nil {
		t.Fatalf("failed to add dep: %v", err)
	}

	deps, err := db.GetDeps(task2.ID)
	if err != nil {
		t.Fatalf("failed to get deps: %v", err)
	}

	if len(deps) != 1 {
		t.Errorf("expected 1 dep, got %d", len(deps))
	}
	if deps[0] != task1.ID {
		t.Errorf("dep = %q, want %q", deps[0], task1.ID)
	}
}

func TestAddDep_NonexistentItem(t *testing.T) {
	db := setupTestDB(t)

	task1 := createTestItem(t, db, "Task 1")

	err := db.AddDep(task1.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent dependency")
	}

	err = db.AddDep("nonexistent", task1.ID)
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestAddDep_Duplicate(t *testing.T) {
	db := setupTestDB(t)

	task1 := createTestItem(t, db, "Task 1")
	task2 := createTestItem(t, db, "Task 2")

	db.AddDep(task2.ID, task1.ID)

	// Adding duplicate should not error (INSERT OR IGNORE)
	if err := db.AddDep(task2.ID, task1.ID); err != nil {
		t.Errorf("duplicate dep should not error: %v", err)
	}

	deps, _ := db.GetDeps(task2.ID)
	if len(deps) != 1 {
		t.Errorf("expected 1 dep after duplicate, got %d", len(deps))
	}
}

func TestHasUnmetDeps(t *testing.T) {
	db := setupTestDB(t)

	task1 := createTestItem(t, db, "Task 1")
	task2 := createTestItem(t, db, "Task 2")

	db.AddDep(task2.ID, task1.ID)

	// task1 is open, so task2 has unmet deps
	unmet, err := db.HasUnmetDeps(task2.ID)
	if err != nil {
		t.Fatalf("failed to check deps: %v", err)
	}
	if !unmet {
		t.Error("expected unmet deps when dependency is open")
	}

	// Mark task1 as done
	db.UpdateStatus(task1.ID, model.StatusDone)

	unmet, _ = db.HasUnmetDeps(task2.ID)
	if unmet {
		t.Error("expected no unmet deps when dependency is done")
	}
}

func TestHasUnmetDeps_NoDeps(t *testing.T) {
	db := setupTestDB(t)

	task := createTestItem(t, db, "Task")

	unmet, err := db.HasUnmetDeps(task.ID)
	if err != nil {
		t.Fatalf("failed to check deps: %v", err)
	}
	if unmet {
		t.Error("expected no unmet deps for task with no dependencies")
	}
}

func TestGetDeps_Empty(t *testing.T) {
	db := setupTestDB(t)

	task := createTestItem(t, db, "Task")

	deps, err := db.GetDeps(task.ID)
	if err != nil {
		t.Fatalf("failed to get deps: %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected 0 deps, got %d", len(deps))
	}
}

func TestGetAllDeps(t *testing.T) {
	db := setupTestDB(t)

	task1 := createTestItem(t, db, "Task 1")
	task2 := createTestItem(t, db, "Task 2")
	task3 := createTestItem(t, db, "Task 3")

	// task2 depends on task1, task3 depends on task1
	db.AddDep(task2.ID, task1.ID)
	db.AddDep(task3.ID, task1.ID)

	edges, err := db.GetAllDeps("")
	if err != nil {
		t.Fatalf("failed to get all deps: %v", err)
	}

	if len(edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(edges))
	}

	// Check edge details
	for _, e := range edges {
		if e.DependsOnID != task1.ID {
			t.Errorf("expected all deps on %s, got %s", task1.ID, e.DependsOnID)
		}
		if e.DependsOnTitle != "Task 1" {
			t.Errorf("expected dep title 'Task 1', got %q", e.DependsOnTitle)
		}
	}
}

func TestGetAllDeps_FilterByProject(t *testing.T) {
	db := setupTestDB(t)

	// Create tasks in "test" project (default from createTestItem)
	task1 := createTestItem(t, db, "Task 1")
	task2 := createTestItem(t, db, "Task 2")
	db.AddDep(task2.ID, task1.ID)

	// Create task in different project
	otherTask := &model.Item{
		ID:      model.GenerateID(model.ItemTypeTask),
		Project: "other",
		Type:    model.ItemTypeTask,
		Title:   "Other Task",
		Status:  model.StatusOpen,
	}
	db.CreateItem(otherTask)
	otherTask2 := &model.Item{
		ID:      model.GenerateID(model.ItemTypeTask),
		Project: "other",
		Type:    model.ItemTypeTask,
		Title:   "Other Task 2",
		Status:  model.StatusOpen,
	}
	db.CreateItem(otherTask2)
	db.AddDep(otherTask2.ID, otherTask.ID)

	// Filter by "test" project
	edges, err := db.GetAllDeps("test")
	if err != nil {
		t.Fatalf("failed to get deps: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("expected 1 edge for 'test' project, got %d", len(edges))
	}

	// Filter by "other" project
	edges, err = db.GetAllDeps("other")
	if err != nil {
		t.Fatalf("failed to get deps: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("expected 1 edge for 'other' project, got %d", len(edges))
	}
}

func TestGetAllDeps_Empty(t *testing.T) {
	db := setupTestDB(t)

	createTestItem(t, db, "Task 1")

	edges, err := db.GetAllDeps("")
	if err != nil {
		t.Fatalf("failed to get all deps: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(edges))
	}
}
