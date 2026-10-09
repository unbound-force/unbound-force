package schemas_test

import (
	"path/filepath"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

func TestLessonProposalSchema_ValidFixtureAccepted(t *testing.T) {
	schemaPath := filepath.Join(repoSchemasDir(), "lesson-proposal", "v1.0.0.schema.json")
	fixturePath := filepath.Join(repoSchemasDir(), "lesson-proposal", "samples", "sample-lesson-proposal.json")

	if err := schemas.ValidateArtifact(schemaPath, fixturePath); err != nil {
		t.Fatalf("validate lesson proposal fixture: %v", err)
	}
}

func TestLessonProposalSchema_InvalidFixturesRejected(t *testing.T) {
	schemaPath := filepath.Join(repoSchemasDir(), "lesson-proposal", "v1.0.0.schema.json")
	fixtures := []string{
		"invalid-empty-sources.json",
		"invalid-extra-property.json",
		"invalid-hash.json",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			fixturePath := filepath.Join(repoSchemasDir(), "lesson-proposal", "samples", fixture)
			if err := schemas.ValidateArtifact(schemaPath, fixturePath); err == nil {
				t.Fatalf("expected schema to reject %s", fixture)
			}
		})
	}
}
