package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/unbound-force/unbound-force/internal/artifacts"
)

func TestCollectDivisor_ReviewVerdictMigrationDecisions(t *testing.T) {
	tests := []struct {
		name               string
		schemaVersion      string
		decision           artifacts.ReviewVerdictDecision
		progressionAllowed bool
		requiredAction     string
	}{
		{
			name:               "v2 approval permits progression",
			schemaVersion:      artifacts.ReviewVerdictSchemaVersionV2,
			decision:           artifacts.ReviewVerdictApproved,
			progressionAllowed: true,
			requiredAction:     reviewActionNone,
		},
		{
			name:           "v2 advisory requires human resolution",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV2,
			decision:       artifacts.ReviewVerdictEscalated,
			requiredAction: reviewActionHumanResolution,
		},
		{
			name:           "v2 blocking decision requires findings addressed",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV2,
			decision:       artifacts.ReviewVerdictChangesRequested,
			requiredAction: reviewActionAddressFindings,
		},
		{
			name:           "v2 inconclusive requires rerun or human resolution",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV2,
			decision:       artifacts.ReviewVerdictInconclusive,
			requiredAction: reviewActionRerunOrHumanResolution,
		},
		{
			name:           "v2 unavailable requires rerun or human resolution",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV2,
			decision:       artifacts.ReviewVerdictUnavailable,
			requiredAction: reviewActionRerunOrHumanResolution,
		},
		{
			name:               "historical v1 remains readable",
			schemaVersion:      artifacts.ReviewVerdictSchemaVersionV1,
			decision:           artifacts.ReviewVerdictApproved,
			progressionAllowed: true,
			requiredAction:     reviewActionNone,
		},
		{
			name:           "historical v1 blocking decision remains readable",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV1,
			decision:       artifacts.ReviewVerdictChangesRequested,
			requiredAction: reviewActionAddressFindings,
		},
		{
			name:           "historical v1 advisory remains readable",
			schemaVersion:  artifacts.ReviewVerdictSchemaVersionV1,
			decision:       artifacts.ReviewVerdictEscalated,
			requiredAction: reviewActionHumanResolution,
		},
		{
			name:               "same-major v2 minor remains readable",
			schemaVersion:      "2.1.0",
			decision:           artifacts.ReviewVerdictApproved,
			progressionAllowed: true,
			requiredAction:     reviewActionNone,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			writeReviewVerdictArtifact(t, dir, test.schemaVersion, test.decision)

			collection, err := CollectDivisor(dir, time.Time{})
			if err != nil {
				t.Fatalf("CollectDivisor returned error: %v", err)
			}
			if collection == nil {
				t.Fatal("CollectDivisor returned nil collection")
			}
			if collection.DataPoints != 1 {
				t.Fatalf("DataPoints = %d, want 1", collection.DataPoints)
			}

			decisions, ok := collection.RawData["review_decisions"].([]reviewVerdictMetric)
			if !ok {
				t.Fatalf("review_decisions has type %T, want []reviewVerdictMetric", collection.RawData["review_decisions"])
			}
			if len(decisions) != 1 {
				t.Fatalf("len(review_decisions) = %d, want 1", len(decisions))
			}
			got := decisions[0]
			if got.SchemaVersion != test.schemaVersion {
				t.Errorf("SchemaVersion = %q, want %q", got.SchemaVersion, test.schemaVersion)
			}
			if got.CouncilDecision != test.decision {
				t.Errorf("CouncilDecision = %q, want %q", got.CouncilDecision, test.decision)
			}
			if got.AutomatedProgressionAllowed != test.progressionAllowed {
				t.Errorf("AutomatedProgressionAllowed = %t, want %t", got.AutomatedProgressionAllowed, test.progressionAllowed)
			}
			if got.RequiredAction != test.requiredAction {
				t.Errorf("RequiredAction = %q, want %q", got.RequiredAction, test.requiredAction)
			}
			if skipped := collection.RawData["skipped_verdict_count"]; skipped != 0 {
				t.Errorf("skipped_verdict_count = %v, want 0", skipped)
			}
		})
	}
}

func TestCollectDivisor_RejectsUnsupportedMajorAndInvalidDecision(t *testing.T) {
	dir := t.TempDir()
	writeReviewVerdictArtifact(t, dir, "3.0.0", artifacts.ReviewVerdictApproved)
	writeReviewVerdictArtifact(t, dir, artifacts.ReviewVerdictSchemaVersionV2, "APPROVE")
	writeReviewVerdictArtifact(t, dir, artifacts.ReviewVerdictSchemaVersionV1, artifacts.ReviewVerdictInconclusive)

	collection, err := CollectDivisor(dir, time.Time{})
	if err != nil {
		t.Fatalf("CollectDivisor returned error: %v", err)
	}
	if collection == nil {
		t.Fatal("CollectDivisor returned nil collection")
	}
	if collection.DataPoints != 0 {
		t.Errorf("DataPoints = %d, want 0", collection.DataPoints)
	}
	if skipped := collection.RawData["skipped_verdict_count"]; skipped != 3 {
		t.Errorf("skipped_verdict_count = %v, want 3", skipped)
	}
}

func TestReviewVerdictV1OnlyConsumer_RejectsV2Major(t *testing.T) {
	compatible, warning := artifacts.CheckSchemaVersion(
		&artifacts.Envelope{SchemaVersion: artifacts.ReviewVerdictSchemaVersionV2},
		artifacts.ReviewVerdictSchemaVersionV1,
	)
	if compatible {
		t.Error("v1-only consumer accepted review-verdict v2")
	}
	if warning == "" {
		t.Fatal("v1-only consumer rejection returned no migration warning")
	}
}

func writeReviewVerdictArtifact(
	t *testing.T,
	dir string,
	schemaVersion string,
	decision artifacts.ReviewVerdictDecision,
) {
	t.Helper()

	payload, err := json.Marshal(map[string]interface{}{
		"persona_verdicts":     []interface{}{},
		"council_decision":     decision,
		"iteration_count":      1,
		"unresolved_findings":  []interface{}{},
		"pr_url":               "https://github.com/unbound-force/unbound-force/pull/42",
		"convention_pack_used": "go",
	})
	if err != nil {
		t.Fatalf("marshal review-verdict payload: %v", err)
	}
	envelope := artifacts.Envelope{
		Hero:          "the-divisor",
		Version:       "1.0.0",
		Timestamp:     "2026-10-02T00:00:00Z",
		ArtifactType:  "review-verdict",
		SchemaVersion: schemaVersion,
		Payload:       payload,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal review-verdict envelope: %v", err)
	}
	path := filepath.Join(dir, schemaVersion+"-"+string(decision)+"-review-verdict.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write review-verdict envelope: %v", err)
	}
}
