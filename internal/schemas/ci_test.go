package schemas_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// TestSchemaRegistry_NoDrift generates schemas to a temp directory
// and compares them against the committed versions. Any difference
// indicates the Go structs changed without regenerating schemas.
func TestSchemaRegistry_NoDrift(t *testing.T) {
	// Generate schemas to temp dir
	tmpDir := t.TempDir()
	if err := schemas.GenerateAll(tmpDir); err != nil {
		t.Fatalf("GenerateAll: %v", err)
	}

	// Compare each generated schema against committed version
	repoDir := repoSchemasDir()
	for _, typeName := range schemas.RegisteredTypeNames() {
		generated := filepath.Join(tmpDir, typeName, "v1.0.0.schema.json")
		committed := filepath.Join(repoDir, typeName, "v1.0.0.schema.json")

		genBytes, err := os.ReadFile(generated)
		if err != nil {
			t.Fatalf("read generated %s: %v", typeName, err)
		}
		comBytes, err := os.ReadFile(committed)
		if err != nil {
			t.Fatalf("read committed %s: %v", typeName, err)
		}

		if !bytes.Equal(genBytes, comBytes) {
			t.Errorf("schema drift detected for %q: committed schema differs from generated. Run GenerateAll to update.", typeName)
		}
	}
}

// handAuthoredSchemas lists schema types that are hand-authored
// (not generated from Go structs). These need dedicated validation
// tests since they are not covered by the registry-based tests.
var handAuthoredSchemas = []string{
	"feedback-triage",
}

// TestHandAuthoredSchemas_SampleValidates validates the positive
// sample for each hand-authored schema against its schema file.
func TestHandAuthoredSchemas_SampleValidates(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, typeName := range handAuthoredSchemas {
		t.Run(typeName, func(t *testing.T) {
			schemaPath := filepath.Join(schemasDir, typeName, "v1.0.0.schema.json")
			samplePath := filepath.Join(schemasDir, typeName, "samples", "sample-"+typeName+".json")

			if _, err := os.Stat(schemaPath); err != nil {
				t.Fatalf("schema file missing: %v", err)
			}
			if _, err := os.Stat(samplePath); err != nil {
				t.Fatalf("sample file missing: %v", err)
			}

			if err := schemas.ValidateArtifact(schemaPath, samplePath); err != nil {
				t.Errorf("positive sample validation failed: %v", err)
			}
		})
	}
}

// TestHandAuthoredSchemas_NegativeFixturesRejected validates that
// invalid sample files are correctly rejected by the schema.
func TestHandAuthoredSchemas_NegativeFixturesRejected(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, typeName := range handAuthoredSchemas {
		samplesDir := filepath.Join(schemasDir, typeName, "samples")
		schemaPath := filepath.Join(schemasDir, typeName, "v1.0.0.schema.json")

		if _, err := os.Stat(schemaPath); err != nil {
			t.Fatalf("schema file missing for %s: %v", typeName, err)
		}

		entries, err := os.ReadDir(samplesDir)
		if err != nil {
			t.Fatalf("read samples dir for %s: %v", typeName, err)
		}

		var invalidCount int
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), "invalid-") {
				continue
			}
			invalidCount++

			t.Run(typeName+"/"+entry.Name(), func(t *testing.T) {
				fixturePath := filepath.Join(samplesDir, entry.Name())
				err := schemas.ValidateArtifact(schemaPath, fixturePath)
				if err == nil {
					t.Errorf("expected schema to reject %s, but validation passed", entry.Name())
				}
			})
		}

		if invalidCount == 0 {
			t.Errorf("no invalid-* fixtures found for %s", typeName)
		}
		t.Logf("validated %d negative fixtures for %s", invalidCount, typeName)
	}
}

// TestHandAuthoredSchemas_DirectoryStructure verifies hand-authored
// schemas have the expected directory structure.
func TestHandAuthoredSchemas_DirectoryStructure(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, typeName := range handAuthoredSchemas {
		t.Run(typeName, func(t *testing.T) {
			typeDir := filepath.Join(schemasDir, typeName)

			info, err := os.Stat(typeDir)
			if err != nil {
				t.Fatalf("type directory missing: %v", err)
			}
			if !info.IsDir() {
				t.Fatal("expected directory, got file")
			}

			schemaPath := filepath.Join(typeDir, "v1.0.0.schema.json")
			if _, err := os.Stat(schemaPath); err != nil {
				t.Errorf("schema file missing: %v", err)
			}

			samplesDir := filepath.Join(typeDir, "samples")
			if _, err := os.Stat(samplesDir); err != nil {
				t.Errorf("samples directory missing: %v", err)
			}

			readmePath := filepath.Join(typeDir, "README.md")
			if _, err := os.Stat(readmePath); err != nil {
				t.Errorf("README.md missing: %v", err)
			}
		})
	}
}

// isSchemaFile returns true if the path ends with .schema.json.
func isSchemaFile(path string) bool {
	return strings.HasSuffix(path, ".schema.json")
}
