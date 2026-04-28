## Why

The org-infra repository has a well-designed `/review_pr`
command that reviews GitHub PRs with CI causality analysis,
token-efficient diff handling, spec awareness, fix-branch
offers for pre-existing failures, and in-line comment
posting with human confirmation gates. Currently it is
hand-maintained in org-infra only.

This creates two problems:

1. **No portability.** Other repositories that use `uf init`
   do not get a PR review command. Each repo would need to
   copy and adapt the command independently, leading to
   drift across repos.
2. **No update path.** Improvements to the command require
   manual propagation. The `uf init` scaffold engine
   solves this — tool-owned command files are automatically
   updated on re-run.

Meanwhile, the existing `/review-council` command reviews
the local working tree pre-PR using the Divisor multi-agent
council. These commands are complementary:

```
  /review-council  →  Pre-PR (local tree, multi-agent)
  /review-pr       →  Post-PR (GitHub PR, single-agent)
```

Adding `/review-pr` to the scaffold fills the post-PR gap
in the UF toolchain.

## What Changes

Add a standardized `review-pr.md` command to the `uf init`
scaffold, adapted from org-infra's `review_pr.md` with the
following modifications:

- **Rename** from `review_pr` (underscore) to `review-pr`
  (kebab-case) per UF naming convention
- **PR number optional** — auto-detect from current branch
  when no argument is given
- **Generic constitution reference** — evaluate against
  whatever principles the project's constitution defines,
  not hardcoded org-infra principles I-VII
- **Convention pack awareness** — load `.opencode/uf/packs/`
  when available for more specific code quality findings,
  graceful degradation when packs are absent
- **Severity pack reference** — use `severity.md` pack
  definitions when present, inline fallback otherwise

All original operational capabilities are preserved:
CI causality analysis, local tool detection, scoped diff
fetching, spec awareness (both `specs/` and `openspec/`),
fix-branch creation for pre-existing failures, and in-line
PR comment posting with human confirmation.

## Capabilities

### New Capabilities

- `/review-pr [N]` command deployed to all repos via
  `uf init` as a tool-owned scaffold asset
- Auto-detection of PR from current branch when no number
  is provided
- Convention pack integration for repos with
  `.opencode/uf/packs/` (additive to constitution checks)
- Severity pack integration for consistent severity
  definitions across the UF ecosystem

### Modified Capabilities

- Scaffold asset count increases by 1 file
- AGENTS.md updated with `/review-pr` documentation and
  relationship to `/review-council`

### Removed Capabilities

None. This is purely additive.

## Impact

**Files created:**
- `internal/scaffold/assets/opencode/command/review-pr.md`
- `.opencode/command/review-pr.md` (canonical source /
  live copy)

**Files modified:**
- `internal/scaffold/scaffold_test.go` — add to
  `expectedAssetPaths`, update file count assertions
- `AGENTS.md` — add command documentation

**External dependencies:**
- `gh` CLI (GitHub CLI) must be installed and authenticated
  for the command to function. This is already a
  recommended tool in the UF ecosystem.

## Constitution Alignment

Assessed against the Unbound Force org constitution
(`.specify/memory/constitution.md` v1.1.0).

### I. Autonomous Collaboration

**Assessment**: N/A

This change adds a developer-facing command file. It does
not affect inter-hero artifact communication, envelope
formats, or hero-to-hero data exchange. The command
produces terminal output and optional GitHub PR comments —
not hero artifacts.

### II. Composability First

**Assessment**: PASS

The command is independently usable without any other UF
hero or tool being present. It detects available tools
(linters, test runners, convention packs, severity pack)
and uses them when found, but functions fully without
them. It does not modify or depend on `/review-council`.

### III. Observable Quality

**Assessment**: PASS

The command produces structured output with severity
levels, categorized findings, and a clear verdict
(APPROVE / REQUEST CHANGES / COMMENT). CI status is
presented in a machine-readable table with causality
classification. When convention packs are available,
findings reference specific numbered rules (CS-001,
SC-002, etc.).

### IV. Testability

**Assessment**: PASS

This change produces no new Go logic — it adds a Markdown
command file and updates test assertions. The scaffold
engine's existing test infrastructure (drift detection,
asset path validation, file count assertions) provides
automated verification. The command itself is a prompt
file, not executable code, so traditional unit testing
does not apply. Verification is handled through the
scaffold test suite and manual invocation.
