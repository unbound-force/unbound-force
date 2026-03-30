package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/unbound-force/unbound-force/internal/backlog"
)

// GHRunner is an interface for running GitHub CLI commands
type GHRunner interface {
	Run(args ...string) ([]byte, error)
}

// DefaultGHRunner uses the real gh cli
type DefaultGHRunner struct{}

func (d *DefaultGHRunner) Run(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh %s failed: %w (stderr: %s)", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// Syncer handles bidirectional sync between local backlog and GitHub issues
type Syncer struct {
	repo   *backlog.Repository
	runner GHRunner
	out    io.Writer
}

// NewSyncer creates a new Syncer
func NewSyncer(repo *backlog.Repository, out io.Writer) *Syncer {
	return &Syncer{
		repo:   repo,
		runner: &DefaultGHRunner{},
		out:    out,
	}
}

// SetRunner replaces the GHRunner used by this Syncer.
// Intended for use in tests to inject a stub runner.
func (s *Syncer) SetRunner(r GHRunner) {
	s.runner = r
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
			out, err := s.runner.Run("issue", "create", "--title", fmt.Sprintf("[%s] %s", item.ID, item.Title), "--body", body)
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
					log.Info("Created GitHub Issue", "issue", num, "id", item.ID)
					_, _ = fmt.Fprintf(s.out, "Created GitHub Issue #%d for %s\n", num, item.ID)
				}
			}
		} else {
			// Update
			_, err := s.runner.Run("issue", "edit", fmt.Sprintf("%d", *item.GitHubIssueNumber), "--title", fmt.Sprintf("[%s] %s", item.ID, item.Title), "--body", body)
			if err != nil {
				return err
			}
			log.Info("Updated GitHub Issue", "issue", *item.GitHubIssueNumber, "id", item.ID)
			_, _ = fmt.Fprintf(s.out, "Updated GitHub Issue #%d for %s\n", *item.GitHubIssueNumber, item.ID)
		}
	}

	return nil
}

// Pull from GitHub
func (s *Syncer) Pull() error {
	out, err := s.runner.Run("issue", "list", "--json", "number,title,body,state,updatedAt")
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
			// Simple naive implementation
			item.Title = strings.TrimPrefix(issue.Title, fmt.Sprintf("[%s] ", item.ID))

			statusMap := map[string]string{"OPEN": "ready", "CLOSED": "done"}
			if newStatus, ok := statusMap[issue.State]; ok {
				item.Status = newStatus
			}

			if err := s.repo.Save(item); err != nil {
				return err
			}
			log.Info("Pulled updates", "issue", issue.Number, "id", item.ID)
			_, _ = fmt.Fprintf(s.out, "Pulled updates for %s from Issue #%d\n", item.ID, issue.Number)
		} else {
			// Create new local item
			id, err := s.repo.NextID()
			if err != nil {
				return err
			}
			num := issue.Number
			newItem := &backlog.Item{
				ID:                id,
				Title:             issue.Title,
				Type:              "story", // default
				Priority:          "P3",    // default
				Status:            "ready",
				Body:              issue.Body,
				GitHubIssueNumber: &num,
			}
			if err := s.repo.Save(newItem); err != nil {
				return err
			}
			log.Info("Imported unmapped issue", "issue", num, "id", id)
			_, _ = fmt.Fprintf(s.out, "Imported unmapped Issue #%d as %s\n", num, id)
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

	_, _ = fmt.Fprintf(s.out, "%-10s %-15s %s\n", "ID", "SYNC STATUS", "GITHUB ISSUE")
	_, _ = fmt.Fprintln(s.out, "--------------------------------------------------")

	for _, item := range items {
		status := "un-synced"
		issue := "none"
		if item.GitHubIssueNumber != nil {
			status = "synced" // naive
			issue = fmt.Sprintf("#%d", *item.GitHubIssueNumber)
		}
		_, _ = fmt.Fprintf(s.out, "%-10s %-15s %s\n", item.ID, status, issue)
	}

	return nil
}

// Sync is bidirectional wrapper
func (s *Syncer) Sync() error {
	_, _ = fmt.Fprintln(s.out, "Pulling updates from GitHub...")
	if err := s.Pull(); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(s.out, "Pushing updates to GitHub...")
	if err := s.Push(""); err != nil {
		return err
	}
	return nil
}

// SyncProject wrapper (stub)
func (s *Syncer) SyncProject() error {
	_, _ = fmt.Fprintln(s.out, "GitHub Project sync not fully implemented yet.")
	return nil
}
