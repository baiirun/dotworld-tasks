package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnboard_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	// Check output message
	if !strings.Contains(output, "Created CLAUDE.md") {
		t.Errorf("expected 'Created CLAUDE.md' message, got: %s", output)
	}

	// Check file was created with correct content
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	if !strings.Contains(string(content), "## Task Tracking") {
		t.Error("missing '## Task Tracking' header")
	}
	if !strings.Contains(string(content), "prog prime") {
		t.Error("missing 'prog prime' reference")
	}
	if !strings.Contains(string(content), "prog ready") {
		t.Error("missing 'prog ready' command")
	}
}

func TestOnboard_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Create existing file
	existingContent := "# My Project\n\nSome existing content.\n"
	if err := os.WriteFile(claudePath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing CLAUDE.md: %v", err)
	}

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	// Check output message includes filename
	if !strings.Contains(output, "Added prog integration to CLAUDE.md") {
		t.Errorf("expected 'Added prog integration to CLAUDE.md' message, got: %s", output)
	}

	// Check file has both old and new content
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	if !strings.Contains(string(content), "# My Project") {
		t.Error("missing original content")
	}
	if !strings.Contains(string(content), "## Task Tracking") {
		t.Error("missing appended Task Tracking section")
	}
}

func TestOnboard_Idempotent(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Create file that already has Task Tracking section
	existingContent := "# My Project\n\n## Task Tracking\n\nAlready configured.\n"
	if err := os.WriteFile(claudePath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing CLAUDE.md: %v", err)
	}

	// Pre-install the hook so both CLAUDE.md and hook are "already done"
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		t.Fatalf("failed to create settings dir: %v", err)
	}
	existingSettings := `{"hooks":{"SessionStart":[{"matcher":"","hooks":[{"type":"command","command":"prog prime"}]}]}}`
	if err := os.WriteFile(settingsPath, []byte(existingSettings), 0644); err != nil {
		t.Fatalf("failed to write existing settings.json: %v", err)
	}

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	// Check output messages for both components
	if !strings.Contains(output, "CLAUDE.md already has Task Tracking section") {
		t.Errorf("expected CLAUDE.md already configured message, got: %s", output)
	}
	if !strings.Contains(output, "SessionStart hook already installed") {
		t.Errorf("expected hook already installed message, got: %s", output)
	}

	// Check file wasn't modified
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	if string(content) != existingContent {
		t.Error("file was modified when it should have been left alone")
	}
}

func TestOnboard_LowercaseFile(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "claude.md") // lowercase
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Create existing lowercase file
	existingContent := "# My Project\n"
	if err := os.WriteFile(claudePath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing claude.md: %v", err)
	}

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	// Should show the actual lowercase filename
	if !strings.Contains(output, "claude.md") {
		t.Errorf("expected output to mention 'claude.md', got: %s", output)
	}

	// Should have appended to lowercase file, not created CLAUDE.md
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude.md: %v", err)
	}

	if !strings.Contains(string(content), "## Task Tracking") {
		t.Error("should have appended to existing claude.md")
	}

	// Verify no separate CLAUDE.md was created (check actual filenames)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() == "CLAUDE.md" {
			t.Error("should not have created separate CLAUDE.md")
		}
	}
}

func TestOnboard_SnippetContent(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	// Check all key commands are present
	commands := []string{
		"prog ready",
		"prog add",
		"prog start",
		"prog log",
		"prog done",
		"prog prime",
	}

	for _, cmd := range commands {
		if !strings.Contains(string(content), cmd) {
			t.Errorf("missing command reference: %s", cmd)
		}
	}
}

