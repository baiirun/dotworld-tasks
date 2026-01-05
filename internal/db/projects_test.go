package db

import (
	"testing"
)

func TestEnsureProject(t *testing.T) {
	db := setupTestDB(t)

	// First call should create the project
	err := db.EnsureProject("myproject")
	if err != nil {
		t.Fatalf("failed to ensure project: %v", err)
	}

	// Second call should be idempotent (no error)
	err = db.EnsureProject("myproject")
	if err != nil {
		t.Fatalf("failed on second ensure: %v", err)
	}

	// Project should appear in list
	projects, err := db.ListProjects()
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}

	if len(projects) != 1 || projects[0] != "myproject" {
		t.Errorf("expected [myproject], got %v", projects)
	}
}

func TestListProjectsEmpty(t *testing.T) {
	db := setupTestDB(t)

	projects, err := db.ListProjects()
	if err != nil {
		t.Fatalf("failed to list projects: %v", err)
	}

	if len(projects) != 0 {
		t.Errorf("expected empty list, got %v", projects)
	}
}
