package schemas_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

type consolidationFixtureOperation struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	From      string          `json:"from"`
	Value     json.RawMessage `json:"value"`
}

type consolidationFixtureCase struct {
	ID          string                          `json:"id"`
	Templates   []string                        `json:"templates"`
	Operations  []consolidationFixtureOperation `json:"operations"`
	SchemaValid bool                            `json:"schema_valid"`
}

type consolidationFixtureSuite struct {
	BaseFixture string                                     `json:"base_fixture"`
	Templates   map[string][]consolidationFixtureOperation `json:"templates"`
	Cases       []consolidationFixtureCase                 `json:"cases"`
}

func TestReviewDispatchConsolidationFixtures_StructuralSchemaBoundary(t *testing.T) {
	repositoryRoot := filepath.Join(repoSchemasDir(), "..")
	fixturePath := filepath.Join(
		repositoryRoot,
		".opencode", "test", "fixtures", "review-dispatch", "consolidation-cases.json",
	)
	fixtureData, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read consolidation fixtures: %v", err)
	}
	var suite consolidationFixtureSuite
	if err := json.Unmarshal(fixtureData, &suite); err != nil {
		t.Fatalf("decode consolidation fixtures: %v", err)
	}
	if suite.BaseFixture != "schemas/review-dispatch/samples/sample-review-dispatch.json" {
		t.Fatalf("base_fixture = %q, want registered review-dispatch sample", suite.BaseFixture)
	}

	baseData, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(suite.BaseFixture)))
	if err != nil {
		t.Fatalf("read consolidation base fixture: %v", err)
	}
	var base map[string]any
	if err := json.Unmarshal(baseData, &base); err != nil {
		t.Fatalf("decode consolidation base fixture: %v", err)
	}
	schemaData, err := os.ReadFile(filepath.Join(repoSchemasDir(), "review-dispatch", "v1.0.0.schema.json"))
	if err != nil {
		t.Fatalf("read review-dispatch schema: %v", err)
	}

	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			payload := cloneFixtureValue(t, base).(map[string]any)
			for _, templateName := range testCase.Templates {
				template, ok := suite.Templates[templateName]
				if !ok {
					t.Fatalf("unknown consolidation fixture template %q", templateName)
				}
				applyFixtureOperations(t, payload, template)
			}
			applyFixtureOperations(t, payload, testCase.Operations)

			encoded, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("encode materialized fixture: %v", err)
			}
			validationErr := schemas.ValidateBytes(schemaData, encoded)
			if testCase.SchemaValid && validationErr != nil {
				t.Fatalf("structural schema rejected fixture assigned to semantic validation: %v", validationErr)
			}
			if !testCase.SchemaValid && validationErr == nil {
				t.Fatal("structural schema accepted fixture assigned to shape validation")
			}
		})
	}
}

func applyFixtureOperations(t *testing.T, root any, operations []consolidationFixtureOperation) {
	t.Helper()
	for _, operation := range operations {
		var value any
		switch operation.Operation {
		case "set":
			if len(operation.Value) == 0 {
				t.Fatalf("set operation %q has no value", operation.Path)
			}
			if err := json.Unmarshal(operation.Value, &value); err != nil {
				t.Fatalf("decode value for %q: %v", operation.Path, err)
			}
		case "copy":
			var err error
			value, err = readFixturePointer(root, operation.From)
			if err != nil {
				t.Fatalf("copy %q: %v", operation.From, err)
			}
		default:
			t.Fatalf("unsupported fixture operation %q", operation.Operation)
		}
		if err := setFixturePointer(root, operation.Path, cloneFixtureValue(t, value)); err != nil {
			t.Fatalf("set %q: %v", operation.Path, err)
		}
	}
}

func cloneFixtureValue(t *testing.T, value any) any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode fixture value: %v", err)
	}
	var clone any
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("decode fixture value: %v", err)
	}
	return clone
}

func fixturePointerSegments(pointer string) ([]string, error) {
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("pointer must start with /: %q", pointer)
	}
	parts := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for index, part := range parts {
		parts[index] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
	}
	return parts, nil
}

func fixtureArrayIndex(segment string, length int, allowAppend bool) (int, error) {
	if segment == "" || (len(segment) > 1 && segment[0] == '0') {
		return 0, fmt.Errorf("invalid array index %q", segment)
	}
	index, err := strconv.Atoi(segment)
	if err != nil || index < 0 {
		return 0, fmt.Errorf("invalid array index %q", segment)
	}
	maximum := length - 1
	if allowAppend {
		maximum = length
	}
	if index > maximum {
		return 0, fmt.Errorf("array index %d is outside 0..%d", index, maximum)
	}
	return index, nil
}

func readFixturePointer(root any, pointer string) (any, error) {
	segments, err := fixturePointerSegments(pointer)
	if err != nil {
		return nil, err
	}
	current := root
	for _, segment := range segments {
		switch container := current.(type) {
		case map[string]any:
			value, ok := container[segment]
			if !ok {
				return nil, fmt.Errorf("property %q does not exist", segment)
			}
			current = value
		case []any:
			index, indexErr := fixtureArrayIndex(segment, len(container), false)
			if indexErr != nil {
				return nil, indexErr
			}
			current = container[index]
		default:
			return nil, fmt.Errorf("segment %q parent is not a container", segment)
		}
	}
	return current, nil
}

func setFixturePointer(root any, pointer string, value any) error {
	segments, err := fixturePointerSegments(pointer)
	if err != nil {
		return err
	}
	if len(segments) == 0 {
		return fmt.Errorf("pointer must identify a property")
	}
	_, err = setFixtureValue(root, segments, value)
	return err
}

func setFixtureValue(current any, segments []string, value any) (any, error) {
	segment := segments[0]
	last := len(segments) == 1
	switch container := current.(type) {
	case map[string]any:
		if last {
			container[segment] = value
			return container, nil
		}
		child, ok := container[segment]
		if !ok {
			return nil, fmt.Errorf("property %q does not exist", segment)
		}
		updated, err := setFixtureValue(child, segments[1:], value)
		if err != nil {
			return nil, err
		}
		container[segment] = updated
		return container, nil
	case []any:
		index, err := fixtureArrayIndex(segment, len(container), last)
		if err != nil {
			return nil, err
		}
		if last {
			if index == len(container) {
				return append(container, value), nil
			}
			container[index] = value
			return container, nil
		}
		updated, err := setFixtureValue(container[index], segments[1:], value)
		if err != nil {
			return nil, err
		}
		container[index] = updated
		return container, nil
	default:
		return nil, fmt.Errorf("segment %q parent is not a container", segment)
	}
}
