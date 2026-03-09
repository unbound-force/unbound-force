package sync

import (
	"bytes"
	"testing"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

type StubGHRunner struct {
	Out []byte
	Err error
}

func (m *StubGHRunner) Run(args ...string) ([]byte, error) {
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