func TestOnboard_ForceReplacesSection(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Create file with old Task Tracking section
	oldContent := "# My Project\n\n## Task Tracking\n\nOld instructions here.\n"
	if err := os.WriteFile(claudePath, []byte(oldContent), 0644); err != nil {
		t.Fatalf("failed to write existing CLAUDE.md: %v", err)
	}

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(true, settingsPath); err != nil {
			t.Fatalf("runOnboard --force failed: %v", err)
		}
	})

	// Check output message
	if !strings.Contains(output, "Updated Task Tracking section") {
		t.Errorf("expected 'Updated Task Tracking section' message, got: %s", output)
	}

	// Check content was replaced
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	if strings.Contains(string(content), "Old instructions here") {
		t.Error("old content should have been replaced")
	}
	if !strings.Contains(string(content), "prog prime") {
		t.Error("should have new snippet content")
	}
	if !strings.Contains(string(content), "# My Project") {
		t.Error("content before section should be preserved")
	}
}

func TestOnboard_ForcePreservesContentAfterSection(t *testing.T) {
	dir := t.TempDir()
	claudePath := filepath.Join(dir, "CLAUDE.md")
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Create file with Task Tracking in the middle
	oldContent := `# My Project

## Task Tracking

Old instructions.

## Other Section

This should be preserved.
`
	if err := os.WriteFile(claudePath, []byte(oldContent), 0644); err != nil {
		t.Fatalf("failed to write existing CLAUDE.md: %v", err)
	}

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	captureOutput(func() {
		if err := runOnboardWithSettings(true, settingsPath); err != nil {
			t.Fatalf("runOnboard --force failed: %v", err)
		}
	})

	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	// Check that content after the section is preserved
	if !strings.Contains(string(content), "## Other Section") {
		t.Error("section after Task Tracking should be preserved")
	}
	if !strings.Contains(string(content), "This should be preserved") {
		t.Error("content after Task Tracking should be preserved")
	}
	// Check Task Tracking was updated
	if !strings.Contains(string(content), "prog prime") {
		t.Error("Task Tracking section should have new content")
	}
}

func TestReplaceTaskTrackingSection(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		snippet  string
		expected string
	}{
		{
			name:     "section at end",
			content:  "# Project\n\n## Task Tracking\n\nOld stuff.\n",
			snippet:  "## Task Tracking\n\nNew stuff.\n",
			expected: "# Project\n\n## Task Tracking\n\nNew stuff.\n",
		},
		{
			name:     "section in middle",
			content:  "# Project\n\n## Task Tracking\n\nOld.\n\n## Other\n\nKeep this.\n",
			snippet:  "## Task Tracking\n\nNew.\n",
			expected: "# Project\n\n## Task Tracking\n\nNew.\n\n## Other\n\nKeep this.\n",
		},
		{
			name:     "section only",
			content:  "## Task Tracking\n\nOld.\n",
			snippet:  "## Task Tracking\n\nNew.\n",
			expected: "## Task Tracking\n\nNew.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceTaskTrackingSection(tt.content, tt.snippet)
			if result != tt.expected {
				t.Errorf("got:\n%q\nwant:\n%q", result, tt.expected)
			}
		})
	}
}

// Tests for SessionStart hook installation

func TestInstallSessionStartHook_CreatesSettings(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	added, err := installSessionStartHook(settingsPath)
	if err != nil {
		t.Fatalf("installSessionStartHook failed: %v", err)
	}

	if !added {
		t.Error("expected hook to be added")
	}

	// Verify file was created with correct structure
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Verify hook structure
	if len(settings.Hooks["SessionStart"]) != 1 {
		t.Fatalf("expected 1 SessionStart matcher, got %d", len(settings.Hooks["SessionStart"]))
	}

	matcher := settings.Hooks["SessionStart"][0]
	if matcher.Matcher != "" {
		t.Errorf("expected empty matcher, got %q", matcher.Matcher)
	}

	if len(matcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(matcher.Hooks))
	}

	hook := matcher.Hooks[0]
	if hook.Type != "command" {
		t.Errorf("expected type 'command', got %q", hook.Type)
	}
	if hook.Command != "prog prime" {
		t.Errorf("expected command 'prog prime', got %q", hook.Command)
	}
}

func TestInstallSessionStartHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// First install
	added1, err := installSessionStartHook(settingsPath)
	if err != nil {
		t.Fatalf("first installSessionStartHook failed: %v", err)
	}
	if !added1 {
		t.Error("expected first call to add hook")
	}

	// Second install should be idempotent
	added2, err := installSessionStartHook(settingsPath)
	if err != nil {
		t.Fatalf("second installSessionStartHook failed: %v", err)
	}
	if added2 {
		t.Error("expected second call to not add hook")
	}

	// Verify only one hook exists
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Count total hooks
	totalHooks := 0
	for _, matcher := range settings.Hooks["SessionStart"] {
		totalHooks += len(matcher.Hooks)
	}

	if totalHooks != 1 {
		t.Errorf("expected 1 hook, got %d", totalHooks)
	}
}

func TestInstallSessionStartHook_MergesWithExisting(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	settingsPath := filepath.Join(settingsDir, "settings.json")

	// Create existing settings with other hooks
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("failed to create settings dir: %v", err)
	}

	existingSettings := ClaudeSettings{
		Hooks: map[string][]HookMatcher{
			"SessionStart": {
				{
					Matcher: "",
					Hooks: []Hook{
						{Type: "command", Command: "bd prime"},
					},
				},
			},
			"PreCompact": {
				{
					Matcher: "",
					Hooks: []Hook{
						{Type: "command", Command: "other-hook"},
					},
				},
			},
		},
		EnabledPlugins: map[string]bool{
			"some-plugin": true,
		},
	}

	data, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal existing settings: %v", err)
	}
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("failed to write existing settings: %v", err)
	}

	// Install hook
	added, err := installSessionStartHook(settingsPath)
	if err != nil {
		t.Fatalf("installSessionStartHook failed: %v", err)
	}
	if !added {
		t.Error("expected hook to be added")
	}

	// Verify merged settings
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(content, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Check SessionStart has both hooks
	sessionStart := settings.Hooks["SessionStart"]
	if len(sessionStart) != 1 {
		t.Fatalf("expected 1 SessionStart matcher, got %d", len(sessionStart))
	}

	hooks := sessionStart[0].Hooks
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks in SessionStart, got %d", len(hooks))
	}

	// Verify both hooks are present
	foundBdPrime := false
	foundTasksPrime := false
	for _, h := range hooks {
		if h.Command == "bd prime" {
			foundBdPrime = true
		}
		if h.Command == "prog prime" {
			foundTasksPrime = true
		}
	}

	if !foundBdPrime {
		t.Error("lost existing 'bd prime' hook")
	}
	if !foundTasksPrime {
		t.Error("missing 'prog prime' hook")
	}

	// Check PreCompact is preserved
	if len(settings.Hooks["PreCompact"]) != 1 {
		t.Error("PreCompact hooks should be preserved")
	}

	// Check plugins are preserved
	if !settings.EnabledPlugins["some-plugin"] {
		t.Error("enabledPlugins should be preserved")
	}
}

func TestInstallSessionStartHook_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	settingsPath := filepath.Join(settingsDir, "settings.json")

	// Create settings with prog prime already installed
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		t.Fatalf("failed to create settings dir: %v", err)
	}

	existingSettings := ClaudeSettings{
		Hooks: map[string][]HookMatcher{
			"SessionStart": {
				{
					Matcher: "",
					Hooks: []Hook{
						{Type: "command", Command: "prog prime"},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal existing settings: %v", err)
	}
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("failed to write existing settings: %v", err)
	}

	// Try to install - should detect existing
	added, err := installSessionStartHook(settingsPath)
	if err != nil {
		t.Fatalf("installSessionStartHook failed: %v", err)
	}
	if added {
		t.Error("should not add hook when already present")
	}
}

func TestOnboard_InstallsHook(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")

	// Change to temp dir for the test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output := captureOutput(func() {
		if err := runOnboardWithSettings(false, settingsPath); err != nil {
			t.Fatalf("runOnboard failed: %v", err)
		}
	})

	// Check hook installation message
	if !strings.Contains(output, "Installed SessionStart hook") {
		t.Errorf("expected hook installation message, got: %s", output)
	}

	// Verify settings file was created
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	if !strings.Contains(string(content), "prog prime") {
		t.Error("settings.json should contain 'prog prime' command")
	}
}
