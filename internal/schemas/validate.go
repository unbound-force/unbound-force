package schemas

import (
	"bytes"
	"fmt"
	"os"

	jschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// ValidateArtifact loads a JSON Schema file and validates a JSON
// artifact against it. Returns nil if the artifact conforms to the
// schema, or a descriptive error if validation fails. Uses
// santhosh-tekuri/jsonschema for validation (separate from the
// invopop/jsonschema generator — different concerns per SRP).
func ValidateArtifact(schemaPath, artifactPath string) error {
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema %q: %w", schemaPath, err)
	}

	artifactData, err := os.ReadFile(artifactPath)
	if err != nil {
		return fmt.Errorf("read artifact %q: %w", artifactPath, err)
	}

	return ValidateBytes(schemaData, artifactData)
}

// ValidateBytes validates JSON artifact bytes against JSON Schema
// bytes. This is the lower-level validation function used by
// ValidateArtifact and available for programmatic use.
func ValidateBytes(schemaData, artifactData []byte) error {
	schemaInst, err := jschema.UnmarshalJSON(bytes.NewReader(schemaData))
	if err != nil {
		return fmt.Errorf("parse schema JSON: %w", err)
	}

	c := jschema.NewCompiler()
	if err := c.AddResource("schema.json", schemaInst); err != nil {
		return fmt.Errorf("add schema resource: %w", err)
	}

	sch, err := c.Compile("schema.json")
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}

	artifactInst, err := jschema.UnmarshalJSON(bytes.NewReader(artifactData))
	if err != nil {
		return fmt.Errorf("parse artifact JSON: %w", err)
	}

	if err := sch.Validate(artifactInst); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
