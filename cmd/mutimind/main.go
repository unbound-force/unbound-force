package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/backlog"
	"github.com/unbound-force/unbound-force/internal/sync"
)

type AppParams struct {
	Stdout       io.Writer
	BacklogDir   string
	ArtifactsDir string
	OutputFormat string
	GHRunner     sync.GHRunner
}

// newSyncerFromParams creates a Syncer, injecting the GHRunner from params
// when provided (for tests), otherwise using the default gh CLI runner.
func newSyncerFromParams(p *AppParams, repo *backlog.Repository, out io.Writer) *sync.Syncer {
	s := sync.NewSyncer(repo, out)
	if p.GHRunner != nil {
		s.SetRunner(p.GHRunner)
	}
	return s
}

func newRootCmd() *cobra.Command {
	return newRootCmdWithParams(&AppParams{})
}

// newRootCmdWithParams builds the root command using a caller-supplied AppParams.
// Fields already set on params are preserved; flag bindings fill in the rest at
// parse time. This allows tests to inject stubs (e.g. GHRunner) without
// spawning a real process.
func newRootCmdWithParams(params *AppParams) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mutimind",
		Short: "Muti-Mind CLI for backlog management",
	}

	rootCmd.PersistentFlags().StringVar(&params.OutputFormat, "format", "text", "Output format (text|json)")
	rootCmd.PersistentFlags().StringVar(&params.BacklogDir, "backlog-dir", ".uf/muti-mind/backlog", "Backlog directory")
	rootCmd.PersistentFlags().StringVar(&params.ArtifactsDir, "artifacts-dir", ".uf/muti-mind/artifacts", "Artifacts directory")

	rootCmd.AddCommand(newInitCmd(params))
	rootCmd.AddCommand(newAddCmd(params))
	rootCmd.AddCommand(newListCmd(params))
	rootCmd.AddCommand(newUpdateCmd(params))
	rootCmd.AddCommand(newShowCmd(params))
	rootCmd.AddCommand(newSyncPushCmd(params))
	rootCmd.AddCommand(newSyncPullCmd(params))
	rootCmd.AddCommand(newSyncStatusCmd(params))
	rootCmd.AddCommand(newSyncCmd(params))
	rootCmd.AddCommand(newSyncProjectCmd(params))
	rootCmd.AddCommand(newGenerateArtifactCmd(params))
	rootCmd.AddCommand(newDecideCmd(params))

	return rootCmd
}

func newInitCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the Muti-Mind environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(p.BacklogDir, 0755); err != nil {
				return fmt.Errorf("failed to create backlog directory: %w", err)
			}
			if err := os.MkdirAll(p.ArtifactsDir, 0755); err != nil {
				return fmt.Errorf("failed to create artifacts directory: %w", err)
			}
			// Write default config if not exists
			configPath := ".uf/muti-mind/config.yaml"
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				_ = os.WriteFile(configPath, []byte("version: 1\n"), 0644)
			}
			cmd.Printf("Muti-Mind initialized in %s\n", p.BacklogDir)
			return nil
		},
	}
}

func newAddCmd(p *AppParams) *cobra.Command {
	cmd := &cobra.Command{
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

			repo := backlog.NewRepository(p.BacklogDir)
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
	cmd.Flags().String("type", "", "Item type (epic|story|task|bug)")
	cmd.Flags().String("title", "", "Item title")
	cmd.Flags().String("priority", "", "Priority (P1-P5)")
	cmd.Flags().String("description", "", "Item description")
	return cmd
}

func newListCmd(p *AppParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backlog items",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			sprint, _ := cmd.Flags().GetString("sprint")

			repo := backlog.NewRepository(p.BacklogDir)
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

			if p.OutputFormat == "json" {
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
	cmd.Flags().String("status", "", "Filter by status")
	cmd.Flags().String("sprint", "", "Filter by sprint")
	return cmd
}

func newUpdateCmd(p *AppParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a backlog item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			repo := backlog.NewRepository(p.BacklogDir)
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
	cmd.Flags().String("priority", "", "New priority")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("sprint", "", "New sprint")
	return cmd
}

func newShowCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "show [id]",
		Short: "Show details of a backlog item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			repo := backlog.NewRepository(p.BacklogDir)
			item, err := repo.Get(id)
			if err != nil {
				return err
			}

			if p.OutputFormat == "json" {
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
}

func newSyncPushCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "sync-push [id]",
		Short: "Push local backlog items to GitHub Issues",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := ""
			if len(args) > 0 {
				id = args[0]
			}
			repo := backlog.NewRepository(p.BacklogDir)
			syncer := newSyncerFromParams(p, repo, cmd.OutOrStdout())
			return syncer.Push(id)
		},
	}
}

func newSyncPullCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "sync-pull",
		Short: "Pull GitHub Issues into the local backlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := backlog.NewRepository(p.BacklogDir)
			syncer := newSyncerFromParams(p, repo, cmd.OutOrStdout())
			return syncer.Pull()
		},
	}
}

func newSyncStatusCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "sync-status",
		Short: "Report on the synchronization state",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := backlog.NewRepository(p.BacklogDir)
			syncer := newSyncerFromParams(p, repo, cmd.OutOrStdout())
			return syncer.Status()
		},
	}
}

func newSyncCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Bidirectional sync including conflict detection",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := backlog.NewRepository(p.BacklogDir)
			syncer := newSyncerFromParams(p, repo, cmd.OutOrStdout())
			return syncer.Sync()
		},
	}
}

func newSyncProjectCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "sync-project",
		Short: "Sync GitHub Projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := backlog.NewRepository(p.BacklogDir)
			syncer := newSyncerFromParams(p, repo, cmd.OutOrStdout())
			return syncer.SyncProject()
		},
	}
}

func newGenerateArtifactCmd(p *AppParams) *cobra.Command {
	return &cobra.Command{
		Use:   "generate-artifact [item_id]",
		Short: "Generate JSON artifact for a backlog item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			repo := backlog.NewRepository(p.BacklogDir)
			item, err := repo.Get(id)
			if err != nil {
				return err
			}
			if err := artifacts.GenerateBacklogItemArtifact(p.ArtifactsDir, item); err != nil {
				return err
			}
			cmd.Printf("Generated backlog-item artifact for %s in %s\n", id, p.ArtifactsDir)
			return nil
		},
	}
}

func newDecideCmd(p *AppParams) *cobra.Command {
	cmd := &cobra.Command{
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

			valid := map[string]bool{"accept": true, "reject": true, "conditional": true}
			if !valid[decision] {
				return fmt.Errorf("invalid decision %q: must be accept, reject, or conditional", decision)
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

			if err := artifacts.GenerateAcceptanceDecision(p.ArtifactsDir, dec); err != nil {
				return err
			}

			cmd.Printf("Generated acceptance-decision artifact for %s\n", itemID)
			return nil
		},
	}
	cmd.Flags().String("item", "", "Backlog item ID")
	cmd.Flags().String("decision", "", "Decision (accept|reject|conditional)")
	cmd.Flags().String("rationale", "", "Rationale for decision")
	cmd.Flags().String("report-ref", "", "Gaze report reference")
	cmd.Flags().StringSlice("met", nil, "Criteria met")
	cmd.Flags().StringSlice("failed", nil, "Criteria failed")
	return cmd
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
