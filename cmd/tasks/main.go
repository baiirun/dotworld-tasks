package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/baiirun/dotworld-tasks/internal/db"
	"github.com/baiirun/dotworld-tasks/internal/model"
	"github.com/spf13/cobra"
)

var (
	flagProject  string
	flagStatus   string
	flagEpic     bool
	flagPriority int
	flagDependsOn string
)

func openDB() (*db.DB, error) {
	path, err := db.DefaultPath()
	if err != nil {
		return nil, err
	}
	database, err := db.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%w (try running 'tasks init' first)", err)
	}
	return database, nil
}

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Lightweight task management for agents",
	Long: `A CLI for managing tasks, epics, and dependencies.
Designed for AI agents to track work across sessions.

Database: ~/.world/tasks/tasks.db

Quick start:
  tasks init
  tasks add "Build feature X" -p myproject
  tasks ready -p myproject
  tasks start <id>
  tasks done <id>`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the tasks database",
	Long:  "Creates the database at ~/.world/tasks/tasks.db if it doesn't exist.",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.Init(); err != nil {
			return err
		}
		fmt.Println("Initialized tasks database")
		return nil
	},
}

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Create a new task or epic",
	Long: `Create a new task (or epic with -e flag).

Returns the generated ID (ts-XXXXXX for tasks, ep-XXXXXX for epics).

Examples:
  tasks add "Fix login bug" -p myproject
  tasks add "Auth system" -p myproject -e
  tasks add "Critical fix" --priority 1`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		itemType := model.ItemTypeTask
		if flagEpic {
			itemType = model.ItemTypeEpic
		}

		item := &model.Item{
			ID:        model.GenerateID(itemType),
			Project:   flagProject,
			Type:      itemType,
			Title:     strings.Join(args, " "),
			Status:    model.StatusOpen,
			Priority:  flagPriority,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.CreateItem(item); err != nil {
			return err
		}
		fmt.Println(item.ID)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List all tasks, optionally filtered by project and/or status.

Examples:
  tasks list
  tasks list -p myproject
  tasks list --status open
  tasks list -p myproject --status blocked`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		var status *model.Status
		if flagStatus != "" {
			s := model.Status(flagStatus)
			if !s.IsValid() {
				return fmt.Errorf("invalid status: %s (valid: open, in_progress, blocked, done)", flagStatus)
			}
			status = &s
		}

		items, err := database.ListItems(flagProject, status)
		if err != nil {
			return err
		}

		printItemsTable(items)
		return nil
	},
}

var readyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Show tasks ready for work (unblocked)",
	Long: `Show tasks that are ready to work on.

A task is "ready" when:
  - Status is "open" (not in_progress, blocked, or done)
  - All dependencies are "done"

Results are sorted by priority (1=high first).

Examples:
  tasks ready
  tasks ready -p myproject`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		items, err := database.ReadyItems(flagProject)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			fmt.Println("No ready tasks")
			return nil
		}
		printItemsTable(items)
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Long: `Show full details for a task including description, logs, and dependencies.

Example:
  tasks show ts-a1b2c3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		item, err := database.GetItem(args[0])
		if err != nil {
			return err
		}

		logs, err := database.GetLogs(args[0])
		if err != nil {
			return err
		}

		deps, err := database.GetDeps(args[0])
		if err != nil {
			return err
		}

		printItemDetail(item, logs, deps)
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start working on a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.UpdateStatus(args[0], model.StatusInProgress); err != nil {
			return err
		}
		fmt.Printf("Started %s\n", args[0])
		return nil
	},
}

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark a task as done",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.UpdateStatus(args[0], model.StatusDone); err != nil {
			return err
		}
		fmt.Printf("Completed %s\n", args[0])
		return nil
	},
}

