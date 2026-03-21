package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

// TestGenerateAll_ProducesValidSchemas verifies that GenerateAll
// creates all 8 schema files (envelope + 7 artifact types) in the
// output directory. Each file must be valid JSON and contain a
// $schema field indicating draft 2020-12 (SC-001, SC-002).
func TestGenerateAll_ProducesValidSchemas(t *testing.T) {
	dir := t.TempDir()

	if err := schemas.GenerateAll(dir); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	expectedTypes := schemas.RegisteredTypeNames()
	if len(expectedTypes) != 8 {
		t.Fatalf("expected 8 registered types, got %d", len(expectedTypes))
	}

	for _, typeName := range expectedTypes {
		schemaPath := filepath.Join(dir, typeName, "v1.0.0.schema.json")
		data, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Errorf("schema file for %s not found: %v", typeName, err)
			continue
		}

		// Verify valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("schema for %s is not valid JSON: %v", typeName, err)
			continue
		}

		// Verify draft 2020-12
		schemaField, ok := parsed["$schema"]
		if !ok {
			t.Errorf("schema for %s missing $schema field", typeName)
			continue
		}
		expected := "https://json-schema.org/draft/2020-12/schema"
		if schemaField != expected {
			t.Errorf("schema for %s has $schema=%q, want %q", typeName, schemaField, expected)
		}

		// Verify type is "object"
		typeField, ok := parsed["type"]
		if !ok || typeField != "object" {
			t.Errorf("schema for %s missing or wrong type field: got %v", typeName, typeField)
		}
	}
}

// TestValidateArtifact_AllSamples validates each sample artifact
// against its corresponding generated schema. This is the primary
// round-trip test: Go struct → JSON Schema → validate sample JSON
// (SC-001, SC-002).
func TestValidateArtifact_AllSamples(t *testing.T) {
	dir := t.TempDir()

	if err := schemas.GenerateAll(dir); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	// Sample files map: type name → sample file path (relative to repo root)
	samples := map[string]string{
		"envelope":            "../../schemas/envelope/samples/sample-envelope.json",
		"quality-report":      "../../schemas/quality-report/samples/sample-quality-report.json",
		"review-verdict":      "../../schemas/review-verdict/samples/sample-review-verdict.json",
		"backlog-item":        "../../schemas/backlog-item/samples/sample-backlog-item.json",
		"acceptance-decision": "../../schemas/acceptance-decision/samples/sample-acceptance-decision.json",
		"metrics-snapshot":    "../../schemas/metrics-snapshot/samples/sample-metrics-snapshot.json",
		"coaching-record":     "../../schemas/coaching-record/samples/sample-coaching-record.json",
		"workflow-record":     "../../schemas/workflow-record/samples/sample-workflow-record.json",
	}

	for typeName, samplePath := range samples {
		t.Run(typeName, func(t *testing.T) {
			schemaPath := filepath.Join(dir, typeName, "v1.0.0.schema.json")
			if err := schemas.ValidateArtifact(schemaPath, samplePath); err != nil {
				t.Errorf("sample for %s failed validation: %v", typeName, err)
			}
		})
	}
}

// TestValidateArtifact_InvalidArtifact verifies that validation
// correctly rejects an artifact missing a required field. Tests
// US1-AS2: "Given an artifact missing the hero field, When validated,
// Then the schema reports a required property error."
func TestValidateArtifact_InvalidArtifact(t *testing.T) {
	dir := t.TempDir()

	if err := schemas.GenerateAll(dir); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	// Create an invalid envelope: missing required "hero" field
	invalidArtifact := map[string]interface{}{
		"version":        "1.0.0",
		"timestamp":      "2026-03-21T10:00:00Z",
		"artifact_type":  "quality-report",
		"schema_version": "1.0.0",
		"context":        map[string]interface{}{},
		"payload":        map[string]interface{}{},
	}

	artifactData, err := json.Marshal(invalidArtifact)
	if err != nil {
		t.Fatalf("marshal invalid artifact: %v", err)
	}

	artifactPath := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(artifactPath, artifactData, 0o644); err != nil {
		t.Fatalf("write invalid artifact: %v", err)
	}

	schemaPath := filepath.Join(dir, "envelope", "v1.0.0.schema.json")
	err = schemas.ValidateArtifact(schemaPath, artifactPath)
	if err == nil {
		t.Fatal("expected validation error for artifact missing 'hero' field, got nil")
	}
	if !strings.Contains(err.Error(), "hero") && !strings.Contains(err.Error(), "required") {
		t.Errorf("error should mention missing required field, got: %v", err)
	}

	t.Logf("correctly rejected invalid artifact: %v", err)
}

// TestGenerateSchema_ReturnsSchema verifies that GenerateSchema
// produces a non-nil schema for a simple struct.
func TestGenerateSchema_ReturnsSchema(t *testing.T) {
	type Simple struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	schema, err := schemas.GenerateSchema(Simple{})
	if err != nil {
		t.Fatalf("GenerateSchema: %v", err)
	}
	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
}

// TestWriteSchema_CreatesFile verifies that WriteSchema creates
// the output file with valid JSON content.
func TestWriteSchema_CreatesFile(t *testing.T) {
	type Tiny struct {
		X int `json:"x"`
	}

	schema, err := schemas.GenerateSchema(Tiny{})
	if err != nil {
		t.Fatalf("GenerateSchema: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test", "v1.0.0.schema.json")
	if err := schemas.WriteSchema(schema, path); err != nil {
		t.Fatalf("WriteSchema: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written schema: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("written schema is not valid JSON: %v", err)
	}

	if parsed["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Errorf("unexpected $schema: %v", parsed["$schema"])
	}
}
