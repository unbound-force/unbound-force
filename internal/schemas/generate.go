// Package schemas provides JSON Schema generation from Go structs and
// validation of JSON artifacts against those schemas. Schemas are
// generated using invopop/jsonschema (draft 2020-12) and validated
// using santhosh-tekuri/jsonschema. Go structs are the source of truth;
// schemas are derived, never hand-authored.
package schemas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
)

// GenerateSchema produces a JSON Schema (draft 2020-12) from a Go
// struct value using reflection. The value should be a zero-value
// instance of the struct to generate a schema for.
func GenerateSchema(v interface{}) (*jsonschema.Schema, error) {
	r := &jsonschema.Reflector{
		DoNotReference: true,
	}
	schema := r.Reflect(v)
	if schema == nil {
		return nil, fmt.Errorf("reflect returned nil schema for %T", v)
	}
	return schema, nil
}

// WriteSchema serializes a JSON Schema to a file at the given path.
// Parent directories are created if they do not exist. The output
// is pretty-printed with 2-space indentation for readability.
func WriteSchema(schema *jsonschema.Schema, path string) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", dir, err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write schema %q: %w", path, err)
	}
	return nil
}
