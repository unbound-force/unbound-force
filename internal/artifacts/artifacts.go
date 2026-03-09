package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

const Version = "1.0.0"

func writeArtifact(dir, artifactType, id string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	envelope := Envelope{
		Hero:          "muti-mind",
		Version:       Version,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		ArtifactType:  artifactType,
		SchemaVersion: "1.0.0",
		Payload:       payloadBytes,
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

// GenerateBacklogItemArtifact generates a backlog-item JSON artifact
func GenerateBacklogItemArtifact(dir string, item *backlog.Item) error {
	return writeArtifact(dir, "backlog-item", item.ID, item)
}

// GenerateAcceptanceDecision generates an acceptance-decision JSON artifact
func GenerateAcceptanceDecision(dir string, decision *AcceptanceDecision) error {
	return writeArtifact(dir, "acceptance-decision", decision.ItemID, decision)
}
