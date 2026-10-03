package schemas_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
	"gopkg.in/yaml.v3"
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

// handAuthoredSchema identifies a schema that is maintained independently of
// the Go schema generator.
type handAuthoredSchema struct {
	typeName   string
	schemaFile string
	samplesDir string
}

// handAuthoredSchemas lists schemas that need dedicated fixture validation
// because they are not generated from Go structs.
var handAuthoredSchemas = []handAuthoredSchema{
	{typeName: "feedback-triage", schemaFile: "v1.0.0.schema.json"},
	{typeName: "issue-triage", schemaFile: "v1.0.0.schema.json"},
	{typeName: "lesson-proposal", schemaFile: "v1.0.0.schema.json"},
	{typeName: "review-matrix", schemaFile: "v2.schema.json"},
	{typeName: "reviewer-capabilities", schemaFile: "v1.0.0.schema.json"},
	{typeName: "review-dispatch", schemaFile: "v1.0.0.schema.json"},
	{typeName: "sibling-repos", schemaFile: "v1.0.0.schema.json"},
	{
		typeName:   "review-verdict-v2",
		schemaFile: filepath.Join("review-verdict", "v2.0.0.schema.json"),
		samplesDir: filepath.Join("review-verdict", "samples", "v2"),
	},
}

func handAuthoredPaths(schemasDir string, handAuthored handAuthoredSchema) (string, string) {
	typeDir := filepath.Join(schemasDir, handAuthored.typeName)
	schemaPath := filepath.Join(typeDir, handAuthored.schemaFile)
	samplesDir := filepath.Join(typeDir, "samples")
	if handAuthored.samplesDir != "" {
		schemaPath = filepath.Join(schemasDir, handAuthored.schemaFile)
		samplesDir = filepath.Join(schemasDir, handAuthored.samplesDir)
	}
	return schemaPath, samplesDir
}

// TestHandAuthoredSchemas_PositiveFixturesValidate validates every positive
// fixture for each hand-authored schema against its schema file.
func TestHandAuthoredSchemas_PositiveFixturesValidate(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, handAuthored := range handAuthoredSchemas {
		t.Run(handAuthored.typeName, func(t *testing.T) {
			schemaPath, samplesDir := handAuthoredPaths(schemasDir, handAuthored)

			entries, err := os.ReadDir(samplesDir)
			if err != nil {
				t.Fatalf("read samples directory: %v", err)
			}

			var positiveCount int
			for _, entry := range entries {
				if entry.IsDir() || strings.HasPrefix(entry.Name(), "invalid-") || filepath.Ext(entry.Name()) != ".json" {
					continue
				}
				positiveCount++
				fixturePath := filepath.Join(samplesDir, entry.Name())
				if err := schemas.ValidateArtifact(schemaPath, fixturePath); err != nil {
					t.Errorf("positive fixture %s failed validation: %v", entry.Name(), err)
				}
			}

			if positiveCount == 0 {
				t.Error("no positive JSON fixtures found")
			}
		})
	}
}

// TestHandAuthoredSchemas_NegativeFixturesRejected validates that
// invalid sample files are correctly rejected by the schema.
func TestHandAuthoredSchemas_NegativeFixturesRejected(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, handAuthored := range handAuthoredSchemas {
		schemaPath, samplesDir := handAuthoredPaths(schemasDir, handAuthored)

		if _, err := os.Stat(schemaPath); err != nil {
			t.Fatalf("schema file missing for %s: %v", handAuthored.typeName, err)
		}

		entries, err := os.ReadDir(samplesDir)
		if err != nil {
			t.Fatalf("read samples dir for %s: %v", handAuthored.typeName, err)
		}

		var invalidCount int
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), "invalid-") {
				continue
			}
			invalidCount++

			t.Run(handAuthored.typeName+"/"+entry.Name(), func(t *testing.T) {
				fixturePath := filepath.Join(samplesDir, entry.Name())
				err := schemas.ValidateArtifact(schemaPath, fixturePath)
				if err == nil {
					t.Errorf("expected schema to reject %s, but validation passed", entry.Name())
				}
			})
		}

		if invalidCount == 0 {
			t.Errorf("no invalid-* fixtures found for %s", handAuthored.typeName)
		}
		t.Logf("validated %d negative fixtures for %s", invalidCount, handAuthored.typeName)
	}
}

