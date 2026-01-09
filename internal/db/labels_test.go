package db

import (
	"testing"
	"time"

	"github.com/baiirun/prog/internal/model"
)

func TestCreateLabel(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	label := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "bug",
		Project:   "test",
		Color:     "#ff0000",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := db.CreateLabel(label); err != nil {
		t.Fatalf("failed to create label: %v", err)
	}

	// Verify it was created
	got, err := db.GetLabelByName("test", "bug")
	if err != nil {
		t.Fatalf("failed to get label: %v", err)
	}

	if got.Name != label.Name {
		t.Errorf("name = %q, want %q", got.Name, label.Name)
	}
	if got.Project != label.Project {
		t.Errorf("project = %q, want %q", got.Project, label.Project)
	}
	if got.Color != label.Color {
		t.Errorf("color = %q, want %q", got.Color, label.Color)
	}
}

func TestGetLabel(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	label := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "feature",
		Project:   "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateLabel(label); err != nil {
		t.Fatalf("failed to create label: %v", err)
	}

	// Get by ID
	got, err := db.GetLabel(label.ID)
	if err != nil {
		t.Fatalf("failed to get label by ID: %v", err)
	}
	if got.Name != "feature" {
		t.Errorf("name = %q, want %q", got.Name, "feature")
	}
}

func TestGetLabelByName_NotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := db.GetLabelByName("test", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent label")
	}
}

func TestListLabels(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	labels := []string{"alpha", "beta", "gamma"}
	for _, name := range labels {
		label := &model.Label{
			ID:        model.GenerateLabelID(),
			Name:      name,
			Project:   "test",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.CreateLabel(label); err != nil {
			t.Fatalf("failed to create label %s: %v", name, err)
		}
	}

	// Add a label in different project
	other := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "other",
		Project:   "other-project",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateLabel(other); err != nil {
		t.Fatalf("failed to create other label: %v", err)
	}

	// List should return only labels for "test" project
	got, err := db.ListLabels("test")
	if err != nil {
		t.Fatalf("failed to list labels: %v", err)
	}

	if len(got) != 3 {
		t.Errorf("got %d labels, want 3", len(got))
	}

	// Should be sorted by name
	if got[0].Name != "alpha" {
		t.Errorf("first label = %q, want %q", got[0].Name, "alpha")
	}
}

func TestRenameLabel(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	label := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "old-name",
		Project:   "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateLabel(label); err != nil {
		t.Fatalf("failed to create label: %v", err)
	}

	if err := db.RenameLabel("test", "old-name", "new-name"); err != nil {
		t.Fatalf("failed to rename label: %v", err)
	}

	// Old name should not work
	_, err := db.GetLabelByName("test", "old-name")
	if err == nil {
		t.Error("expected error for old name")
	}

	// New name should work
	got, err := db.GetLabelByName("test", "new-name")
	if err != nil {
		t.Fatalf("failed to get renamed label: %v", err)
	}
	if got.ID != label.ID {
		t.Error("renamed label has different ID")
	}
}

func TestDeleteLabel(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	label := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "to-delete",
		Project:   "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateLabel(label); err != nil {
		t.Fatalf("failed to create label: %v", err)
	}

	// Create an item and attach the label
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if err := db.AddLabelToItem(item.ID, "test", "to-delete"); err != nil {
		t.Fatalf("failed to add label to item: %v", err)
	}

	// Delete the label
	if err := db.DeleteLabel("test", "to-delete"); err != nil {
		t.Fatalf("failed to delete label: %v", err)
	}

	// Label should be gone
	_, err := db.GetLabelByName("test", "to-delete")
	if err == nil {
		t.Error("expected error for deleted label")
	}

	// Item should have no labels
	labels, err := db.GetItemLabels(item.ID)
	if err != nil {
		t.Fatalf("failed to get item labels: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("item has %d labels, want 0", len(labels))
	}
}

