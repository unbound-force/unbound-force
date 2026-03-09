package backlog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrItemNotFound = errors.New("backlog item not found")
)

// Repository handles CRUD operations for backlog items
type Repository struct {
	dir string
}

// NewRepository creates a new backlog repository
func NewRepository(dir string) *Repository {
	return &Repository{dir: dir}
}

// ensureDir ensures the backlog directory exists
func (r *Repository) ensureDir() error {
	return os.MkdirAll(r.dir, 0755)
}

// getFilePath returns the path for an item
func (r *Repository) getFilePath(id string) string {
	return filepath.Join(r.dir, fmt.Sprintf("%s.md", id))
}

// NextID generates the next available BI-NNN ID
func (r *Repository) NextID() (string, error) {
	items, err := r.List()
	if err != nil {
		return "", err
	}

	maxID := 0
	re := regexp.MustCompile(`^BI-(\d+)$`)
	for _, item := range items {
		matches := re.FindStringSubmatch(item.ID)
		if len(matches) == 2 {
			var idNum int
			_, _ = fmt.Sscanf(matches[1], "%d", &idNum)
			if idNum > maxID {
				maxID = idNum
			}
		}
	}

	return fmt.Sprintf("BI-%03d", maxID+1), nil
}

// Save writes an item to the filesystem
func (r *Repository) Save(item *Item) error {
	if err := r.ensureDir(); err != nil {
		return fmt.Errorf("failed to create backlog directory: %w", err)
	}

	item.ModifiedAt = time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = item.ModifiedAt
	}

	frontmatter, err := yaml.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(frontmatter)
	buf.WriteString("---\n\n")
	buf.WriteString(item.Body)
	if !strings.HasSuffix(item.Body, "\n") {
		buf.WriteString("\n")
	}

	path := r.getFilePath(item.ID)
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// Get reads an item from the filesystem
func (r *Repository) Get(id string) (*Item, error) {
	path := r.getFilePath(id)
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format: missing frontmatter")
	}

	var item Item
	if err := yaml.Unmarshal(parts[1], &item); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	item.Body = strings.TrimSpace(string(parts[2]))
	return &item, nil
}

// List returns all backlog items
func (r *Repository) List() ([]*Item, error) {
	if err := r.ensureDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backlog directory: %w", err)
	}

	var items []*Item
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".md")
		item, err := r.Get(id)
		if err == nil {
			items = append(items, item)
		}
	}

	// Sort by priority then ID
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority < items[j].Priority
		}
		return items[i].ID < items[j].ID
	})

	return items, nil
}
