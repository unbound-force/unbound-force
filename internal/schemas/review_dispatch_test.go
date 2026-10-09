package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/unbound-force/unbound-force/internal/artifacts"
	"github.com/unbound-force/unbound-force/internal/schemas"
)

func TestReviewDispatchEnvelope_ValidHICSample(t *testing.T) {
	schemasDir := repoSchemasDir()
	envelopePath := filepath.Join(schemasDir, "samples", "sample-review-dispatch-envelope.json")
	if err := schemas.ValidateArtifact(
		filepath.Join(schemasDir, "envelope", "v1.0.0.schema.json"),
		envelopePath,
	); err != nil {
		t.Fatalf("validate envelope: %v", err)
	}

	data, err := os.ReadFile(envelopePath)
	if err != nil {
		t.Fatalf("read envelope sample: %v", err)
	}
	var envelope artifacts.Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope sample: %v", err)
	}
	if envelope.Hero != "the-divisor" {
		t.Errorf("hero = %q, want the-divisor", envelope.Hero)
	}
	if envelope.ArtifactType != "review-dispatch" {
		t.Errorf("artifact_type = %q, want review-dispatch", envelope.ArtifactType)
	}
	if envelope.SchemaVersion != "1.0.0" {
		t.Errorf("schema_version = %q, want 1.0.0", envelope.SchemaVersion)
	}

	schemaData, err := os.ReadFile(filepath.Join(
		schemasDir, "review-dispatch", "v1.0.0.schema.json",
	))
	if err != nil {
		t.Fatalf("read review-dispatch schema: %v", err)
	}
	if err := schemas.ValidateBytes(schemaData, envelope.Payload); err != nil {
		t.Fatalf("validate envelope payload: %v", err)
	}
}

func TestReviewDispatchSchema_RunStateAndWorkflowBoundaries(t *testing.T) {
	schemaData, err := os.ReadFile(filepath.Join(
		repoSchemasDir(), "review-dispatch", "v1.0.0.schema.json",
	))
	if err != nil {
		t.Fatalf("read review-dispatch schema: %v", err)
	}
	sampleData, err := os.ReadFile(filepath.Join(
		repoSchemasDir(), "review-dispatch", "samples", "sample-review-dispatch.json",
	))
	if err != nil {
		t.Fatalf("read review-dispatch sample: %v", err)
	}

	var sample map[string]any
	if err := json.Unmarshal(sampleData, &sample); err != nil {
		t.Fatalf("decode review-dispatch sample: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "success requires null error",
			mutate: func(payload map[string]any) {
				run := payload["runs"].([]any)[0].(map[string]any)
				run["error"] = map[string]any{
					"code": "unexpected", "message": "failure", "retryable": false,
				}
			},
		},
		{
			name: "failed requires error",
			mutate: func(payload map[string]any) {
				run := payload["runs"].([]any)[0].(map[string]any)
				run["status"] = "failed"
				run["workflow_verdict"] = nil
				run["findings"] = []any{}
				run["error"] = nil
			},
		},
		{
			name: "non-success cannot vote",
			mutate: func(payload map[string]any) {
				run := payload["runs"].([]any)[0].(map[string]any)
				run["status"] = "skipped"
				run["started_at"] = nil
				run["finished_at"] = nil
				run["usage"] = nil
			},
		},
		{
			name: "workflow-specific verdict",
			mutate: func(payload map[string]any) {
				payload["command"] = "triage-issue"
				payload["mode"] = "triage"
				payload["workflow_result"] = map[string]any{
					"kind": "triage", "value": "VALID",
				}
			},
		},
		{
			name: "host request provenance is null",
			mutate: func(payload map[string]any) {
				run := payload["runs"].([]any)[0].(map[string]any)
				run["source"] = "host"
			},
		},
		{
			name: "contributing run ids are unique",
			mutate: func(payload map[string]any) {
				finding := payload["findings"].([]any)[0].(map[string]any)
				id := finding["run_ids"].([]any)[0]
				finding["run_ids"] = []any{id, id}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var payload map[string]any
			encoded, marshalErr := json.Marshal(sample)
			if marshalErr != nil {
				t.Fatalf("copy sample: %v", marshalErr)
			}
			if unmarshalErr := json.Unmarshal(encoded, &payload); unmarshalErr != nil {
				t.Fatalf("decode sample copy: %v", unmarshalErr)
			}
			test.mutate(payload)
			invalid, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				t.Fatalf("encode invalid payload: %v", marshalErr)
			}
			if err := schemas.ValidateBytes(schemaData, invalid); err == nil {
				t.Fatal("expected schema rejection")
			}
		})
	}
}

