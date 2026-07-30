package sync

import (
	"bytes"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

type StubGHRunner struct {
	Out   []byte
	Err   error
	Calls int
}

func (m *StubGHRunner) Run(args ...string) ([]byte, error) {
	m.Calls++
	return m.Out, m.Err
}

func TestSyncer_Push_CreatesIssue(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Test Item", Body: "Test Body"})

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)
	syncer.runner = &StubGHRunner{Out: []byte("https://github.com/repo/issues/42\n")}

	err := syncer.Push("BI-001")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	item, _ := repo.Get("BI-001")
	if item.GitHubIssueNumber == nil || *item.GitHubIssueNumber != 42 {
		t.Errorf("Expected issue number 42, got %v", item.GitHubIssueNumber)
	}
}

func TestSyncer_Push_UpdatesExistingIssue(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	num := 42
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Test Item", GitHubIssueNumber: &num})

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)
	syncer.runner = &StubGHRunner{Out: []byte("https://github.com/repo/issues/42\n")}

	err := syncer.Push("BI-001")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}
}

func TestSyncer_Sync_CallsPullThenPush(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Test Item"})

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)
	syncer.runner = &StubGHRunner{Out: []byte(`[]`)} // empty pull, empty push response doesn't matter much for stub here since pull overrides

	err := syncer.Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("Pulling updates from GitHub...")) {
		t.Errorf("Expected 'Pulling updates from GitHub...' in output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Pushing updates to GitHub...")) {
		t.Errorf("Expected 'Pushing updates to GitHub...' in output")
	}
}

func TestSyncer_SyncProject_ReturnsNil(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)

	err := syncer.SyncProject()
	if err != nil {
		t.Fatalf("SyncProject failed: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("GitHub Project sync not fully implemented yet.")) {
		t.Errorf("Expected 'not fully implemented' in output")
	}
}

func TestSyncer_Pull_MapsKnownIssues(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	num := 42
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Test Item", GitHubIssueNumber: &num})

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)
	syncer.runner = &StubGHRunner{
		Out: []byte(`[{"number":42,"title":"[BI-001] Updated Title","body":"Updated Body","state":"CLOSED","updatedAt":"2023-01-01T00:00:00Z"}]`),
	}

	err := syncer.Pull()
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	item, _ := repo.Get("BI-001")
	if item.Title != "Updated Title" {
		t.Errorf("Expected 'Updated Title', got %s", item.Title)
	}
	if item.Status != "done" {
		t.Errorf("Expected 'done' status, got %s", item.Status)
	}
}

func TestSyncer_Pull_CreatesUnmappedIssues(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)
	syncer.runner = &StubGHRunner{
		Out: []byte(`[{"number":99,"title":"New Bug","body":"Something broke","state":"OPEN","updatedAt":"2023-01-01T00:00:00Z"}]`),
	}

	err := syncer.Pull()
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	items, _ := repo.List()
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Title != "New Bug" {
		t.Errorf("Expected 'New Bug', got %s", items[0].Title)
	}
	if items[0].GitHubIssueNumber == nil || *items[0].GitHubIssueNumber != 99 {
		t.Errorf("Expected issue #99 mapped")
	}
}

func TestSyncer_SetRunner(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)

	// Verify default runner is a DefaultGHRunner.
	if syncer.runner == nil {
		t.Fatal("expected non-nil default runner")
	}

	// Inject a stub runner and verify it is used.
	stub := &StubGHRunner{Out: []byte(`[]`)}
	syncer.SetRunner(stub)

	// Pull uses the runner; verify no error with stub.
	err := syncer.Pull()
	if err != nil {
		t.Fatalf("Pull after SetRunner: %v", err)
	}
}

func TestSyncer_Status(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	num := 42
	_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Test Item", GitHubIssueNumber: &num})

	buf := new(bytes.Buffer)
	syncer := NewSyncer(repo, buf)

	err := syncer.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("synced")) {
		t.Errorf("Expected 'synced' in output")
	}
}

func TestSyncer_Push_DryRun(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(repo *backlog.Repository)
		pushID    string
		wantErr   bool
		wantCalls int
		check     func(t *testing.T, output string)
	}{
		{
			name: "items to create",
			setup: func(repo *backlog.Repository) {
				_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "New Item"})
			},
			pushID:    "",
			wantCalls: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "CREATE") {
					t.Errorf("expected CREATE in output, got: %s", output)
				}
				if !strings.Contains(output, "BI-001") {
					t.Errorf("expected BI-001 in output, got: %s", output)
				}
				if !strings.Contains(output, "1 to create, 0 to update") {
					t.Errorf("expected summary '1 to create, 0 to update', got: %s", output)
				}
			},
		},
		{
			name: "items to update",
			setup: func(repo *backlog.Repository) {
				num := 42
				_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Existing Item", GitHubIssueNumber: &num})
			},
			pushID:    "",
			wantCalls: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "UPDATE") {
					t.Errorf("expected UPDATE in output, got: %s", output)
				}
				if !strings.Contains(output, "Issue #42") {
					t.Errorf("expected 'Issue #42' in output, got: %s", output)
				}
				if !strings.Contains(output, "0 to create, 1 to update") {
					t.Errorf("expected summary '0 to create, 1 to update', got: %s", output)
				}
			},
		},
		{
			name: "mixed create and update",
			setup: func(repo *backlog.Repository) {
				_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "New Item"})
				num := 42
				_ = repo.Save(&backlog.Item{ID: "BI-002", Title: "Existing Item", GitHubIssueNumber: &num})
			},
			pushID:    "",
			wantCalls: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "CREATE") {
					t.Errorf("expected CREATE in output, got: %s", output)
				}
				if !strings.Contains(output, "UPDATE") {
					t.Errorf("expected UPDATE in output, got: %s", output)
				}
				if !strings.Contains(output, "1 to create, 1 to update") {
					t.Errorf("expected summary '1 to create, 1 to update', got: %s", output)
				}
			},
		},
		{
			name: "single item filter",
			setup: func(repo *backlog.Repository) {
				_ = repo.Save(&backlog.Item{ID: "BI-001", Title: "Target Item"})
				_ = repo.Save(&backlog.Item{ID: "BI-002", Title: "Other Item"})
			},
			pushID:    "BI-001",
			wantCalls: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "BI-001") {
					t.Errorf("expected BI-001 in output, got: %s", output)
				}
				if strings.Contains(output, "BI-002") {
					t.Errorf("expected BI-002 NOT in output, got: %s", output)
				}
			},
		},
		{
			name:      "empty backlog",
			setup:     nil,
			pushID:    "",
			wantCalls: 0,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "No items pending sync") {
					t.Errorf("expected 'No items pending sync', got: %s", output)
				}
			},
		},
		{
			name:      "non-existent item ID",
			setup:     nil,
			pushID:    "BI-999",
			wantErr:   true,
			wantCalls: 0,
			check:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			repo := backlog.NewRepository(dir)
			if tt.setup != nil {
				tt.setup(repo)
			}

			buf := new(bytes.Buffer)
			stub := &StubGHRunner{}
			syncer := NewSyncer(repo, buf)
			syncer.SetRunner(stub)
			syncer.DryRun = true

			err := syncer.Push(tt.pushID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if stub.Calls != tt.wantCalls {
					t.Errorf("expected %d GHRunner calls, got %d", tt.wantCalls, stub.Calls)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stub.Calls != tt.wantCalls {
				t.Errorf("expected %d GHRunner calls, got %d", tt.wantCalls, stub.Calls)
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}
