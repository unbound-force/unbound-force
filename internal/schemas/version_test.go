package schemas_test

import (
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

// TestCheckCompatibility_SameVersion verifies that identical
// versions are compatible.
func TestCheckCompatibility_SameVersion(t *testing.T) {
	compatible, err := schemas.CheckCompatibility("1.0.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !compatible {
		t.Error("expected compatible=true for same version")
	}
}

// TestCheckCompatibility_MinorBump_Compatible verifies that a
// minor version bump is backward compatible (SC-003). A consumer
// expecting v1.0.0 can parse a v1.1.0 artifact.
func TestCheckCompatibility_MinorBump_Compatible(t *testing.T) {
	compatible, err := schemas.CheckCompatibility("1.1.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !compatible {
		t.Error("expected compatible=true for minor bump (1.1.0 vs 1.0.0)")
	}
}

// TestCheckCompatibility_PatchBump_Compatible verifies that a
// patch version bump is backward compatible.
func TestCheckCompatibility_PatchBump_Compatible(t *testing.T) {
	compatible, err := schemas.CheckCompatibility("1.0.1", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !compatible {
		t.Error("expected compatible=true for patch bump (1.0.1 vs 1.0.0)")
	}
}

// TestCheckCompatibility_MajorBump_Incompatible verifies that a
// major version mismatch is incompatible (SC-004). The error
// message includes migration guidance per FR-007.
func TestCheckCompatibility_MajorBump_Incompatible(t *testing.T) {
	compatible, err := schemas.CheckCompatibility("2.0.0", "1.0.0")
	if err == nil {
		t.Fatal("expected error for major version mismatch, got nil")
	}
	if compatible {
		t.Error("expected compatible=false for major bump (2.0.0 vs 1.0.0)")
	}

	// Verify error contains migration guidance
	errMsg := err.Error()
	if !containsAll(errMsg, "incompatible", "MAJOR", "migration") {
		t.Errorf("error message should contain migration guidance, got: %s", errMsg)
	}
}

// TestCheckCompatibility_ConsumerAhead_Compatible verifies that
// a consumer with a higher minor version can still read a lower
// minor version artifact (backward compatible).
func TestCheckCompatibility_ConsumerAhead_Compatible(t *testing.T) {
	compatible, err := schemas.CheckCompatibility("1.0.0", "1.2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !compatible {
		t.Error("expected compatible=true when consumer minor is ahead")
	}
}

// TestCheckCompatibility_InvalidVersion verifies that malformed
// version strings produce a clear error.
func TestCheckCompatibility_InvalidVersion(t *testing.T) {
	tests := []struct {
		name     string
		producer string
		consumer string
	}{
		{"empty producer", "", "1.0.0"},
		{"empty consumer", "1.0.0", ""},
		{"non-numeric", "abc", "1.0.0"},
		{"missing patch", "1.0", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := schemas.CheckCompatibility(tt.producer, tt.consumer)
			if err == nil {
				t.Errorf("expected error for producer=%q consumer=%q", tt.producer, tt.consumer)
			}
		})
	}
}

// containsAll checks if s contains all of the given substrings
// (case-insensitive).
func containsAll(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if !strings.Contains(lower, strings.ToLower(sub)) {
			return false
		}
	}
	return true
}
