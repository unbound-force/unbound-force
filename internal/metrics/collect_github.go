package metrics

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/unbound-force/unbound-force/internal/sync"
)

var validRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

// CollectGitHub collects metrics from the GitHub API via the gh CLI.
func CollectGitHub(runner sync.GHRunner, repo string, period time.Duration) (*SourceCollection, error) {
	if !validRepoPattern.MatchString(repo) {
		return nil, fmt.Errorf("invalid repository format %q: expected owner/repo", repo)
	}

	now := time.Now().UTC()
	since := now.Add(-period)
	sinceStr := since.Format("2006-01-02")

	raw := make(map[string]interface{})

	// Collect PR data
	prData, err := runner.Run("api", fmt.Sprintf("repos/%s/pulls", repo),
		"--json", "number,title,state,createdAt,mergedAt,closedAt",
		"--jq", fmt.Sprintf("[.[] | select(.createdAt >= \"%s\")]", sinceStr),
		"--paginate")
	if err != nil {
		return nil, fmt.Errorf("collect PRs: %w", err)
	}

	var prs []map[string]interface{}
	if err := json.Unmarshal(prData, &prs); err != nil {
		// Try parsing as empty result
		prs = nil
	}
	raw["prs"] = prs
	raw["pr_count"] = len(prs)

	// Compute merge times
	var totalMergeHours float64
	mergedCount := 0
	for _, pr := range prs {
		createdStr, _ := pr["createdAt"].(string)
		mergedStr, _ := pr["mergedAt"].(string)
		if createdStr != "" && mergedStr != "" {
			created, err1 := time.Parse(time.RFC3339, createdStr)
			merged, err2 := time.Parse(time.RFC3339, mergedStr)
			if err1 == nil && err2 == nil {
				totalMergeHours += merged.Sub(created).Hours()
				mergedCount++
			}
		}
	}
	if mergedCount > 0 {
		raw["avg_merge_hours"] = totalMergeHours / float64(mergedCount)
	}

	// Collect CI runs
	runData, err := runner.Run("run", "list",
		"--repo", repo,
		"--json", "status,conclusion,createdAt",
		"--limit", "100")
	if err == nil {
		var runs []map[string]interface{}
		if json.Unmarshal(runData, &runs) == nil {
			passed := 0
			for _, r := range runs {
				if conc, ok := r["conclusion"].(string); ok && conc == "success" {
					passed++
				}
			}
			if len(runs) > 0 {
				raw["ci_pass_rate"] = float64(passed) / float64(len(runs)) * 100
			}
			raw["ci_runs"] = len(runs)
		}
	}

	// Collect issues
	issueData, err := runner.Run("api", fmt.Sprintf("repos/%s/issues", repo),
		"--json", "number,state,createdAt,closedAt",
		"--jq", fmt.Sprintf("[.[] | select(.createdAt >= \"%s\")]", sinceStr),
		"--paginate")
	if err == nil {
		var issues []map[string]interface{}
		if json.Unmarshal(issueData, &issues) == nil {
			raw["issues"] = issues
			raw["issue_count"] = len(issues)
			opened := 0
			closed := 0
			for _, i := range issues {
				if state, ok := i["state"].(string); ok {
					if state == "open" || state == "OPEN" {
						opened++
					} else {
						closed++
					}
				}
			}
			raw["issues_opened"] = opened
			raw["issues_closed"] = closed
		}
	}

	dataPoints := len(prs)
	if v, ok := raw["ci_runs"]; ok {
		if n, isInt := v.(int); isInt {
			dataPoints += n
		}
	}
	if v, ok := raw["issue_count"]; ok {
		if n, isInt := v.(int); isInt {
			dataPoints += n
		}
	}

	return &SourceCollection{
		Source:      "github",
		CollectedAt: now,
		DataPoints:  dataPoints,
		RawData:     raw,
	}, nil
}