func TestReviewDispatchSchema_AllRunStatusesValidate(t *testing.T) {
	schemaData, sample := readDispatchSchemaAndSample(t)
	tests := []struct {
		status string
		apply  func(map[string]any)
	}{
		{
			status: "pending",
			apply: func(run map[string]any) {
				run["started_at"] = nil
				run["finished_at"] = nil
				run["usage"] = nil
				run["workflow_verdict"] = nil
				run["findings"] = []any{}
			},
		},
		{
			status: "running",
			apply: func(run map[string]any) {
				run["finished_at"] = nil
				run["usage"] = nil
				run["workflow_verdict"] = nil
				run["findings"] = []any{}
			},
		},
		{status: "success", apply: func(map[string]any) {}},
		{
			status: "failed",
			apply: func(run map[string]any) {
				run["error"] = map[string]any{
					"code":      "provider_unavailable",
					"message":   "Provider unavailable.",
					"retryable": true,
				}
				run["workflow_verdict"] = nil
				run["findings"] = []any{}
			},
		},
		{
			status: "skipped",
			apply:  applySkippedRun,
		},
		{
			status: "budget_skipped",
			apply:  applySkippedRun,
		},
		{
			status: "limit_skipped",
			apply:  applySkippedRun,
		},
		{
			status: "cancelled",
			apply: func(run map[string]any) {
				run["error"] = map[string]any{
					"code":      "cancelled",
					"message":   "Run cancelled.",
					"retryable": false,
				}
				run["workflow_verdict"] = nil
				run["findings"] = []any{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			payload := cloneJSONMap(t, sample)
			run := payload["runs"].([]any)[0].(map[string]any)
			run["status"] = test.status
			test.apply(run)
			if err := validateJSONMap(schemaData, payload); err != nil {
				t.Fatalf("status %q failed structural validation: %v", test.status, err)
			}
		})
	}
}

func TestReviewDispatchSchema_WorkflowVerdictsValidate(t *testing.T) {
	schemaData, sample := readDispatchSchemaAndSample(t)
	tests := []struct {
		name       string
		command    string
		mode       string
		runVerdict string
		resultKind string
		result     string
	}{
		{
			name:       "council",
			command:    "review-council",
			mode:       "code",
			runVerdict: "APPROVE WITH ADVISORIES",
			resultKind: "council",
			result:     "APPROVE WITH ADVISORIES",
		},
		{
			name:       "triage",
			command:    "triage-issue",
			mode:       "triage",
			runVerdict: "NEEDS-CLARIFICATION",
			resultKind: "triage",
			result:     "NEEDS-CLARIFICATION",
		},
		{
			name:       "feedback",
			command:    "address-feedback",
			mode:       "feedback",
			runVerdict: "AUTHOR-DECIDES",
			resultKind: "feedback",
			result:     "AUTHOR-DECIDES",
		},
		{
			name:       "test review",
			command:    "speckit-testreview",
			mode:       "test",
			runVerdict: "APPROVE",
			resultKind: "test-review",
			result:     "APPROVE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := cloneJSONMap(t, sample)
			payload["command"] = test.command
			payload["mode"] = test.mode
			payload["workflow_result"] = map[string]any{
				"kind":  test.resultKind,
				"value": test.result,
			}
			run := payload["runs"].([]any)[0].(map[string]any)
			run["workflow_verdict"] = test.runVerdict
			if err := validateJSONMap(schemaData, payload); err != nil {
				t.Fatalf("workflow %q failed validation: %v", test.name, err)
			}
		})
	}
}

