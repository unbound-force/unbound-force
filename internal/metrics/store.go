package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Store handles reading and writing metrics data to the filesystem.
type Store struct {
	DataDir string // Root data directory (e.g., ".uf/mx-f/data")
}

// NewStore creates a new metrics store.
func NewStore(dataDir string) *Store {
	return &Store{DataDir: filepath.Join(dataDir, "data")}
}

// WriteCollection saves a source collection to disk.
func (s *Store) WriteCollection(source string, coll SourceCollection) error {
	dir := filepath.Join(s.DataDir, source)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data dir %q: %w", dir, err)
	}

	ts := coll.CollectedAt.UTC().Format(time.RFC3339)
	// Replace colons in timestamp for filesystem safety
	ts = strings.ReplaceAll(ts, ":", "-")
	filename := fmt.Sprintf("%s.json", ts)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(coll, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal collection: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// ReadCollections reads all collections for a source, optionally filtered by time.
func (s *Store) ReadCollections(source string, since time.Time) ([]SourceCollection, error) {
	dir := filepath.Join(s.DataDir, source)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %q: %w", dir, err)
	}

	var results []SourceCollection
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var coll SourceCollection
		if err := json.Unmarshal(data, &coll); err != nil {
			continue
		}
		if !since.IsZero() && coll.CollectedAt.Before(since) {
			continue
		}
		results = append(results, coll)
	}

	// Sort by collection time ascending
	sort.Slice(results, func(i, j int) bool {
		return results[i].CollectedAt.Before(results[j].CollectedAt)
	})
	return results, nil
}

// WriteSnapshot saves a metrics snapshot to disk.
func (s *Store) WriteSnapshot(snap MetricsSnapshot) error {
	dir := filepath.Join(s.DataDir, "snapshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create snapshots dir: %w", err)
	}

	ts := snap.Timestamp.UTC().Format(time.RFC3339)
	ts = strings.ReplaceAll(ts, ":", "-")
	filename := fmt.Sprintf("%s.json", ts)
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ReadSnapshots reads all snapshots, optionally filtered by time.
func (s *Store) ReadSnapshots(since time.Time) ([]MetricsSnapshot, error) {
	dir := filepath.Join(s.DataDir, "snapshots")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read snapshots dir: %w", err)
	}

	var results []MetricsSnapshot
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var snap MetricsSnapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			continue
		}
		if !since.IsZero() && snap.Timestamp.Before(since) {
			continue
		}
		results = append(results, snap)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})
	return results, nil
}
