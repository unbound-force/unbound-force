package backlog_test

import (
	"os"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

func TestBacklogRepository(t *testing.T) {
	dir, err := os.MkdirTemp("", "backlog-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	repo := backlog.NewRepository(dir)

	// Test NextID
	id, err := repo.NextID()
	if err != nil {
		t.Fatalf("NextID failed: %v", err)
	}
	if id != "BI-001" {
		t.Errorf("Expected BI-001, got %s", id)
	}

	// Test Save
	item := &backlog.Item{
		ID:        id,
		Title:     "Test Item",
		Type:      "story",
		Priority:  "P1",
		Status:    "draft",
		CreatedAt: time.Now(),
		Body:      "This is a test.",
	}

	if err := repo.Save(item); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test Get
	savedItem, err := repo.Get(id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if savedItem.Title != "Test Item" {
		t.Errorf("Expected title 'Test Item', got '%s'", savedItem.Title)
	}
	if savedItem.Body != "This is a test." {
		t.Errorf("Expected body 'This is a test.', got '%s'", savedItem.Body)
	}

	// Test List
	items, err := repo.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}
}
