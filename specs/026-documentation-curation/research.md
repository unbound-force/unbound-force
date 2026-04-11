# Research: Documentation Curation

**Branch**: `026-documentation-curation` | **Date**: 2026-04-11

## R01: Existing Divisor Agent Pattern Analysis

**Question**: What is the canonical structure of a
`divisor-*.md` agent file?

**Findings**: All 8 existing Divisor agents follow a
consistent pattern:

```yaml
---
description: "<role summary>"
mode: subagent
model: google-vertex-anthropic/claude-opus-4-6@default
temperature: 0.1
tools:
  write: false
  edit: false
  bash: false
---
```

Followed by:
1. `# Role: <Persona Name>` — one-line role description
   with exclusive domain declaration
2. Mode announcement (Code Review / Spec Review)
3. `## Step 0: Prior Learnings (optional)` — Dewey
   integration with graceful degradation
4. `## Source Documents` — numbered list of files to read
5. `## Code Review Mode` — review scope + audit checklist
6. `## Spec Review Mode` — review scope + audit checklist
7. `## Output Format` — finding template
8. `## Decision Criteria` — APPROVE / REQUEST CHANGES
9. `## Out of Scope` — explicit ownership boundaries

**Exceptions noted**:
- Content agents (Scribe, Herald, Envoy) have
  `write: true` and `edit: true` because they produce
  content, not review findings.
- The Adversary has `read: true` explicitly listed.
- No existing Divisor agent has `bash: true`.

**Decision**: The Curator follows the standard review
agent pattern (subagent, read-only) but adds
`bash: true` as a documented exception for `gh` CLI
operations. Temperature 0.2 for judgment calls on
content significance.

## R02: Review Council Discovery Mechanism

**Question**: How does `/review-council` discover and
invoke Divisor agents?

**Findings**: From `review-council.md` lines 94-116:

1. Reads `.opencode/agents/` directory listing
2. Filters for files matching `divisor-*.md`
3. Extracts agent names by stripping `.md`
4. Delegates to each discovered agent in parallel
5. Uses the Known Persona Roles reference table for
   targeted context, but discovery is independent of
   the table

**Implication**: Adding `divisor-curator.md` to
`.opencode/agents/` automatically includes the Curator
in every review. No command changes needed. However,
the reference table SHOULD be updated to provide the
Curator's focus areas for targeted delegation.

**Decision**: Update the reference table in
`review-council.md` with the Curator's row. This is
informational — the Curator works without it, but
targeted context improves review quality.

## R03: Scaffold Engine Asset Registration

**Question**: What changes are needed to register a new
Divisor agent in the scaffold engine?

**Findings**: From `scaffold.go` and `scaffold_test.go`:

1. `isDivisorAsset()` (line 310) matches any path with
   prefix `opencode/agents/divisor-`. No code change
   needed — the Curator matches automatically.
2. `expectedAssetPaths` (line 87) is a canonical list
   of all embedded assets. Must add the new entry.
3. Current count: 54 files (11 agents, 15 commands,
   9 packs, 6 templates, 1 config, 5 scripts, 5 OpenSpec
   schema files, 1 OpenSpec config, 1 Swarm skill).
4. After adding `divisor-curator.md`: 55 files.

**Decision**: Add one entry to `expectedAssetPaths`:
`"opencode/agents/divisor-curator.md"`. Copy the live
agent file to `internal/scaffold/assets/opencode/agents/
divisor-curator.md`. Sync the modified Guard asset copy.

## R04: Guard Agent Current Structure

**Question**: Where should the "Documentation
Completeness" checklist item be added in the Guard?

**Findings**: The Guard's Code Review Mode has 5 audit
sections:

1. Intent Drift Detection
2. Constitution Alignment
3. Zero-Waste Mandate
4. Cross-Component Value Preservation [PACK]
5. Gatekeeping Integrity

The "Documentation Completeness" check fits naturally
as item 6 — it's a distinct concern (documentation
accuracy) that doesn't overlap with the existing 5
sections. It complements section 4 (Cross-Component
Value) but is specific enough to warrant its own item.

**Decision**: Add `#### 6. Documentation Completeness`
after section 5 in the Guard's Code Review audit. The
check verifies that AGENTS.md and README.md were updated
when user-facing behavior changed. Severity: MEDIUM
(per spec FR-012).

## R05: Content Convention Pack Applicability

**Question**: Should the Curator reference the content
convention pack?

**Findings**: The `content.md` pack defines standards
for content-producing agents (Scribe, Herald, Envoy).
The Curator does not produce content — it triages and
files issues. However, the Curator's GitHub issue
descriptions should follow content standards for
clarity and consistency.

**Decision**: The Curator SHOULD reference `content.md`
for issue description quality (VB rules for voice,
FA rules for accuracy) but is not bound by TD/BA/PR
sections since it doesn't produce documentation, blog
posts, or PR communications. The Source Documents
section will include `content.md` as optional context.

## R06: Bash Access Precedent Analysis

**Question**: Are there precedents for restricted bash
access in the agent ecosystem?

**Findings**: No existing Divisor agent has `bash: true`.
The Cobalt-Crush developer agent has full tool access
(write, edit, bash) because it implements code. Content
agents (Scribe, Herald, Envoy) have write/edit but no
bash. The constitution-check agent has no bash.

The Curator's bash need is narrow and well-defined:
- `gh issue create --repo unbound-force/website ...`
- `gh issue list --repo unbound-force/website ...`

**Decision**: Document the bash restriction explicitly
in the agent file with a dedicated "Bash Access
Restriction" section. This makes the scope auditable
by the Adversary agent's "Gate Tampering" check.

## R07: Duplicate Issue Prevention

**Question**: How should the Curator prevent filing
duplicate issues?

**Findings**: The `gh issue list` command supports
filtering by label and search:
```bash
gh issue list --repo unbound-force/website \
  --label docs --search "keyword" --state open
```

**Decision**: Before filing any issue, the Curator
MUST search existing open issues in the website repo
using `gh issue list` with the relevant label and
a keyword from the proposed issue title. If a matching
issue exists, reference it in the finding instead of
creating a duplicate. The search should be broad enough
to catch similar issues but not so broad that it
produces false matches.

## R08: User-Facing Change Detection Heuristic

**Question**: How does the Curator determine whether a
change is "user-facing"?

**Findings**: From the spec's Assumptions section:

User-facing file paths:
- `cmd/` — CLI commands and flags
- `.opencode/agents/` — agent capabilities
- `.opencode/command/` — slash commands
- `internal/scaffold/` — scaffold output (affects
  what `uf init` produces)
- `AGENTS.md` — project documentation
- `README.md` — project documentation

Internal-only paths (no documentation needed):
- `internal/` (excluding `scaffold/`) — business logic
- `*_test.go` — test files
- `.github/` — CI/CD configuration
- `specs/` — specification artifacts
- `openspec/` — tactical change artifacts

**Decision**: The Curator uses path-based heuristics to
classify changes. The heuristic is documented in the
agent file so it can be refined over time without code
changes.