func TestReviewDispatchSchema_DefersSemanticMappingsAndArithmetic(t *testing.T) {
	schemaData, sample := readDispatchSchemaAndSample(t)
	payload := cloneJSONMap(t, sample)
	payload["verdict"] = "APPROVE"
	payload["workflow_result"] = map[string]any{
		"kind":  "council",
		"value": "REQUEST CHANGES",
	}
	payload["run_counts"] = map[string]any{
		"total":          99,
		"success":        1,
		"failed":         1,
		"skipped":        1,
		"budget_skipped": 1,
		"limit_skipped":  1,
		"cancelled":      1,
	}
	payload["coverage"] = map[string]any{
		"preflight_verdict": "PASS",
		"checks_total":      1,
		"checks_passed":     2,
	}

	if err := validateJSONMap(schemaData, payload); err != nil {
		t.Fatalf("structural schema must defer semantic validation: %v", err)
	}
}

func readDispatchSchemaAndSample(t *testing.T) ([]byte, map[string]any) {
	t.Helper()
	base := filepath.Join(repoSchemasDir(), "review-dispatch")
	schemaData, err := os.ReadFile(filepath.Join(base, "v1.0.0.schema.json"))
	if err != nil {
		t.Fatalf("read review-dispatch schema: %v", err)
	}
	sampleData, err := os.ReadFile(filepath.Join(base, "samples", "sample-review-dispatch.json"))
	if err != nil {
		t.Fatalf("read review-dispatch sample: %v", err)
	}
	var sample map[string]any
	if err := json.Unmarshal(sampleData, &sample); err != nil {
		t.Fatalf("decode review-dispatch sample: %v", err)
	}
	return schemaData, sample
}

func cloneJSONMap(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("encode JSON map: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(encoded, &clone); err != nil {
		t.Fatalf("decode JSON map: %v", err)
	}
	return clone
}

func validateJSONMap(schemaData []byte, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return schemas.ValidateBytes(schemaData, data)
}

func applySkippedRun(run map[string]any) {
	run["usage"] = nil
	run["workflow_verdict"] = nil
	run["findings"] = []any{}
}

func TestReviewVerdictSchemas_HistoricalAndCurrentReads(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		samplePath string
		read       func([]byte) error
	}{
		{
			name:       "historical v1",
			version:    "v1.0.0.schema.json",
			samplePath: filepath.Join("samples", "sample-review-verdict.json"),
			read: func(data []byte) error {
				var payload schemas.ReviewVerdictPayload
				return json.Unmarshal(data, &payload)
			},
		},
		{
			name:       "current v2",
			version:    "v2.0.0.schema.json",
			samplePath: filepath.Join("samples", "v2", "sample-review-verdict.json"),
			read: func(data []byte) error {
				var payload schemas.ReviewVerdictV2Payload
				if err := json.Unmarshal(data, &payload); err != nil {
					return err
				}
				if payload.CouncilDecision != schemas.ReviewVerdictUnavailable {
					return &unexpectedDecisionError{got: payload.CouncilDecision}
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := filepath.Join(repoSchemasDir(), "review-verdict")
			data, err := os.ReadFile(filepath.Join(base, test.samplePath))
			if err != nil {
				t.Fatalf("read sample: %v", err)
			}
			if err := schemas.ValidateArtifact(
				filepath.Join(base, test.version), filepath.Join(base, test.samplePath),
			); err != nil {
				t.Fatalf("validate sample: %v", err)
			}
			if err := test.read(data); err != nil {
				t.Fatalf("read payload: %v", err)
			}
		})
	}
}

type unexpectedDecisionError struct {
	got schemas.ReviewVerdictDecision
}

func (e *unexpectedDecisionError) Error() string {
	return "unexpected review decision: " + string(e.got)
}

func TestReviewVerdictCompatibility_MajorBoundary(t *testing.T) {
	tests := []struct {
		name       string
		producer   string
		consumer   string
		compatible bool
	}{
		{
			name: "same v2 major", producer: "2.1.0",
			consumer: schemas.ReviewVerdictSchemaVersionV2, compatible: true,
		},
		{
			name:       "v1-only consumer rejects v2",
			producer:   schemas.ReviewVerdictSchemaVersionV2,
			consumer:   schemas.ReviewVerdictSchemaVersionV1,
			compatible: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compatible, err := schemas.CheckCompatibility(test.producer, test.consumer)
			if compatible != test.compatible {
				t.Errorf("compatible = %t, want %t", compatible, test.compatible)
			}
			if test.compatible && err != nil {
				t.Errorf("same-major compatibility error: %v", err)
			}
			if !test.compatible && err == nil {
				t.Error("major mismatch must return migration error")
			}
		})
	}
}
