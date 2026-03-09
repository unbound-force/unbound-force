package backlog_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

func TestRepository_NextID_EmptyBacklog(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	id, err := repo.NextID()
	if err != nil {
		t.Fatalf("NextID failed: %v", err)
	}
	if id != "BI-001" {
		t.Errorf("Expected BI-001, got %s", id)
	}
}

func TestRepository_NextID_WithExistingItems(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	_ = repo.Save(&backlog.Item{ID: "BI-005"})
	_ = repo.Save(&backlog.Item{ID: "BI-042"})

	id, err := repo.NextID()
	if err != nil {
		t.Fatalf("NextID failed: %v", err)
	}
	if id != "BI-043" {
		t.Errorf("Expected BI-043, got %s", id)
	}
}

func TestRepository_Save_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)
	now := time.Now().UTC().Truncate(time.Second) // yaml might truncate sub-second
	num := 123
	item := &backlog.Item{
		ID:                "BI-001",
		Title:             "Test Item",
		Type:              "story",
		Priority:          "P1",
		Status:            "draft",
		Sprint:            "Sprint 1",
		EffortEstimate:    "M",
		Dependencies:      []string{"BI-002"},
		RelatedSpecs:      []string{"spec.md"},
		GitHubIssueNumber: &num,
		CreatedAt:         now,
		Body:              "This is a test.",
	}

	if err := repo.Save(item); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	savedItem, err := repo.Get("BI-001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if savedItem.Title != "Test Item" {
		t.Errorf("Expected title 'Test Item', got '%s'", savedItem.Title)
	}
	if savedItem.Body != "This is a test." {
		t.Errorf("Expected body 'This is a test.', got '%s'", savedItem.Body)
	}
	if savedItem.Sprint != "Sprint 1" {
		t.Errorf("Expected sprint 'Sprint 1', got '%s'", savedItem.Sprint)
	}
	if len(savedItem.Dependencies) != 1 || savedItem.Dependencies[0] != "BI-002" {
		t.Errorf("Expected dependencies ['BI-002'], got %v", savedItem.Dependencies)
	}
	if savedItem.GitHubIssueNumber == nil || *savedItem.GitHubIssueNumber != 123 {
		t.Errorf("Expected github issue 123")
	}
	if !savedItem.CreatedAt.Equal(now) {
		t.Errorf("Expected created at %v, got %v", now, savedItem.CreatedAt)
	}
}

func TestRepository_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	_, err := repo.Get("BI-999")
	if !errors.Is(err, backlog.ErrItemNotFound) {
		t.Errorf("Expected ErrItemNotFound, got %v", err)
	}
}

func TestRepository_Get_MalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	malformed := []byte("---\nbad yaml\n---\nbody")
	os.WriteFile(filepath.Join(dir, "BI-001.md"), malformed, 0644)

	_, err := repo.Get("BI-001")
	if err == nil {
		t.Errorf("Expected error for malformed frontmatter, got nil")
	}
}

func TestRepository_List_SortsByPriorityThenID(t *testing.T) {
	dir := t.TempDir()
	repo := backlog.NewRepository(dir)

	_ = repo.Save(&backlog.Item{ID: "BI-003", Priority: "P1"})
	_ = repo.Save(&backlog.Item{ID: "BI-002", Priority: "P2"})
	_ = repo.Save(&backlog.Item{ID: "BI-001", Priority: "P1"})

	items, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	if items[0].ID != "BI-001" || items[1].ID != "BI-003" || items[2].ID != "BI-002" {
		t.Errorf("Expected BI-001, BI-003, BI-002, got %s, %s, %s", items[0].ID, items[1].ID, items[2].ID)
	}
}
