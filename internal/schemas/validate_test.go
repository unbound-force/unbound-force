package schemas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateArtifact_ValidFile(t *testing.T) {
	// Create a simple schema and artifact in t.TempDir()
	dir := t.TempDir()
	schema := `{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`
	artifact := `{"name":"test"}`
	schemaPath := filepath.Join(dir, "schema.json")
	artPath := filepath.Join(dir, "artifact.json")
	os.WriteFile(schemaPath, []byte(schema), 0o644)
	os.WriteFile(artPath, []byte(artifact), 0o644)

	err := ValidateArtifact(schemaPath, artPath)
	if err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateArtifact_MissingRequired(t *testing.T) {
	dir := t.TempDir()
	schema := `{"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}`
	artifact := `{"age":42}`
	schemaPath := filepath.Join(dir, "schema.json")
	artPath := filepath.Join(dir, "artifact.json")
	os.WriteFile(schemaPath, []byte(schema), 0o644)
	os.WriteFile(artPath, []byte(artifact), 0o644)

	err := ValidateArtifact(schemaPath, artPath)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention missing field 'name', got: %v", err)
	}
}

func TestValidateArtifact_MalformedSchema(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{not json"), 0o644)
	os.WriteFile(filepath.Join(dir, "art.json"), []byte(`{}`), 0o644)

	err := ValidateArtifact(filepath.Join(dir, "bad.json"), filepath.Join(dir, "art.json"))
	if err == nil {
		t.Fatal("expected error for malformed schema")
	}
}

func TestValidateArtifact_MalformedArtifact(t *testing.T) {
	dir := t.TempDir()
	schema := `{"type":"object"}`
	os.WriteFile(filepath.Join(dir, "schema.json"), []byte(schema), 0o644)
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{not json"), 0o644)

	err := ValidateArtifact(filepath.Join(dir, "schema.json"), filepath.Join(dir, "bad.json"))
	if err == nil {
		t.Fatal("expected error for malformed artifact")
	}
}

func TestValidateArtifact_SchemaNotFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "art.json"), []byte(`{}`), 0o644)

	err := ValidateArtifact(filepath.Join(dir, "nonexistent.json"), filepath.Join(dir, "art.json"))
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}

func TestValidateBytes_Valid(t *testing.T) {
	schema := []byte(`{"type":"object","required":["x"],"properties":{"x":{"type":"integer"}}}`)
	artifact := []byte(`{"x":42}`)

	err := ValidateBytes(schema, artifact)
	if err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateBytes_Invalid(t *testing.T) {
	schema := []byte(`{"type":"object","required":["x"],"properties":{"x":{"type":"integer"}}}`)
	artifact := []byte(`{"x":"not a number"}`)

	err := ValidateBytes(schema, artifact)
	if err == nil {
		t.Fatal("expected validation error for wrong type")
	}
}