func TestAddLabelToItem(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add label (should create it)
	if err := db.AddLabelToItem(item.ID, "test", "new-label"); err != nil {
		t.Fatalf("failed to add label: %v", err)
	}

	// Label should exist
	label, err := db.GetLabelByName("test", "new-label")
	if err != nil {
		t.Fatalf("label was not created: %v", err)
	}

	// Item should have the label
	labels, err := db.GetItemLabels(item.ID)
	if err != nil {
		t.Fatalf("failed to get item labels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("got %d labels, want 1", len(labels))
	}
	if labels[0].ID != label.ID {
		t.Error("wrong label attached to item")
	}
}

func TestAddLabelToItem_Idempotent(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add same label twice
	if err := db.AddLabelToItem(item.ID, "test", "bug"); err != nil {
		t.Fatalf("first add failed: %v", err)
	}
	if err := db.AddLabelToItem(item.ID, "test", "bug"); err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	// Should only have one label
	labels, err := db.GetItemLabels(item.ID)
	if err != nil {
		t.Fatalf("failed to get item labels: %v", err)
	}
	if len(labels) != 1 {
		t.Errorf("got %d labels, want 1", len(labels))
	}
}

func TestRemoveLabelFromItem(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add then remove
	if err := db.AddLabelToItem(item.ID, "test", "temp"); err != nil {
		t.Fatalf("failed to add label: %v", err)
	}
	if err := db.RemoveLabelFromItem(item.ID, "test", "temp"); err != nil {
		t.Fatalf("failed to remove label: %v", err)
	}

	// Item should have no labels
	labels, err := db.GetItemLabels(item.ID)
	if err != nil {
		t.Fatalf("failed to get item labels: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("got %d labels, want 0", len(labels))
	}

	// Label should still exist (just not attached)
	_, err = db.GetLabelByName("test", "temp")
	if err != nil {
		t.Error("label was deleted when it should only be detached")
	}
}

func TestGetItemLabels_MultipleLabels(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	item := &model.Item{
		ID:        model.GenerateID(model.ItemTypeTask),
		Project:   "test",
		Type:      model.ItemTypeTask,
		Title:     "Test task",
		Status:    model.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add multiple labels
	labelNames := []string{"bug", "urgent", "backend"}
	for _, name := range labelNames {
		if err := db.AddLabelToItem(item.ID, "test", name); err != nil {
			t.Fatalf("failed to add label %s: %v", name, err)
		}
	}

	// Get labels
	labels, err := db.GetItemLabels(item.ID)
	if err != nil {
		t.Fatalf("failed to get item labels: %v", err)
	}

	if len(labels) != 3 {
		t.Fatalf("got %d labels, want 3", len(labels))
	}

	// Should be sorted by name
	if labels[0].Name != "backend" {
		t.Errorf("first label = %q, want %q", labels[0].Name, "backend")
	}
	if labels[1].Name != "bug" {
		t.Errorf("second label = %q, want %q", labels[1].Name, "bug")
	}
	if labels[2].Name != "urgent" {
		t.Errorf("third label = %q, want %q", labels[2].Name, "urgent")
	}
}

func TestSetLabelColor(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	label := &model.Label{
		ID:        model.GenerateLabelID(),
		Name:      "colorful",
		Project:   "test",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.CreateLabel(label); err != nil {
		t.Fatalf("failed to create label: %v", err)
	}

	if err := db.SetLabelColor("test", "colorful", "#00ff00"); err != nil {
		t.Fatalf("failed to set color: %v", err)
	}

	got, err := db.GetLabelByName("test", "colorful")
	if err != nil {
		t.Fatalf("failed to get label: %v", err)
	}
	if got.Color != "#00ff00" {
		t.Errorf("color = %q, want %q", got.Color, "#00ff00")
	}
}

func TestEnsureLabel(t *testing.T) {
	db := setupTestDB(t)

	// First call creates
	label1, err := db.EnsureLabel("test", "ensured")
	if err != nil {
		t.Fatalf("first ensure failed: %v", err)
	}

	// Second call returns existing
	label2, err := db.EnsureLabel("test", "ensured")
	if err != nil {
		t.Fatalf("second ensure failed: %v", err)
	}

	if label1.ID != label2.ID {
		t.Error("ensure created duplicate label")
	}
}
