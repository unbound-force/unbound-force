package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/coaching"
	"github.com/unbound-force/unbound-force/internal/dashboard"
	"github.com/unbound-force/unbound-force/internal/impediment"
	"github.com/unbound-force/unbound-force/internal/metrics"
	"github.com/unbound-force/unbound-force/internal/sprint"
	"github.com/unbound-force/unbound-force/internal/sync"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// MxFParams provides dependency injection for all mxf subcommands.
type MxFParams struct {
	DataDir  string
	Stdout   io.Writer
	Stderr   io.Writer
	GHRunner sync.GHRunner
	Now      func() time.Time
	Format   string
}

func newRootCmd() *cobra.Command {
	return newRootCmdWithParams(&MxFParams{})
}

func newRootCmdWithParams(params *MxFParams) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mxf",
		Short: "Mx F -- Flow Facilitator and Continuous Improvement Coach",
		Long: `Mx F is the Manager hero of the Unbound Force swarm. It provides
metrics collection, querying, impediment tracking, dashboard
visualization, sprint management, and retrospective facilitation.

Use 'mxf collect' to gather metrics from GitHub, Gaze, Divisor,
and Muti-Mind. Use 'mxf metrics' to query and analyze the data.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	}

	rootCmd.PersistentFlags().StringVar(&params.DataDir, "data-dir", ".mx-f", "Root data directory")
	rootCmd.PersistentFlags().StringVar(&params.Format, "format", "text", "Output format (text|json)")

	rootCmd.AddCommand(newCollectCmd(params))
	rootCmd.AddCommand(newMetricsCmd(params))
	rootCmd.AddCommand(newImpedimentCmd(params))
	rootCmd.AddCommand(newDashboardCmd(params))
	rootCmd.AddCommand(newSprintCmd(params))
	rootCmd.AddCommand(newStandupCmd(params))
	rootCmd.AddCommand(newRetroCmd(params))

	return rootCmd
}

// defaults fills in zero-value fields with production defaults.
func (p *MxFParams) defaults() {
	if p.Stdout == nil {
		p.Stdout = os.Stdout
	}
	if p.Stderr == nil {
		p.Stderr = os.Stderr
	}
	if p.Now == nil {
		p.Now = time.Now
	}
	if p.DataDir == "" {
		p.DataDir = ".mx-f"
	}
	if p.GHRunner == nil {
		p.GHRunner = &sync.DefaultGHRunner{}
	}
}

func newCollectCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect metrics from data sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			source, _ := cmd.Flags().GetString("source")
			repo, _ := cmd.Flags().GetString("repo")
			period, _ := cmd.Flags().GetString("period")
			return runCollect(*p, source, repo, period)
		},
	}
	cmd.Flags().String("source", "all", "Data source: github, gaze, divisor, muti-mind, all")
	cmd.Flags().String("repo", "", "GitHub repository (owner/repo)")
	cmd.Flags().String("period", "90d", "Collection period (e.g., 30d, 90d)")
	return cmd
}

func newMetricsCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Query and analyze collected metrics",
	}
	for _, sub := range []struct{ name, desc string }{
		{"summary", "Consolidated health snapshot"},
		{"velocity", "Velocity per sprint with trend"},
		{"cycle-time", "Cycle time statistics"},
		{"bottlenecks", "Pipeline bottleneck analysis"},
		{"health", "Traffic-light health indicators"},
	} {
		subName := sub.name
		subCmd := &cobra.Command{
			Use:   subName,
			Short: sub.desc,
			RunE: func(cmd *cobra.Command, args []string) error {
				p.defaults()
				sprints, _ := cmd.Flags().GetInt("sprints")
				period, _ := cmd.Flags().GetString("period")
				return runMetrics(*p, subName, p.Format, sprints, period)
			},
		}
		subCmd.Flags().Int("sprints", 0, "Number of sprints (0=all)")
		subCmd.Flags().String("period", "30d", "Time period")
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func newImpedimentCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "impediment",
		Short: "Track and manage impediments",
	}

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Log a new impediment",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			title, _ := cmd.Flags().GetString("title")
			severity, _ := cmd.Flags().GetString("severity")
			owner, _ := cmd.Flags().GetString("owner")
			desc, _ := cmd.Flags().GetString("description")
			return runImpedimentAdd(*p, title, severity, owner, desc)
		},
	}
	addCmd.Flags().String("title", "", "Impediment title (required)")
	_ = addCmd.MarkFlagRequired("title")
	addCmd.Flags().String("severity", "medium", "Severity: critical, high, medium, low")
	addCmd.Flags().String("owner", "", "Owner responsible")
	addCmd.Flags().String("description", "", "Detailed description")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List impediments",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			status, _ := cmd.Flags().GetString("status")
			return runImpedimentList(*p, status, p.Format)
		},
	}
	listCmd.Flags().String("status", "open", "Filter: open, resolved, all")

	resolveCmd := &cobra.Command{
		Use:   "resolve [IMP-NNN]",
		Short: "Resolve an impediment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			resolution, _ := cmd.Flags().GetString("resolution")
			return runImpedimentResolve(*p, args[0], resolution)
		},
	}
	resolveCmd.Flags().String("resolution", "", "Resolution description (required)")
	_ = resolveCmd.MarkFlagRequired("resolution")

	detectCmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect impediments from metrics trends",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			return runImpedimentDetect(*p)
		},
	}

	cmd.AddCommand(addCmd, listCmd, resolveCmd, detectCmd)
	return cmd
}

func newDashboardCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Trend visualizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			html, _ := cmd.Flags().GetBool("html")
			output, _ := cmd.Flags().GetString("output")
			return runDashboard(*p, "", html, output)
		},
	}
	cmd.Flags().Bool("html", false, "Generate HTML dashboard")
	cmd.Flags().String("output", "dashboard.html", "HTML output path")

	for _, sub := range []struct{ name, desc string }{
		{"velocity", "Velocity bar chart"},
		{"cycle-time", "Cycle time sparkline"},
		{"health", "Health indicators with sparklines"},
	} {
		subName := sub.name
		subCmd := &cobra.Command{
			Use:   subName,
			Short: sub.desc,
			RunE: func(cmd *cobra.Command, args []string) error {
				p.defaults()
				html, _ := cmd.Flags().GetBool("html")
				output, _ := cmd.Flags().GetString("output")
				return runDashboard(*p, subName, html, output)
			},
		}
		subCmd.Flags().Bool("html", false, "Generate HTML dashboard")
		subCmd.Flags().String("output", "dashboard.html", "HTML output path")
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func newSprintCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sprint",
		Short: "Sprint lifecycle management",
	}
	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Begin sprint planning",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			goal, _ := cmd.Flags().GetString("goal")
			return runSprint(*p, "plan", goal)
		},
	}
	planCmd.Flags().String("goal", "", "Sprint goal")

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "Sprint review summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			return runSprint(*p, "review", "")
		},
	}
	cmd.AddCommand(planCmd, reviewCmd)
	return cmd
}

func newStandupCmd(p *MxFParams) *cobra.Command {
	return &cobra.Command{
		Use:   "standup",
		Short: "Daily standup report",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			return runStandup(*p)
		},
	}
}

func newRetroCmd(p *MxFParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retro",
		Short: "Retrospective facilitation",
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Begin a structured retrospective",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			return runRetro(*p, "start", "")
		},
	}
	actionsCmd := &cobra.Command{
		Use:   "actions",
		Short: "List retrospective action items",
		RunE: func(cmd *cobra.Command, args []string) error {
			p.defaults()
			status, _ := cmd.Flags().GetString("status")
			return runRetro(*p, "actions", status)
		},
	}
	actionsCmd.Flags().String("status", "all", "Filter: pending, completed, stale, all")
	cmd.AddCommand(startCmd, actionsCmd)
	return cmd
}

// --- Implementation functions delegating to domain packages ---

func runCollect(p MxFParams, source, repo, period string) error {
	dur, err := metrics.ParsePeriod(period)
	if err != nil {
		return err
	}
	store := metrics.NewStore(p.DataDir)
	c := &metrics.Collector{
		GHRunner:    p.GHRunner,
		ArtifactDir: filepath.Join(p.DataDir, ".."),
		Store:       store,
		Stdout:      p.Stdout,
		Now:         p.Now,
	}
	sources := []string{source}
	return c.Collect(sources, repo, dur)
}

func runMetrics(p MxFParams, sub, format string, sprints int, period string) error {
	dur, err := metrics.ParsePeriod(period)
	if err != nil {
		return err
	}
	store := metrics.NewStore(p.DataDir)
	q := metrics.NewQuery(store)

	switch sub {
	case "summary":
		snap, err := q.Summary(dur)
		if err != nil {
			return err
		}
		if format == "json" {
			return outputJSON(p.Stdout, snap)
		}
		return outputMetricsSummary(p.Stdout, snap)
	case "velocity":
		points, err := q.Velocity(sprints)
		if err != nil {
			return err
		}
		if format == "json" {
			return outputJSON(p.Stdout, points)
		}
		for _, pt := range points {
			_, _ = fmt.Fprintf(p.Stdout, "  %-12s  %.1f items\n", pt.Sprint, pt.Velocity)
		}
		return nil
	case "cycle-time":
		stats, err := q.CycleTime(dur)
		if err != nil {
			return err
		}
		if format == "json" {
			return outputJSON(p.Stdout, stats)
		}
		_, _ = fmt.Fprintf(p.Stdout, "Cycle Time (last %s)\n", period)
		_, _ = fmt.Fprintf(p.Stdout, "  Avg: %.1fh  Median: %.1fh  P90: %.1fh  P99: %.1fh\n",
			stats.Avg, stats.Median, stats.P90, stats.P99)
		return nil
	case "bottlenecks":
		results, err := q.Bottlenecks()
		if err != nil {
			return err
		}
		if format == "json" {
			return outputJSON(p.Stdout, results)
		}
		fmt.Fprintf(p.Stdout, "Bottleneck Analysis\n")
		for _, r := range results {
			barLen := int(r.AvgWaitDays * 10)
			if barLen < 1 && r.AvgWaitDays > 0 {
				barLen = 1
			}
			bar := strings.Repeat("█", barLen)
			fmt.Fprintf(p.Stdout, "  %-15s %.1fd avg  %s\n", r.Stage, r.AvgWaitDays, bar)
		}
		return nil
	case "health":
		indicators, err := q.Health()
		if err != nil {
			return err
		}
		if format == "json" {
			return outputJSON(p.Stdout, indicators)
		}
		return dashboard.RenderHealthIndicators("Health", indicators, p.Stdout)
	default:
		return fmt.Errorf("unknown metrics subcommand: %s", sub)
	}
}

func runImpedimentAdd(p MxFParams, title, severity, owner, desc string) error {
	repo := impediment.NewRepository(filepath.Join(p.DataDir, "impediments"))
	imp, err := repo.Add(title, severity, owner, desc, p.Now())
	if err != nil {
		return err
	}
	fmt.Fprintf(p.Stdout, "Created impediment %s: %q (severity: %s, owner: %s)\n",
		imp.ID, imp.Title, imp.Severity, imp.Owner)
	return nil
}

func runImpedimentList(p MxFParams, status, format string) error {
	repo := impediment.NewRepository(filepath.Join(p.DataDir, "impediments"))
	imps, err := repo.List(status)
	if err != nil {
		return err
	}
	if format == "json" {
		return outputJSON(p.Stdout, imps)
	}
	if len(imps) == 0 {
		_, _ = fmt.Fprintln(p.Stdout, "No impediments found.")
		return nil
	}
	fmt.Fprintf(p.Stdout, "%-8s  %-8s  %-4s  %-10s  %s\n", "ID", "Severity", "Age", "Owner", "Title")
	for _, imp := range imps {
		stale := ""
		if imp.IsStale() {
			stale = " (stale)"
		}
		fmt.Fprintf(p.Stdout, "%-8s  %-8s  %dd    %-10s  %s%s\n",
			imp.ID, imp.Severity, imp.AgeDays(), imp.Owner, imp.Title, stale)
	}
	return nil
}

func runImpedimentResolve(p MxFParams, id, resolution string) error {
	repo := impediment.NewRepository(filepath.Join(p.DataDir, "impediments"))
	if err := repo.Resolve(id, resolution, p.Now()); err != nil {
		return err
	}
	fmt.Fprintf(p.Stdout, "Resolved %s\nResolution: %s\n", id, resolution)
	return nil
}

func runImpedimentDetect(p MxFParams) error {
	store := metrics.NewStore(p.DataDir)
	repo := impediment.NewRepository(filepath.Join(p.DataDir, "impediments"))
	detected, err := impediment.Detect(store, repo, p.Now())
	if err != nil {
		return err
	}
	if len(detected) == 0 {
		_, _ = fmt.Fprintln(p.Stdout, "No potential impediments detected.")
		return nil
	}
	fmt.Fprintf(p.Stdout, "Detected %d potential impediments:\n\n", len(detected))
	for _, imp := range detected {
		fmt.Fprintf(p.Stdout, "  %s (draft)  severity: %s\n    %s\n\n", imp.ID, imp.Severity, imp.Title)
	}
	return nil
}

func runDashboard(p MxFParams, sub string, html bool, output string) error {
	store := metrics.NewStore(p.DataDir)
	q := metrics.NewQuery(store)

	snap, err := q.Summary(90 * 24 * time.Hour)
	if err != nil {
		return err
	}

	if html {
		indicators := snap.HealthIndicators
		return dashboard.RenderHTML(*snap, indicators, output)
	}

	switch sub {
	case "velocity":
		points, err := q.Velocity(0)
		if err != nil {
			return err
		}
		var bars []dashboard.BarChartPoint
		for _, pt := range points {
			bars = append(bars, dashboard.BarChartPoint{Label: pt.Sprint, Value: pt.Velocity})
		}
		return dashboard.RenderBarChart("Velocity", bars, p.Stdout)
	case "cycle-time":
		snaps, err := store.ReadSnapshots(time.Time{})
		if err != nil {
			return err
		}
		var values []float64
		for _, s := range snaps {
			values = append(values, s.CycleTime.Avg)
		}
		return dashboard.RenderSparkline("Cycle Time (avg hours)", values, p.Stdout)
	case "health":
		return dashboard.RenderHealthIndicators("Health", snap.HealthIndicators, p.Stdout)
	default:
		// Full dashboard
		points, err := q.Velocity(0)
		if err == nil && len(points) > 0 {
			var bars []dashboard.BarChartPoint
			for _, pt := range points {
				bars = append(bars, dashboard.BarChartPoint{Label: pt.Sprint, Value: pt.Velocity})
			}
			_ = dashboard.RenderBarChart("Velocity", bars, p.Stdout)
			_, _ = fmt.Fprintln(p.Stdout)
		}
		return dashboard.RenderHealthIndicators("Health", snap.HealthIndicators, p.Stdout)
	}
}

func runSprint(p MxFParams, sub, goal string) error {
	sprintStore := sprint.NewSprintStore(filepath.Join(p.DataDir, "sprints"))

	switch sub {
	case "plan":
		store := metrics.NewStore(p.DataDir)
		snaps, _ := store.ReadSnapshots(time.Time{})
		avgVel := 10.0
		if len(snaps) > 0 {
			total := 0.0
			for _, s := range snaps {
				total += s.Velocity
			}
			avgVel = total / float64(len(snaps))
		}
		state, err := sprintStore.Plan(goal, avgVel, nil)
		if err != nil {
			return err
		}
		fmt.Fprintf(p.Stdout, "Sprint Planning: %s\n", state.SprintName)
		if goal != "" {
			fmt.Fprintf(p.Stdout, "Goal: %s\n", goal)
		}
		fmt.Fprintf(p.Stdout, "Historical velocity: %.1f items/sprint\n", avgVel)
		fmt.Fprintf(p.Stdout, "Planned items: %d\n", len(state.PlannedItems))
		return nil
	case "review":
		latest, err := sprintStore.Latest()
		if err != nil || latest == nil {
			return fmt.Errorf("no sprint data found. Run `mxf sprint plan` first")
		}
		reviewed, err := sprintStore.Review(latest.SprintName)
		if err != nil {
			return err
		}
		fmt.Fprintf(p.Stdout, "Sprint Review: %s\n", reviewed.SprintName)
		fmt.Fprintf(p.Stdout, "Completed: %d/%d items\n", len(reviewed.CompletedItems), len(reviewed.PlannedItems))
		fmt.Fprintf(p.Stdout, "Velocity: %.1f\n", reviewed.Velocity)
		return nil
	default:
		return fmt.Errorf("unknown sprint subcommand: %s", sub)
	}
}

func runStandup(p MxFParams) error {
	sprintStore := sprint.NewSprintStore(filepath.Join(p.DataDir, "sprints"))
	impRepo := impediment.NewRepository(filepath.Join(p.DataDir, "impediments"))
	store := metrics.NewStore(p.DataDir)
	return sprint.Standup(sprintStore, impRepo, store, p.Stdout)
}

func runRetro(p MxFParams, sub, status string) error {
	retroStore := coaching.NewRetroStore(filepath.Join(p.DataDir, "retros"))

	switch sub {
	case "start":
		date := p.Now().Format("2006-01-02")
		store := metrics.NewStore(p.DataDir)
		snap, _ := store.ReadSnapshots(time.Time{})
		metricsData := make(map[string]interface{})
		if len(snap) > 0 {
			latest := snap[len(snap)-1]
			metricsData["velocity"] = latest.Velocity
			metricsData["ci_pass_rate"] = latest.CIPassRate
			metricsData["review_iterations"] = latest.ReviewIterations
		}

		// Review previous action items
		retros, _ := retroStore.ListRetros()
		prevItems := coaching.ReviewPreviousActions(retros)
		if len(prevItems) > 0 {
			fmt.Fprintf(p.Stdout, "Previous Action Items:\n")
			for _, ai := range prevItems {
				fmt.Fprintf(p.Stdout, "  %s  %-10s  %s (%s)\n", ai.ID, ai.Status, ai.Description, ai.Owner)
			}
			_, _ = fmt.Fprintln(p.Stdout)
		}

		record, err := retroStore.StartRetro(date, metricsData)
		if err != nil {
			return err
		}
		if err := retroStore.SaveRetro(record); err != nil {
			return err
		}
		fmt.Fprintf(p.Stdout, "Retrospective started for %s\n", date)
		fmt.Fprintf(p.Stdout, "Saved to %s\n", filepath.Join(p.DataDir, "retros", date+"-retro.md"))
		return nil
	case "actions":
		retros, err := retroStore.ListRetros()
		if err != nil {
			return err
		}
		if len(retros) == 0 {
			_, _ = fmt.Fprintln(p.Stdout, "No action items found. Run `mxf retro start` first.")
			return nil
		}
		var allItems []coaching.ActionItem
		for _, r := range retros {
			for _, ai := range r.ActionItems {
				if ai.IsStale() {
					ai.Status = "stale"
				}
				if status == "all" || ai.Status == status {
					allItems = append(allItems, ai)
				}
			}
		}
		if len(allItems) == 0 {
			_, _ = fmt.Fprintln(p.Stdout, "No action items matching filter.")
			return nil
		}
		fmt.Fprintf(p.Stdout, "%-6s  %-10s  %-10s  %-12s  %s\n", "ID", "Status", "Owner", "Deadline", "Description")
		for _, ai := range allItems {
			fmt.Fprintf(p.Stdout, "%-6s  %-10s  %-10s  %-12s  %s\n", ai.ID, ai.Status, ai.Owner, ai.Deadline, ai.Description)
		}
		return nil
	default:
		return fmt.Errorf("unknown retro subcommand: %s", sub)
	}
}

// Helper functions

func outputJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func outputMetricsSummary(w io.Writer, snap *metrics.MetricsSnapshot) error {
	fmt.Fprintf(w, "Metrics Summary\n")
	fmt.Fprintf(w, "─────────────────────────────\n")
	fmt.Fprintf(w, "Velocity:         %.1f items/sprint\n", snap.Velocity)
	fmt.Fprintf(w, "Cycle Time:       %.1fh avg / %.1fh median / %.1fh P90\n",
		snap.CycleTime.Avg, snap.CycleTime.Median, snap.CycleTime.P90)
	fmt.Fprintf(w, "Lead Time:        %.1fh\n", snap.LeadTime)
	fmt.Fprintf(w, "Defect Rate:      %.2f defects/item\n", snap.DefectRate)
	fmt.Fprintf(w, "Review Iters:     %.1f avg\n", snap.ReviewIterations)
	fmt.Fprintf(w, "CI Pass Rate:     %.1f%%\n", snap.CIPassRate)
	fmt.Fprintf(w, "Backlog Health:   %d total / %d ready / %d stale\n",
		snap.BacklogHealth.Total, snap.BacklogHealth.Ready, snap.BacklogHealth.Stale)
	fmt.Fprintf(w, "Flow Efficiency:  %.1f%%\n", snap.FlowEfficiency)
	return nil
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
