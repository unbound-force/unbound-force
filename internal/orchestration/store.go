package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WorkflowStore handles persistence of workflow instances as JSON
// files in a directory. Each workflow is stored at
// {dir}/{workflow_id}.json.
type WorkflowStore struct {
	Dir string
}

// Save persists a WorkflowInstance to disk as JSON.
func (s *WorkflowStore) Save(wf *WorkflowInstance) error {
	if err := os.MkdirAll(s.Dir, 0755); err != nil {
		return fmt.Errorf("create workflow directory %q: %w", s.Dir, err)
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow %q: %w", wf.WorkflowID, err)
	}

	path := filepath.Join(s.Dir, wf.WorkflowID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write workflow %q: %w", path, err)
	}
	return nil
}

// Load reads a WorkflowInstance from disk by workflow ID.
func (s *WorkflowStore) Load(workflowID string) (*WorkflowInstance, error) {
	path := filepath.Join(s.Dir, workflowID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow %q: %w", path, err)
	}

	var wf WorkflowInstance
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow %q: %w", path, err)
	}
	return &wf, nil
}

// List reads all workflow files from the store directory, optionally
// filtering by status. Pass an empty string to return all workflows.
// Results are sorted by started_at descending (most recent first).
func (s *WorkflowStore) List(statusFilter string) ([]WorkflowInstance, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workflow directory %q: %w", s.Dir, err)
	}

	var workflows []WorkflowInstance
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(s.Dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip unreadable files
		}

		var wf WorkflowInstance
		if err := json.Unmarshal(data, &wf); err != nil {
			continue // skip malformed files
		}

		if statusFilter == "" || wf.Status == statusFilter {
			workflows = append(workflows, wf)
		}
	}

	// Sort by started_at descending (most recent first)
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].StartedAt.After(workflows[j].StartedAt)
	})

	return workflows, nil
}

// Latest returns the most recent in-progress workflow for a given branch.
// An in-progress workflow is one with StatusActive or StatusAwaitingHuman.
// Returns nil if no in-progress workflow exists for the branch.
// Per research.md R3: awaiting_human workflows are still "in progress"
// from the operator's perspective and must be discoverable.
func (s *WorkflowStore) Latest(branch string) (*WorkflowInstance, error) {
	// List all workflows (unfiltered) and find the first matching branch
	// with an in-progress status. Results are sorted by started_at descending.
	workflows, err := s.List("")
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}

	for _, wf := range workflows {
		if wf.FeatureBranch == branch && (wf.Status == StatusActive || wf.Status == StatusAwaitingHuman) {
			result := wf // copy to avoid returning pointer to loop variable
			return &result, nil
		}
	}
	return nil, nil
}
