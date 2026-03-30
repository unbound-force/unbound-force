# Research: Divisor Council Refinement

**Spec**: 019 | **Date**: 2026-03-30

## 1. Agent Boundary Refactoring

### Problem

The 5 Divisor personas have overlapping review dimensions.
Analysis of the current agent files reveals these
duplications:

| Review Dimension | Currently Covered By | Should Be Owned By |
|-----------------|---------------------|-------------------|
| Hardcoded secrets/credentials | Adversary (§5), SRE (§3) | Adversary only |
| Dependency CVEs | Adversary (§7), SRE (§2) | Adversary only |
| Test isolation | Adversary (§4), Tester (§4) | Tester only |
| Zero-waste mandate | Adversary (§1), Guard (§4) | Guard only |
| Efficiency (O(n²), allocations) | Adversary (§3) | SRE (per clarification) |
| File permissions | Adversary (§5), SRE (§3) | SRE only |
| Plan alignment | Architect (§6), Guard (§1) | Guard only |

### Decision: Exclusive Ownership Model

Each review dimension is assigned to exactly one persona.
The ownership mapping is defined in FR-005 of the spec.

**Approach**: Rather than creating a shared "ownership
registry" file that agents reference at runtime, the
ownership boundaries are encoded directly into each agent's
audit checklist. Each agent's checklist sections are
rewritten to include only their owned dimensions, with an
explicit "Out of Scope" note listing dimensions handled by
other personas.

**Rationale**: Agents are stateless Markdown files consumed
by an LLM. A separate registry would require the LLM to
read two files and cross-reference them, increasing the
chance of misinterpretation. Encoding ownership directly
in the checklist is more reliable and self-contained
(Constitution Principle I: self-describing artifacts).

### Migration Strategy

For each persona, the refactoring follows this pattern:

1. **Keep** checklist sections that are in the persona's
   exclusive domain.
2. **Remove** checklist sections that are now owned by
   another persona.
3. **Add** an "Out of Scope" section listing dimensions
   handled by other personas, to prevent the LLM from
   drifting back into those areas.
4. **Add** any new dimensions being assigned to this
   persona (e.g., efficiency → SRE).

The changes are purely subtractive (removing duplicates)
and additive (adding out-of-scope notes). No new review
logic is introduced.

## 2. Severity Standard Design

### Problem

Each Divisor persona currently defines its own severity
levels inline. The definitions are inconsistent:

- **Adversary**: CRITICAL/HIGH/MEDIUM/LOW listed in output
  format section, no definitions.
- **SRE**: Has inline definitions with examples (most
  detailed).
- **Guard**: CRITICAL/HIGH/MEDIUM/LOW listed, no
  definitions.
- **Architect**: Has a 1-10 alignment score plus severity
  levels.
- **Tester**: Has inline definitions with examples.

### Decision: Shared Convention Pack (`severity.md`)

Create a new tool-owned convention pack at
`.opencode/unbound/packs/severity.md` containing:

1. **Level definitions**: CRITICAL, HIGH, MEDIUM, LOW with
   clear boundary criteria.
2. **Domain-specific examples**: For each persona's domain,
   concrete examples of what constitutes each level.
3. **Auto-fix policy reference**: Which levels are
   auto-fixed (LOW, MEDIUM) vs. reported (HIGH, CRITICAL)
   in Spec Review Mode.

**File ownership**: Tool-owned (auto-updated by `uf init`).
No user-owned `-custom.md` counterpart — severity
definitions should not be customized per-repo to ensure
consistent behavior across the ecosystem.

**Scaffold integration**: Added to `expectedAssetPaths`,
`isToolOwned` returns `true`, `isDivisorAsset` returns
`true`, `shouldDeployPack` always returns `true` (severity
is language-agnostic, like `default.md`). The existing
`shouldDeployPack` function uses name-based matching —
`severity.md` should be added to the same branch as
`default.md` and `default-custom.md` (always-deploy
packs). If the function uses a prefix/suffix pattern
instead of an allowlist, verify it covers `severity.md`
and add an explicit case if not.

### Alternative Considered: Inline Definitions

Embedding identical severity definitions in all 5 agent
files would work but violates DRY. When definitions need
updating, all 5 files must change in lockstep. A shared
pack file is the canonical source, and each agent
references it with a single line.

### Alternative Considered: JSON Schema

A JSON schema for severity levels would enable machine
validation but adds complexity without clear benefit.
The severity pack is consumed by LLMs reading Markdown,
not by Go code parsing JSON. Markdown is the right format.