var blockCmd = &cobra.Command{
	Use:   "block <id> <reason>",
	Short: "Mark a task as blocked",
	Long: `Mark a task as blocked and log the reason.

Use this when you can't proceed and need to hand off to another agent.

Example:
  tasks block ts-a1b2c3 "Need API spec from product team"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		id := args[0]
		reason := strings.Join(args[1:], " ")

		if err := database.UpdateStatus(id, model.StatusBlocked); err != nil {
			return err
		}
		if err := database.AddLog(id, "Blocked: "+reason); err != nil {
			return err
		}
		fmt.Printf("Blocked %s: %s\n", id, reason)
		return nil
	},
}

var logCmd = &cobra.Command{
	Use:   "log <id> <message>",
	Short: "Add a log entry to a task",
	Long: `Add a timestamped log entry to a task's audit trail.

Use this to track progress while working.

Example:
  tasks log ts-a1b2c3 "Implemented token refresh logic"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		id := args[0]
		message := strings.Join(args[1:], " ")

		if err := database.AddLog(id, message); err != nil {
			return err
		}
		fmt.Printf("Logged to %s\n", id)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status overview",
	Long: `Show a summary of project status for agent spin-up.

Includes:
  - Count by status (open, in_progress, blocked, done)
  - Recently completed tasks
  - Currently in-progress tasks
  - Blocked tasks with reasons
  - Ready tasks by priority

Examples:
  tasks status
  tasks status -p myproject`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		report, err := database.ProjectStatus(flagProject)
		if err != nil {
			return err
		}

		printStatusReport(report)
		return nil
	},
}

var appendCmd = &cobra.Command{
	Use:   "append <id> <text>",
	Short: "Append text to a task's description",
	Long: `Append text to a task's description.

Use this to add context, decisions, or handoff notes.

Example:
  tasks append ts-a1b2c3 "Decided to use JWT instead of sessions"`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		id := args[0]
		text := strings.Join(args[1:], " ")

		if err := database.AppendDescription(id, text); err != nil {
			return err
		}
		fmt.Printf("Appended to %s\n", id)
		return nil
	},
}

var depCmd = &cobra.Command{
	Use:   "dep <id> --on <other>",
	Short: "Add a dependency",
	Long: `Add a dependency between tasks.

The first task will be blocked until the dependency is done.

Example:
  tasks dep ts-a1b2c3 --on ts-d4e5f6
  # ts-a1b2c3 is now blocked until ts-d4e5f6 is done`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagDependsOn == "" {
			return fmt.Errorf("--on flag is required")
		}

		database, err := openDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.AddDep(args[0], flagDependsOn); err != nil {
			return err
		}
		fmt.Printf("%s now depends on %s\n", args[0], flagDependsOn)
		return nil
	},
}

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output context for Claude Code hooks",
	Long: `Output essential workflow context for AI agents.

Designed to run on SessionStart and PreCompact hooks to ensure
agents maintain context about the tasks workflow.

Example hook configuration in Claude Code settings:
  "hooks": {
    "SessionStart": [{"command": "tasks prime"}],
    "PreCompact": [{"command": "tasks prime"}]
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := openDB()
		if err != nil {
			// Still output prime content even if DB fails
			printPrimeContent(nil)
			return nil
		}
		defer database.Close()

		report, _ := database.ProjectStatus("")
		printPrimeContent(report)
		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "Project scope")

	// add flags
	addCmd.Flags().BoolVarP(&flagEpic, "epic", "e", false, "Create an epic instead of a task")
	addCmd.Flags().IntVar(&flagPriority, "priority", 2, "Priority (1=high, 2=medium, 3=low)")

	// list flags
	listCmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (open, in_progress, blocked, done)")

	// dep flags
	depCmd.Flags().StringVar(&flagDependsOn, "on", "", "ID of the item this depends on")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(readyCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(appendCmd)
	rootCmd.AddCommand(depCmd)
	rootCmd.AddCommand(primeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Output formatting

func printItemsTable(items []model.Item) {
	if len(items) == 0 {
		fmt.Println("No items")
		return
	}

	fmt.Printf("%-12s %-12s %-4s %s\n", "ID", "STATUS", "PRI", "TITLE")
	fmt.Println(strings.Repeat("-", 60))
	for _, item := range items {
		fmt.Printf("%-12s %-12s %-4d %s\n", item.ID, item.Status, item.Priority, item.Title)
	}
}

func printItemDetail(item *model.Item, logs []model.Log, deps []string) {
	fmt.Printf("ID:          %s\n", item.ID)
	fmt.Printf("Type:        %s\n", item.Type)
	fmt.Printf("Project:     %s\n", item.Project)
	fmt.Printf("Title:       %s\n", item.Title)
	fmt.Printf("Status:      %s\n", item.Status)
	fmt.Printf("Priority:    %d\n", item.Priority)
	fmt.Printf("Created:     %s\n", item.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated:     %s\n", item.UpdatedAt.Format(time.RFC3339))

	if item.ParentID != nil {
		fmt.Printf("Parent:      %s\n", *item.ParentID)
	}

	if item.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", item.Description)
	}

	if len(deps) > 0 {
		fmt.Printf("\nDependencies:\n")
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep)
		}
	}

	if len(logs) > 0 {
		fmt.Printf("\nLogs:\n")
		for _, log := range logs {
			fmt.Printf("  [%s] %s\n", log.CreatedAt.Format("2006-01-02 15:04"), log.Message)
		}
	}
}

func printStatusReport(report *db.StatusReport) {
	project := report.Project
	if project == "" {
		project = "(all)"
	}
	fmt.Printf("Project: %s\n\n", project)

	fmt.Printf("Summary: %d open, %d in progress, %d blocked, %d done (%d ready)\n\n",
		report.Open, report.InProgress, report.Blocked, report.Done, report.Ready)

	if len(report.RecentDone) > 0 {
		fmt.Println("Recently completed:")
		for _, item := range report.RecentDone {
			fmt.Printf("  [%s] %s\n", item.ID, item.Title)
		}
		fmt.Println()
	}

	if len(report.InProgItems) > 0 {
		fmt.Println("In progress:")
		for _, item := range report.InProgItems {
			fmt.Printf("  [%s] %s\n", item.ID, item.Title)
		}
		fmt.Println()
	}

	if len(report.BlockedItems) > 0 {
		fmt.Println("Blocked:")
		for _, item := range report.BlockedItems {
			fmt.Printf("  [%s] %s\n", item.ID, item.Title)
		}
		fmt.Println()
	}

	if len(report.ReadyItems) > 0 {
		fmt.Println("Ready for work:")
		for _, item := range report.ReadyItems {
			fmt.Printf("  [%s] %s (pri %d)\n", item.ID, item.Title, item.Priority)
		}
	}
}

func printPrimeContent(report *db.StatusReport) {
	fmt.Println(`# Tasks CLI Context

