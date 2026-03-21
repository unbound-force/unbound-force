package schemas

import (
	"fmt"
	"path/filepath"
)

// SchemaVersion is the current version for all generated schemas.
const SchemaVersion = "v1.0.0"

// schemaEntry maps an artifact type name to the Go struct used
// for schema generation.
type schemaEntry struct {
	TypeName string
	Value    interface{}
}

// registeredTypes returns the ordered list of artifact types and
// their corresponding Go struct zero-values for schema generation.
// Order is deterministic to ensure reproducible output.
func registeredTypes() []schemaEntry {
	return []schemaEntry{
		{TypeName: "envelope", Value: EnvelopeSchema{}},
		{TypeName: "quality-report", Value: QualityReportPayload{}},
		{TypeName: "review-verdict", Value: ReviewVerdictPayload{}},
		{TypeName: "backlog-item", Value: BacklogItemPayload{}},
		{TypeName: "acceptance-decision", Value: AcceptanceDecisionPayload{}},
		{TypeName: "metrics-snapshot", Value: MetricsSnapshotPayload{}},
		{TypeName: "coaching-record", Value: CoachingRecordPayload{}},
		{TypeName: "workflow-record", Value: WorkflowRecordPayload{}},
	}
}

// RegisteredTypeNames returns the list of artifact type names in
// the registry. Useful for iteration in tests and CI validation.
func RegisteredTypeNames() []string {
	entries := registeredTypes()
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.TypeName
	}
	return names
}

// GenerateAll generates JSON Schemas for all registered artifact
// types and writes them to outputDir/{type}/v1.0.0.schema.json.
// Returns nil if all schemas are generated successfully, or the
// first error encountered.
func GenerateAll(outputDir string) error {
	for _, entry := range registeredTypes() {
		schema, err := GenerateSchema(entry.Value)
		if err != nil {
			return fmt.Errorf("generate schema for %s: %w", entry.TypeName, err)
		}

		path := filepath.Join(outputDir, entry.TypeName, SchemaVersion+".schema.json")
		if err := WriteSchema(schema, path); err != nil {
			return fmt.Errorf("write schema for %s: %w", entry.TypeName, err)
		}
	}
	return nil
}
