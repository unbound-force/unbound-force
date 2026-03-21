package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/unbound-force/unbound-force/internal/backlog"
)

// Envelope represents the Spec 002 Artifact Envelope
type Envelope struct {
	Hero          string          `json:"hero"`
	Version       string          `json:"version"`
	Timestamp     string          `json:"timestamp"`
	ArtifactType  string          `json:"artifact_type"`
	SchemaVersion string          `json:"schema_version"`
	Context       json.RawMessage `json:"context,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

// AcceptanceDecision represents the payload for an acceptance-decision artifact
type AcceptanceDecision struct {
	ItemID         string   `json:"item_id"`
	Decision       string   `json:"decision"` // accept, reject, conditional
	Rationale      string   `json:"rationale"`
	CriteriaMet    []string `json:"criteria_met"`
	CriteriaFailed []string `json:"criteria_failed"`
	GazeReportRef  string   `json:"gaze_report_ref"`
	DecidedAt      string   `json:"decided_at"`
}

// ArtifactContext provides workflow-level metadata for artifact envelopes.
// Fields are optional (omitempty) to support incremental adoption.
type ArtifactContext struct {
	Branch        string `json:"branch,omitempty"`
	Commit        string `json:"commit,omitempty"`
	BacklogItemID string `json:"backlog_item_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	WorkflowID    string `json:"workflow_id,omitempty"`
}

const Version = "1.0.0"

// WriteArtifact writes a JSON artifact envelope to the given directory.
// The hero parameter identifies the producing hero (e.g., "mx-f", "muti-mind").
func WriteArtifact(dir, hero, artifactType, id string, payload interface{}) error {
	return WriteArtifactWithContext(dir, hero, artifactType, id, payload, nil)
}

// WriteArtifactWithContext writes a JSON artifact envelope with optional workflow context.
// If ctx is non-nil, the context fields are included in the envelope for workflow tracking.
func WriteArtifactWithContext(dir, hero, artifactType, id string, payload interface{}, ctx *ArtifactContext) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	envelope := Envelope{
		Hero:          hero,
		Version:       Version,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		ArtifactType:  artifactType,
		SchemaVersion: "1.0.0",
		Payload:       payloadBytes,
	}

	if ctx != nil {
		ctxBytes, err := json.Marshal(ctx)
		if err != nil {
			return fmt.Errorf("marshal context: %w", err)
		}
		envelope.Context = ctxBytes
	}

	envelopeBytes, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create artifacts directory %q: %w", dir, err)
	}

	filename := fmt.Sprintf("%s-%s.json", id, artifactType)
	targetPath := filepath.Join(dir, filename)
	if err := os.WriteFile(targetPath, envelopeBytes, 0644); err != nil {
		return fmt.Errorf("write artifact %q: %w", targetPath, err)
	}
	return nil
}

// ReadEnvelope reads a JSON artifact file and returns the parsed Envelope.
func ReadEnvelope(path string) (*Envelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read artifact %q: %w", path, err)
	}
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse artifact %q: %w", path, err)
	}
	return &env, nil
}

// FindArtifacts discovers artifact files of the given type in a directory tree.
// Returns file paths sorted by filename descending (newest timestamp first).
func FindArtifacts(dir, artifactType string) ([]string, error) {
	var matches []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		env, readErr := ReadEnvelope(path)
		if readErr != nil {
			return nil // skip non-artifact JSON files
		}
		if env.ArtifactType == artifactType {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %q: %w", dir, err)
	}
	// Sort descending by filename (timestamps sort naturally)
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))
	return matches, nil
}

// FindArtifactsByHero discovers artifact files of the given type produced
// by a specific hero. Returns file paths sorted by filename descending.
func FindArtifactsByHero(dir, artifactType, hero string) ([]string, error) {
	all, err := FindArtifacts(dir, artifactType)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range all {
		env, err := ReadEnvelope(path)
		if err != nil {
			continue
		}
		if env.Hero == hero {
			matches = append(matches, path)
		}
	}
	return matches, nil
}

// FindArtifactsSince discovers artifact files of the given type created
// after the specified time. Returns file paths sorted by filename descending.
func FindArtifactsSince(dir, artifactType string, since time.Time) ([]string, error) {
	all, err := FindArtifacts(dir, artifactType)
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, path := range all {
		env, err := ReadEnvelope(path)
		if err != nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339, env.Timestamp)
		if err != nil {
			continue
		}
		if ts.After(since) || ts.Equal(since) {
			matches = append(matches, path)
		}
	}
	return matches, nil
}

// CheckSchemaVersion compares an envelope's schema version with the expected
// version. Returns compatible=true if the major version matches. Returns a
// warning string if minor or patch versions differ.
func CheckSchemaVersion(envelope *Envelope, expectedVersion string) (compatible bool, warning string) {
	if envelope.SchemaVersion == expectedVersion {
		return true, ""
	}

	envParts := splitVersion(envelope.SchemaVersion)
	expParts := splitVersion(expectedVersion)

	if len(envParts) < 1 || len(expParts) < 1 {
		return false, fmt.Sprintf("invalid version format: envelope=%q expected=%q", envelope.SchemaVersion, expectedVersion)
	}

	if envParts[0] != expParts[0] {
		return false, fmt.Sprintf("major version mismatch: envelope=%q expected=%q", envelope.SchemaVersion, expectedVersion)
	}

	return true, fmt.Sprintf("minor/patch version differs: envelope=%q expected=%q", envelope.SchemaVersion, expectedVersion)
}

// splitVersion splits a semver string into its components.
func splitVersion(v string) []string {
	parts := make([]string, 0, 3)
	for _, p := range splitOnDot(v) {
		parts = append(parts, p)
	}
	return parts
}

// splitOnDot splits a string on '.' characters.
func splitOnDot(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return parts
}

// GenerateBacklogItemArtifact generates a backlog-item JSON artifact
func GenerateBacklogItemArtifact(dir string, item *backlog.Item) error {
	return WriteArtifact(dir, "muti-mind", "backlog-item", item.ID, item)
}

// GenerateAcceptanceDecision generates an acceptance-decision JSON artifact
func GenerateAcceptanceDecision(dir string, decision *AcceptanceDecision) error {
	return WriteArtifact(dir, "muti-mind", "acceptance-decision", decision.ItemID, decision)
}