This project uses 'tasks' for cross-session task management.
Run 'tasks status' to see current state.

## SESSION CLOSE PROTOCOL

Before ending ANY session, you MUST complete ALL of these steps:

1. Log progress on active tasks:
   tasks log <id> "What you accomplished"

2. Update task status:
   - tasks done <id>     # if complete
   - tasks block <id> "reason"  # if blocked

3. Add handoff context for next agent:
   tasks append <id> "Next steps: ..."

NEVER end a session without updating task state.
Work is NOT complete until tasks reflect reality.

## Core Rules

- Use 'tasks' for strategic work tracking (persists across sessions)
- Use TodoWrite for tactical within-session checklists
- Always claim work before starting: tasks start <id>
- Log progress frequently, not just at the end

## Essential Commands

# Finding work
tasks status              # Overview of all projects
tasks ready               # Tasks ready to work on
tasks show <id>           # Full details including logs

# Working
tasks start <id>          # Claim a task
tasks log <id> "message"  # Log progress
tasks done <id>           # Mark complete
tasks block <id> "why"    # Mark blocked

# Creating
tasks add "title" -p project    # New task
tasks add "title" -e            # New epic
tasks dep <id> --on <other>     # Add dependency

## Current State`)

	if report != nil {
		fmt.Printf("\n%d open, %d in progress, %d blocked, %d done\n",
			report.Open, report.InProgress, report.Blocked, report.Done)

		if len(report.InProgItems) > 0 {
			fmt.Println("\nIn progress:")
			for _, item := range report.InProgItems {
				fmt.Printf("  [%s] %s\n", item.ID, item.Title)
			}
		}

		if len(report.BlockedItems) > 0 {
			fmt.Println("\nBlocked:")
			for _, item := range report.BlockedItems {
				fmt.Printf("  [%s] %s\n", item.ID, item.Title)
			}
		}
	} else {
		fmt.Println("\n(No database connection - run 'tasks init' if needed)")
	}
}
