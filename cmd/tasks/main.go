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
	return db.Open(path)
}

var rootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Lightweight task management for agents",
	Long:  `A CLI for managing tasks, epics, and dependencies. Designed for AI agents to track work across sessions.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the tasks database",
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
	Args:  cobra.MinimumNArgs(1),
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
	Args:  cobra.ExactArgs(1),
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
	Args:  cobra.MinimumNArgs(2),
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
	Args:  cobra.MinimumNArgs(2),
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
	Args:  cobra.MinimumNArgs(2),
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
	Args:  cobra.ExactArgs(1),
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
