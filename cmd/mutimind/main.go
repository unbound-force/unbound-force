package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/backlog"
	"github.com/unbound-force/unbound-force/internal/sync"
)

var (
	backlogDir   = ".muti-mind/backlog"
	artifactsDir = ".muti-mind/artifacts"
	repo         = backlog.NewRepository(backlogDir)
	syncer       = sync.NewSyncer(repo)
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "mutimind",
	Short: "Muti-Mind CLI for backlog management",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the Muti-Mind environment",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Muti-Mind initialized.")
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new backlog item",
	RunE: func(cmd *cobra.Command, args []string) error {
		itemType, _ := cmd.Flags().GetString("type")
		title, _ := cmd.Flags().GetString("title")
		priority, _ := cmd.Flags().GetString("priority")
		desc, _ := cmd.Flags().GetString("description")

		if title == "" || itemType == "" {
			return fmt.Errorf("--title and --type are required")
		}

		id, err := repo.NextID()
		if err != nil {
			return err
		}

		if priority == "" {
			priority = "P3"
		}

		item := &backlog.Item{
			ID:       id,
			Title:    title,
			Type:     itemType,
			Priority: priority,
			Status:   "draft",
			Body:     desc,
		}

		if err := repo.Save(item); err != nil {
			return err
		}

		cmd.Printf("Created backlog item %s\n", id)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List backlog items",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		sprint, _ := cmd.Flags().GetString("sprint")

		items, err := repo.List()
		if err != nil {
			return err
		}

		var filtered []*backlog.Item
		for _, item := range items {
			if status != "" && item.Status != status {
				continue
			}
			if sprint != "" && item.Sprint != sprint {
				continue
			}
			filtered = append(filtered, item)
		}

		if outputFormat == "json" {
			b, err := json.MarshalIndent(filtered, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(b))
			return nil
		}

		cmd.Printf("%-10s %-10s %-10s %-15s %s\n", "ID", "PRIORITY", "TYPE", "STATUS", "TITLE")
		cmd.Println("--------------------------------------------------------------------------------")
		for _, item := range filtered {
			cmd.Printf("%-10s %-10s %-10s %-15s %s\n", item.ID, item.Priority, item.Type, item.Status, item.Title)
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a backlog item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		item, err := repo.Get(id)
		if err != nil {
			return err
		}

		if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
			item.Priority = priority
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			item.Status = status
		}
		if sprint, _ := cmd.Flags().GetString("sprint"); sprint != "" {
			item.Sprint = sprint
		}

		if err := repo.Save(item); err != nil {
			return err
		}

		cmd.Printf("Updated backlog item %s\n", id)
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show details of a backlog item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		item, err := repo.Get(id)
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			b, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				return err
			}
			cmd.Println(string(b))
			return nil
		}

		cmd.Printf("ID: %s\n", item.ID)
		cmd.Printf("Title: %s\n", item.Title)
		cmd.Printf("Type: %s\n", item.Type)
		cmd.Printf("Priority: %s\n", item.Priority)
		cmd.Printf("Status: %s\n", item.Status)
		if item.Sprint != "" {
			cmd.Printf("Sprint: %s\n", item.Sprint)
		}
		if item.GitHubIssueNumber != nil {
			cmd.Printf("GitHub Issue: #%d\n", *item.GitHubIssueNumber)
		}
		cmd.Printf("Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
		cmd.Println("\nDescription:")
		cmd.Println("----------------------------------------")
		cmd.Println(item.Body)
		return nil
	},
}

var syncPushCmd = &cobra.Command{
	Use:   "sync-push [id]",
	Short: "Push local backlog items to GitHub Issues",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) > 0 {
			id = args[0]
		}
		return syncer.Push(id)
	},
}

var syncPullCmd = &cobra.Command{
	Use:   "sync-pull",
	Short: "Pull GitHub Issues into the local backlog",
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncer.Pull()
	},
}

var syncStatusCmd = &cobra.Command{
	Use:   "sync-status",
	Short: "Report on the synchronization state",
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncer.Status()
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Bidirectional sync including conflict detection",
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncer.Sync()
	},
}

var syncProjectCmd = &cobra.Command{
	Use:   "sync-project",
	Short: "Sync GitHub Projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		return syncer.SyncProject()
	},
}

var generateArtifactCmd = &cobra.Command{
	Use:   "generate-artifact [item_id]",
	Short: "Generate JSON artifact for a backlog item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		item, err := repo.Get(id)
		if err != nil {
			return err
		}
		if err := artifacts.GenerateBacklogItemArtifact(artifactsDir, item); err != nil {
			return err
		}
		fmt.Printf("Generated backlog-item artifact for %s in %s\n", id, artifactsDir)
		return nil
	},
}

var decideCmd = &cobra.Command{
	Use:   "decide",
	Short: "Generate acceptance-decision artifact",
	RunE: func(cmd *cobra.Command, args []string) error {
		itemID, _ := cmd.Flags().GetString("item")
		decision, _ := cmd.Flags().GetString("decision")
		rationale, _ := cmd.Flags().GetString("rationale")
		reportRef, _ := cmd.Flags().GetString("report-ref")
		met, _ := cmd.Flags().GetStringSlice("met")
		failed, _ := cmd.Flags().GetStringSlice("failed")

		if itemID == "" || decision == "" {
			return fmt.Errorf("--item and --decision are required")
		}

		dec := &artifacts.AcceptanceDecision{
			ItemID:         itemID,
			Decision:       decision,
			Rationale:      rationale,
			CriteriaMet:    met,
			CriteriaFailed: failed,
			GazeReportRef:  reportRef,
			DecidedAt:      time.Now().UTC().Format(time.RFC3339),
		}

		if err := artifacts.GenerateAcceptanceDecision(artifactsDir, dec); err != nil {
			return err
		}

		fmt.Printf("Generated acceptance-decision artifact for %s\n", itemID)
		return nil
	},
}

func init() {
	addCmd.Flags().String("type", "", "Item type (epic|story|task|bug)")
	addCmd.Flags().String("title", "", "Item title")
	addCmd.Flags().String("priority", "", "Priority (P1-P5)")
	addCmd.Flags().String("description", "", "Item description")

	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("sprint", "", "Filter by sprint")

	updateCmd.Flags().String("priority", "", "New priority")
	updateCmd.Flags().String("status", "", "New status")
	updateCmd.Flags().String("sprint", "", "New sprint")

	decideCmd.Flags().String("item", "", "Backlog item ID")
	decideCmd.Flags().String("decision", "", "Decision (accept|reject|conditional)")
	decideCmd.Flags().String("rationale", "", "Rationale for decision")
	decideCmd.Flags().String("report-ref", "", "Gaze report reference")
	decideCmd.Flags().StringSlice("met", nil, "Criteria met")
	decideCmd.Flags().StringSlice("failed", nil, "Criteria failed")

	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "text", "Output format (text|json)")

	rootCmd.AddCommand(initCmd, addCmd, listCmd, updateCmd, showCmd, syncPushCmd, syncPullCmd, syncStatusCmd, syncCmd, syncProjectCmd, generateArtifactCmd, decideCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
