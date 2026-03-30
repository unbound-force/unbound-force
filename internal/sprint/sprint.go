package sprint

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/unbound-force/unbound-force/internal/impediment"
	"github.com/unbound-force/unbound-force/internal/metrics"
)

// SprintStore manages sprint state on the filesystem.
type SprintStore struct {
	Dir string
}

// NewSprintStore creates a new sprint store.
func NewSprintStore(dir string) *SprintStore {
	return &SprintStore{Dir: dir}
}

// Plan creates a new sprint with capacity calculation.
func (s *SprintStore) Plan(goal string, avgVelocity float64, backlogItems []string) (*SprintState, error) {
	if err := os.MkdirAll(s.Dir, 0755); err != nil {
		return nil, fmt.Errorf("create sprints dir: %w", err)
	}

	now := time.Now()
	name := fmt.Sprintf("sprint-%s", now.Format("2006-01-02"))
	end := now.Add(14 * 24 * time.Hour) // 2-week sprint

	capacity := int(avgVelocity)
	if capacity < 1 {
		capacity = len(backlogItems)
	}
	if capacity > len(backlogItems) {
		capacity = len(backlogItems)
	}

	planned := backlogItems
	if len(planned) > capacity {
		planned = planned[:capacity]
	}

	state := &SprintState{
		SprintName:   name,
		Goal:         goal,
		StartDate:    now.Format("2006-01-02"),
		EndDate:      end.Format("2006-01-02"),
		PlannedItems: planned,
		Status:       "active",
	}

	return state, s.Save(state)
}

// Review summarizes a completed sprint.
func (s *SprintStore) Review(sprintName string) (*SprintState, error) {
	state, err := s.Load(sprintName)
	if err != nil {
		return nil, err
	}

	state.ComputeVelocity()
	state.Status = "complete"

	return state, s.Save(state)
}

// Save writes sprint state to disk.
func (s *SprintStore) Save(state *SprintState) error {
	if err := os.MkdirAll(s.Dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sprint: %w", err)
	}

	path := filepath.Join(s.Dir, state.SprintName+".json")
	return os.WriteFile(path, data, 0644)
}

// Load reads sprint state from disk.
func (s *SprintStore) Load(name string) (*SprintState, error) {
	path := filepath.Join(s.Dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read sprint %q: %w", path, err)
	}

	var state SprintState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse sprint: %w", err)
	}
	return &state, nil
}

// Latest returns the most recent sprint, or nil if none exist.
func (s *SprintStore) Latest() (*SprintState, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var latest string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			name := e.Name()[:len(e.Name())-5]
			if name > latest {
				latest = name
			}
		}
	}

	if latest == "" {
		return nil, nil
	}
	return s.Load(latest)
}

// Standup produces a daily standup report.
func Standup(store *SprintStore, impRepo *impediment.Repository, metricsStore *metrics.Store, w io.Writer) error {
	now := time.Now()
	fmt.Fprintf(w, "Daily Standup — %s\n", now.Format("2006-01-02"))
	fmt.Fprintf(w, "──────────────────────────\n\n")

	// Current sprint info
	sprint, _ := store.Latest()
	if sprint != nil {
		fmt.Fprintf(w, "Sprint: %s\n", sprint.SprintName)
		if sprint.Goal != "" {
			fmt.Fprintf(w, "Goal: %s\n", sprint.Goal)
		}
		fmt.Fprintf(w, "Planned: %d items  Completed: %d items\n\n",
			len(sprint.PlannedItems), len(sprint.CompletedItems))
	} else {
		fmt.Fprintf(w, "No active sprint.\n\n")
	}

	// Blocked items from impediment tracker
	imps, _ := impRepo.List("open")
	if len(imps) > 0 {
		fmt.Fprintf(w, "Blocked (%d):\n", len(imps))
		for _, imp := range imps {
			stale := ""
			if imp.IsStale() {
				stale = " (stale)"
			}
			fmt.Fprintf(w, "  %s  %s  %s%s\n", imp.ID, imp.Severity, imp.Title, stale)
		}
		_, _ = fmt.Fprintln(w)
	}

	return nil
}
