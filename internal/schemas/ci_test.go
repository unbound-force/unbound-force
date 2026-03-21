package schemas_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

// repoSchemasDir returns the path to the repo's schemas/ directory
// relative to this test file.
func repoSchemasDir() string {
	return filepath.Join("..", "..", "schemas")
}

// TestSchemaRegistry_AllSchemasValid loads every .schema.json file
// under the repo's schemas/ directory and verifies each is valid
// JSON Schema draft 2020-12 (SC-006).
func TestSchemaRegistry_AllSchemasValid(t *testing.T) {
	schemasDir := repoSchemasDir()

	// Walk all schema files
	var schemaFiles []string
	err := filepath.Walk(schemasDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" && isSchemaFile(path) {
			schemaFiles = append(schemaFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk schemas directory: %v", err)
	}

	if len(schemaFiles) == 0 {
		t.Fatal("no schema files found in schemas/ directory")
	}

	for _, path := range schemaFiles {
		t.Run(filepath.Base(filepath.Dir(path))+"/"+filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read schema: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("schema is not valid JSON: %v", err)
			}

			// Verify it declares draft 2020-12
			schemaField, ok := parsed["$schema"]
			if !ok {
				t.Error("missing $schema field")
				return
			}
			expected := "https://json-schema.org/draft/2020-12/schema"
			if schemaField != expected {
				t.Errorf("$schema=%q, want %q", schemaField, expected)
			}
		})
	}

	t.Logf("validated %d schema files", len(schemaFiles))
}

// TestSchemaRegistry_AllSamplesValidate validates every sample
// artifact against its corresponding schema (SC-006). Expects
// samples at schemas/{type}/samples/sample-{type}.json.
func TestSchemaRegistry_AllSamplesValidate(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, typeName := range schemas.RegisteredTypeNames() {
		t.Run(typeName, func(t *testing.T) {
			schemaPath := filepath.Join(schemasDir, typeName, "v1.0.0.schema.json")
			samplePath := filepath.Join(schemasDir, typeName, "samples", "sample-"+typeName+".json")

			// Verify both files exist
			if _, err := os.Stat(schemaPath); err != nil {
				t.Fatalf("schema file missing: %v", err)
			}
			if _, err := os.Stat(samplePath); err != nil {
				t.Fatalf("sample file missing: %v", err)
			}

			if err := schemas.ValidateArtifact(schemaPath, samplePath); err != nil {
				t.Errorf("sample validation failed: %v", err)
			}
		})
	}
}

// TestSchemaRegistry_DirectoryStructure verifies the expected
// directory structure exists in the repo's schemas/ directory
// (SC-005). Each registered type must have a directory with a
// schema file, samples subdirectory, and README.
func TestSchemaRegistry_DirectoryStructure(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, typeName := range schemas.RegisteredTypeNames() {
		t.Run(typeName, func(t *testing.T) {
			typeDir := filepath.Join(schemasDir, typeName)

			// Directory exists
			info, err := os.Stat(typeDir)
			if err != nil {
				t.Fatalf("type directory missing: %v", err)
			}
			if !info.IsDir() {
				t.Fatal("expected directory, got file")
			}

			// Schema file exists
			schemaPath := filepath.Join(typeDir, "v1.0.0.schema.json")
			if _, err := os.Stat(schemaPath); err != nil {
				t.Errorf("schema file missing: %v", err)
			}

			// Samples directory exists
			samplesDir := filepath.Join(typeDir, "samples")
			if _, err := os.Stat(samplesDir); err != nil {
				t.Errorf("samples directory missing: %v", err)
			}

			// README exists
			readmePath := filepath.Join(typeDir, "README.md")
			if _, err := os.Stat(readmePath); err != nil {
				t.Errorf("README.md missing: %v", err)
			}
		})
	}
}

// isSchemaFile returns true if the path ends with .schema.json.
func isSchemaFile(path string) bool {
	return len(path) > 12 && path[len(path)-12:] == ".schema.json"
}
