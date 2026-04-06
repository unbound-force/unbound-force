package schemas_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/schemas"
)

// TestValidateConventionPack_GoPackValid validates that the actual
// Go convention pack at .opencode/unbound/packs/go.md passes
// structural validation (SC-007).
func TestValidateConventionPack_GoPackValid(t *testing.T) {
	packPath := filepath.Join("..", "..", ".opencode", "unbound", "packs", "go.md")

	if _, err := os.Stat(packPath); err != nil {
		t.Fatalf("Go convention pack not found at %s: %v", packPath, err)
	}

	if err := schemas.ValidateConventionPack(packPath); err != nil {
		t.Errorf("Go convention pack validation failed: %v", err)
	}
}

// TestValidateConventionPack_MissingSection verifies that a pack
// without a required H2 section (e.g., Coding Style) fails
// validation.
func TestValidateConventionPack_MissingSection(t *testing.T) {
	dir := t.TempDir()
	packPath := filepath.Join(dir, "incomplete.md")

	// Pack with valid frontmatter but missing "Coding Style" section
	content := `---
pack_id: test
language: Go
version: 1.0.0
---

# Convention Pack: Test

## Architectural Patterns

- AP-001 [MUST] Use dependency injection.

## Security Checks

- SC-001 [MUST] No hardcoded secrets.

## Testing Conventions

- TC-001 [MUST] Use standard testing package.

## Documentation Requirements

- DR-001 [MUST] GoDoc on exports.

## Custom Rules

<!-- none -->
`
	if err := os.WriteFile(packPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write test pack: %v", err)
	}

	err := schemas.ValidateConventionPack(packPath)
	if err == nil {
		t.Fatal("expected validation error for pack missing Coding Style section, got nil")
	}

	t.Logf("correctly rejected pack: %v", err)
}

// TestValidateConventionPack_MissingFrontmatter verifies that a
// pack without the required pack_id frontmatter field fails
// validation.
func TestValidateConventionPack_MissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	packPath := filepath.Join(dir, "no-packid.md")

	// Pack missing pack_id in frontmatter
	content := `---
language: Go
version: 1.0.0
---

# Convention Pack: Test

## Coding Style

- CS-001 [MUST] Format with gofmt.

## Architectural Patterns

- AP-001 [MUST] Use DI.

## Security Checks

- SC-001 [MUST] No secrets.

## Testing Conventions

- TC-001 [MUST] Standard testing.

## Documentation Requirements

- DR-001 [MUST] GoDoc.

## Custom Rules

<!-- none -->
`
	if err := os.WriteFile(packPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write test pack: %v", err)
	}

	err := schemas.ValidateConventionPack(packPath)
	if err == nil {
		t.Fatal("expected validation error for pack missing pack_id, got nil")
	}

	t.Logf("correctly rejected pack: %v", err)
}

// TestValidateConventionPack_NoFrontmatter verifies that a pack
// without any YAML frontmatter fails validation.
func TestValidateConventionPack_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	packPath := filepath.Join(dir, "plain.md")

	content := `# Just a plain Markdown file

No frontmatter here.
`
	if err := os.WriteFile(packPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write test pack: %v", err)
	}

	err := schemas.ValidateConventionPack(packPath)
	if err == nil {
		t.Fatal("expected validation error for pack without frontmatter, got nil")
	}

	t.Logf("correctly rejected pack: %v", err)
}

// TestValidateConventionPack_AllPacksValid validates all convention
// packs in the repo to ensure none have drifted from the required
// structure.
func TestValidateConventionPack_AllPacksValid(t *testing.T) {
	packsDir := filepath.Join("..", "..", ".opencode", "unbound", "packs")

	entries, err := os.ReadDir(packsDir)
	if err != nil {
		t.Fatalf("read packs directory: %v", err)
	}

	// Only validate non-custom packs (custom packs may be stubs)
	validated := 0
	for _, entry := range entries {
		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}
		// Skip custom packs — they are user-owned stubs that may
		// not have all required sections
		if strings.Contains(name, "-custom") {
			continue
		}
		// Skip severity pack — it's a shared severity definitions
		// pack, not a coding convention pack (Spec 019)
		if name == "severity.md" {
			continue
		}
		// Skip content pack — it's a content writing convention
		// pack with different required sections (VB, TD, BA, PR,
		// FA, FT) than coding packs (Coding Style, etc.)
		if name == "content.md" {
			continue
		}

		t.Run(name, func(t *testing.T) {
			packPath := filepath.Join(packsDir, name)
			if err := schemas.ValidateConventionPack(packPath); err != nil {
				t.Errorf("pack %s validation failed: %v", name, err)
			}
		})
		validated++
	}

	if validated == 0 {
		t.Fatal("no convention packs found to validate")
	}
	t.Logf("validated %d convention packs", validated)
}
