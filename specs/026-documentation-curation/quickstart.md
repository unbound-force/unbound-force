# Quickstart: Documentation Curation

**Branch**: `026-documentation-curation` | **Date**: 2026-04-11

## Prerequisites

- Go 1.24+ (for running scaffold tests)
- GitHub CLI (`gh`) installed and authenticated
  (installed by `uf setup`)
- Access to `unbound-force/website` repo (for verifying
  issue filing — optional for implementation)

## Implementation Steps

### Step 1: Create the Curator Agent

Create `.opencode/agents/divisor-curator.md` following
the established `divisor-*.md` pattern. Key differences
from standard Divisor agents:

- `bash: true` (exception — documented restriction)
- `temperature: 0.2` (slightly higher for judgment calls)
- `read: true` explicitly listed
- Includes "Bash Access Restriction" section
- Includes user-facing change detection heuristic

Reference files for pattern:
- `.opencode/agents/divisor-guard.md` (closest pattern)
- `.opencode/agents/divisor-adversary.md` (has explicit
  `read: true`)
- `.opencode/uf/packs/content.md` (content standards)

### Step 2: Enhance the Guard Agent

Add `#### 6. Documentation Completeness` to the Guard's
Code Review audit checklist (after section 5,
"Gatekeeping Integrity"). The check verifies:

- Was AGENTS.md updated when user-facing behavior
  changed?
- Was README.md updated if the change affects the
  project description?
- Skip for internal-only changes (refactoring,
  test-only, CI-only)

Severity: MEDIUM (per spec FR-012).

### Step 3: Update review-council.md Reference Table

Add a row for the Curator in the Known Divisor Persona
Roles reference table:

```markdown
| `divisor-curator` | The Curator | Documentation gaps, blog/tutorial opportunities, website issue filing | Documentation completeness in specs, content coverage |
```

### Step 4: Create Scaffold Asset Copies

```bash
# Copy new Curator agent to scaffold assets
cp .opencode/agents/divisor-curator.md \
   internal/scaffold/assets/opencode/agents/divisor-curator.md

# Sync modified Guard agent to scaffold assets
cp .opencode/agents/divisor-guard.md \
   internal/scaffold/assets/opencode/agents/divisor-guard.md

# Sync modified review-council command to scaffold assets
cp .opencode/command/review-council.md \
   internal/scaffold/assets/opencode/command/review-council.md
```

### Step 5: Update expectedAssetPaths

In `internal/scaffold/scaffold_test.go`, add the new
entry to `expectedAssetPaths`:

```go
// After "opencode/agents/divisor-envoy.md",
"opencode/agents/divisor-curator.md",
```

The total asset count increases from 54 to 55.

### Step 6: Update AGENTS.md

- Add Curator to the Heroes table (or Divisor agent
  listing)
- Update Project Structure to include
  `divisor-curator.md`
- Add Recent Changes entry

### Step 7: Run Tests

```bash
# Verify all tests pass
go test -race -count=1 ./...

# Key tests to watch:
# - TestAssetPaths_MatchExpected (asset count)
# - TestScaffoldOutput_NoOldPathReferences (no stale refs)
# - TestRun_CreatesFiles (scaffold creates new file)
```

## Verification

After implementation, verify:

1. `go test -race -count=1 ./...` passes
2. `divisor-curator.md` exists in both live and asset
   locations
3. Guard's Code Review section has 6 audit items
   (was 5)
4. `review-council.md` reference table has 6 rows
   (was 5)
5. `expectedAssetPaths` has 55 entries (was 54)
6. AGENTS.md mentions the Curator