// TestHandAuthoredSchemas_DirectoryStructure verifies hand-authored
// schemas have the expected directory structure.
func TestHandAuthoredSchemas_DirectoryStructure(t *testing.T) {
	schemasDir := repoSchemasDir()

	for _, handAuthored := range handAuthoredSchemas {
		t.Run(handAuthored.typeName, func(t *testing.T) {
			typeDir := filepath.Join(schemasDir, handAuthored.typeName)
			if handAuthored.samplesDir != "" {
				typeDir = filepath.Join(schemasDir, "review-verdict")
			}

			info, err := os.Stat(typeDir)
			if err != nil {
				t.Fatalf("type directory missing: %v", err)
			}
			if !info.IsDir() {
				t.Fatal("expected directory, got file")
			}

			schemaPath, samplesDir := handAuthoredPaths(schemasDir, handAuthored)
			if _, err := os.Stat(schemaPath); err != nil {
				t.Errorf("schema file missing: %v", err)
			}

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

// TestReviewMatrixSchema_CanonicalPolicyValidates verifies the checked-in YAML
// policy with the same schema used for JSON fixtures.
func TestReviewMatrixSchema_CanonicalPolicyValidates(t *testing.T) {
	schemasDir := repoSchemasDir()
	schemaPath := filepath.Join(schemasDir, "review-matrix", "v2.schema.json")
	matrixPath := filepath.Join(schemasDir, "..", ".uf", "review-matrix.yaml")

	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read review matrix schema: %v", err)
	}
	matrixData, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read canonical review matrix: %v", err)
	}

	var matrix map[string]interface{}
	if err := yaml.Unmarshal(matrixData, &matrix); err != nil {
		t.Fatalf("parse canonical review matrix YAML: %v", err)
	}
	matrixJSON, err := json.Marshal(matrix)
	if err != nil {
		t.Fatalf("convert canonical review matrix to JSON: %v", err)
	}
	if err := schemas.ValidateBytes(schemaData, matrixJSON); err != nil {
		t.Fatalf("canonical review matrix validation failed: %v", err)
	}

	profiles, ok := matrix["profiles"].(map[string]interface{})
	if !ok {
		t.Fatal("canonical profiles are not an object")
	}
	wantModels := map[string]string{
		"lightweight": "opencode-go/qwen3.8-flash",
		"standard":    "opencode-go/deepseek-v4-pro",
		"heavy":       "opencode-go/grok-4.7",
	}
	for profileName, wantModel := range wantModels {
		profile, ok := profiles[profileName].(map[string]interface{})
		if !ok {
			t.Errorf("profile %q is not an object", profileName)
			continue
		}
		if gotModel := profile["model"]; gotModel != wantModel {
			t.Errorf("profile %q model = %v, want %q", profileName, gotModel, wantModel)
		}
	}
}

// TestReviewMatrixSchema_ExplicitRunsAndHostAbsence verifies that positive
// fixtures retain the contract scenarios they are intended to exercise.
func TestReviewMatrixSchema_ExplicitRunsAndHostAbsence(t *testing.T) {
	samplesDir := filepath.Join(repoSchemasDir(), "review-matrix", "samples")
	explicitPath := filepath.Join(samplesDir, "sample-review-matrix.json")
	explicitData, err := os.ReadFile(explicitPath)
	if err != nil {
		t.Fatalf("read explicit-runs fixture: %v", err)
	}

	var explicit struct {
		Profiles map[string]struct {
			Variant string `json:"variant"`
		} `json:"profiles"`
		Runs map[string]map[string][]struct {
			Profile string `json:"profile"`
			Model   string `json:"model"`
			Variant string `json:"variant"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(explicitData, &explicit); err != nil {
		t.Fatalf("parse explicit-runs fixture: %v", err)
	}

	runs := explicit.Runs["code"]["divisor-adversary"]
	if len(runs) != 2 {
		t.Fatalf("ordered explicit runs count = %d, want 2", len(runs))
	}
	if runs[0].Profile != "standard" || runs[0].Variant != "low" {
		t.Errorf("first run = profile %q variant %q, want standard/low", runs[0].Profile, runs[0].Variant)
	}
	if runs[1].Model != "opencode-go/grok-4.7" || runs[1].Variant != "high" {
		t.Errorf("second run = model %q variant %q, want opencode-go/grok-4.7/high", runs[1].Model, runs[1].Variant)
	}
	if got := explicit.Profiles["standard"].Variant; got != "high" || got == runs[0].Variant {
		t.Errorf("profile variant = %q and run variant = %q, want high overridden by low", got, runs[0].Variant)
	}

	hostPath := filepath.Join(samplesDir, "valid-host-fallback-absence.json")
	hostData, err := os.ReadFile(hostPath)
	if err != nil {
		t.Fatalf("read host-fallback fixture: %v", err)
	}
	var host struct {
		Advisor map[string][]string         `json:"advisor"`
		Runs    map[string]map[string][]any `json:"runs"`
	}
	if err := json.Unmarshal(hostData, &host); err != nil {
		t.Fatalf("parse host-fallback fixture: %v", err)
	}
	for mode, agents := range host.Advisor {
		for _, agent := range agents {
			if agent == "divisor-testing" {
				t.Errorf("host-fallback fixture unexpectedly configures divisor-testing in advisor mode %q", mode)
			}
		}
	}
	for mode, agentRuns := range host.Runs {
		if _, configured := agentRuns["divisor-testing"]; configured {
			t.Errorf("host-fallback fixture unexpectedly configures divisor-testing runs in mode %q", mode)
		}
	}
}

// isSchemaFile returns true if the path ends with .schema.json.
func isSchemaFile(path string) bool {
	return strings.HasSuffix(path, ".schema.json")
}
