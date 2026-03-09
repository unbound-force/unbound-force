package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

// Syncer handles bidirectional sync between local backlog and GitHub issues
type Syncer struct {
	repo *backlog.Repository
}

// NewSyncer creates a new Syncer
func NewSyncer(repo *backlog.Repository) *Syncer {
	return &Syncer{repo: repo}
}

func runGH(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh %s failed: %w (stderr: %s)", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// Push local item to GitHub
func (s *Syncer) Push(id string) error {
	var items []*backlog.Item
	var err error

	if id != "" {
		item, err := s.repo.Get(id)
		if err != nil {
			return err
		}
		items = append(items, item)
	} else {
		items, err = s.repo.List()
		if err != nil {
			return err
		}
	}

	for _, item := range items {
		body := item.Body + fmt.Sprintf("\n\n---\n*Created via Muti-Mind ID: %s*", item.ID)

		if item.GitHubIssueNumber == nil {
			// Create
			out, err := runGH("issue", "create", "--title", fmt.Sprintf("[%s] %s", item.ID, item.Title), "--body", body)
			if err != nil {
				return err
			}

			// Parse issue url to get number
			url := strings.TrimSpace(string(out))
			parts := strings.Split(url, "/")
			if len(parts) > 0 {
				var num int
				_, _ = fmt.Sscanf(parts[len(parts)-1], "%d", &num)
				if num > 0 {
					item.GitHubIssueNumber = &num
					if err := s.repo.Save(item); err != nil {
						return err
					}
					fmt.Printf("Created GitHub Issue #%d for %s\n", num, item.ID)
				}
			}
		} else {
			// Update
			_, err := runGH("issue", "edit", fmt.Sprintf("%d", *item.GitHubIssueNumber), "--title", fmt.Sprintf("[%s] %s", item.ID, item.Title), "--body", body)
			if err != nil {
				return err
			}
			fmt.Printf("Updated GitHub Issue #%d for %s\n", *item.GitHubIssueNumber, item.ID)
		}
	}

	return nil
}

// Pull from GitHub
func (s *Syncer) Pull() error {
	// A naive implementation to pull recent issues and try to map them back
	// or create new local backlog items.

	// Example: get recent issues as JSON
	out, err := runGH("issue", "list", "--json", "number,title,body,state,updatedAt")
	if err != nil {
		return err
	}

	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
	}

	if err := json.Unmarshal(out, &issues); err != nil {
		return err
	}

	items, err := s.repo.List()
	if err != nil {
		return err
	}

	// Create mapping
	itemByIssue := make(map[int]*backlog.Item)
	for _, item := range items {
		if item.GitHubIssueNumber != nil {
			itemByIssue[*item.GitHubIssueNumber] = item
		}
	}

	for _, issue := range issues {
		item, exists := itemByIssue[issue.Number]
		if exists {
			// Naive update (assume remote won)
			// In a real implementation we'd check hashes or updated_at
			item.Title = strings.TrimPrefix(issue.Title, fmt.Sprintf("[%s] ", item.ID))

			statusMap := map[string]string{"OPEN": "ready", "CLOSED": "done"}
			if newStatus, ok := statusMap[issue.State]; ok {
				item.Status = newStatus
			}

			if err := s.repo.Save(item); err != nil {
				return err
			}
			fmt.Printf("Pulled updates for %s from Issue #%d\n", item.ID, issue.Number)
		} else {
			fmt.Printf("Skipping unmapped Issue #%d: %s\n", issue.Number, issue.Title)
		}
	}

	return nil
}

// Status returns sync state
func (s *Syncer) Status() error {
	items, err := s.repo.List()
	if err != nil {
		return err
	}

	fmt.Printf("%-10s %-15s %s\n", "ID", "SYNC STATUS", "GITHUB ISSUE")
	fmt.Println("--------------------------------------------------")

	for _, item := range items {
		status := "un-synced"
		issue := "none"
		if item.GitHubIssueNumber != nil {
			status = "synced" // naive
			issue = fmt.Sprintf("#%d", *item.GitHubIssueNumber)
		}
		fmt.Printf("%-10s %-15s %s\n", item.ID, status, issue)
	}

	return nil
}

// Sync is bidirectional wrapper
func (s *Syncer) Sync() error {
	fmt.Println("Pulling updates from GitHub...")
	if err := s.Pull(); err != nil {
		return err
	}
	fmt.Println("Pushing updates to GitHub...")
	if err := s.Push(""); err != nil {
		return err
	}
	return nil
}

// SyncProject wrapper (stub)
func (s *Syncer) SyncProject() error {
	fmt.Println("GitHub Project sync not fully implemented yet.")
	return nil
}