## 3. Hivemind Integration Pattern

### Problem

The Divisor agents are stateless — they review code fresh
each time with no memory of past sessions. If a previous
`/unleash` run discovered that `scaffold.go` requires a
nil guard for `initSubTools`, the next review won't know
about this learning.

### Decision: Prior Learnings Step

Add a "Prior Learnings" step at the start of each agent's
review workflow:

```markdown
### Step 0: Prior Learnings (optional)

If Hivemind MCP tools are available (`hivemind_find`):
1. Query for learnings related to the files being reviewed.
2. Include relevant learnings as "Prior Knowledge" context
   in your review.

If Hivemind is not available, skip this step with an
informational note and proceed with the standard review.
```

**Integration point**: This step runs before the Source
Documents reading step. It uses `hivemind_find` with a
query constructed from the file paths in the diff.

**Graceful degradation**: The step is wrapped in a
conditional check for tool availability. If `hivemind_find`
is not available (MCP tool not registered), the agent
skips the step entirely. This follows the established
3-tier degradation pattern used for Dewey integration.

### Alternative Considered: Coordinator-Level Lookup

Having `/review-council` perform the Hivemind lookup and
inject learnings into each agent's prompt would centralize
the logic. However, this would require the review-council
command to know which files each agent will review, which
breaks the current delegation model where agents
independently determine their review scope from
`git diff`.

## 4. golangci-lint Configuration Strategy

### Problem

Adding `golangci-lint` to CI will likely produce findings
in the existing codebase. The spec clarifies that non-zero
exit = gate failure (same as `go build`).

### Decision: Minimal Configuration

Start with `golangci-lint run` using default settings.
If the existing codebase produces findings, fix them as
part of this spec's implementation (expected and desirable
per the spec's assumptions).

**No `.golangci.yml` initially**: The default linter set
is sufficient. If specific linters need to be
enabled/disabled, a `.golangci.yml` can be added later.
The spec states "the project controls what constitutes an
error vs. warning via `.golangci.yml` configuration" —
this is a future extension point, not a requirement for
this spec.

**CI installation**: Use `golangci/golangci-lint-action`
for GitHub Actions CI (built-in caching, version pinning,
standard approach). Use `go install` for `uf setup`
(local development).

**CI workflow placement**: After `go vet` and before
`go test`. This ordering ensures basic compilation and
vet checks pass before running the more comprehensive
linter.

**`govulncheck` placement**: After `go test`. Vulnerability
checks don't need to block test execution, and running
tests first ensures the codebase is functional before
checking for CVEs.

### Installation in `uf setup`

Both tools are installable via `go install`:
- `golangci-lint`: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
  (or `brew install golangci-lint`)
- `govulncheck`: `go install golang.org/x/vuln/cmd/govulncheck@latest`

`uf setup` will use `go install` as the primary method
(since Go is already a prerequisite) with Homebrew as
fallback for `golangci-lint`.

### Phase 1a Integration

`/review-council` Phase 1a derives commands from
`.github/workflows/`. When `golangci-lint run` and
`govulncheck ./...` appear in the workflow files, Phase 1a
will execute them locally. No hardcoding in the
review-council command — it reads the workflow files
dynamically.

## 5. Legacy File Detection

### Problem

Repos that previously ran `uf init` have `reviewer-*.md`
files on disk. After this change, `uf init` no longer
deploys them, but they remain on disk. Users need to know
they should remove them.

### Decision: Soft Warning in `uf init`

After scaffolding completes, `uf init` checks for
`reviewer-*.md` files in the target's `.opencode/agents/`
directory. If found:

1. Print a warning listing the detected files.
2. Suggest a removal command:
   `rm .opencode/agents/reviewer-*.md`
3. Do NOT delete the files automatically.

**Implementation location**: In `scaffold.Run()`, after
the main file-writing loop completes and before
`printSummary()`. The detection uses `filepath.Glob` on
the target directory.

**Why not auto-delete**: The spec explicitly states "MUST
NOT delete the files" (FR-003a). Auto-deletion could
surprise users who have customized the legacy files or
use them for other purposes. The warning is informational
and non-blocking.

### Alternative Considered: `uf doctor` Check

Adding a doctor check for legacy files would also work,
but `uf init` is the natural place because it's the
command that manages scaffolded files. Users expect
`uf init` to tell them about file state.

A doctor check could be added as a future enhancement
(non-blocking, informational) but is not required by
this spec.
