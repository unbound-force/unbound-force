## Why

The `/uf-init` slash command applies project-specific customizations
to third-party tool files (OpenSpec skills and speckit commands)
after `uf init` runs. Cross-project comparison between
`unbound-force` and `gaze` reveals that `/uf-init` has gaps:
customizations that exist in the meta repo are not reliably
reproduced in downstream repos.

**Problems identified:**

1. **Missing customization steps**: The STOP HERE blocks (preventing
   premature implementation) and review-rationale guardrail
   sentences are present in unbound-force's speckit commands but
   are not defined in any `/uf-init` step. They were likely added
   manually, meaning downstream repos never receive them.

2. **Incomplete branch enforcement**: The dirty working tree check
   (propose), commit-before-archive flow (archive-change), and
   branch-switch confirmation (explore) are missing from gaze,
   suggesting the idempotency checks in Step 2 are too coarse --
   detecting the basic branch check and skipping the enhanced
   variants.

3. **Scaffold comment accumulation**: Repeated `uf init` runs
   across versions leave duplicate `<!-- scaffolded by uf ... -->`
   comments (gaze has 5 per file vs 1 in unbound-force). No
   deduplication step exists.

4. **Legacy directory cleanup**: The old `unbound/packs/` and
   `command/` (singular) directories persist in downstream repos.
   Step 0 handles `command/` migration but may not trigger
   reliably; `unbound/packs/` is not addressed at all.

**Related issues**: #161 (new commands not scaffolded on upgrade),
#162 (skip logic for user-owned files), #201 (tiered scaffolding),
#213 (uf init does not scaffold usable constitution).

## What Changes

Modify `.opencode/commands/uf-init.md` to add missing
customization steps, tighten idempotency checks, and add a
cleanup step for legacy artifacts.

## Capabilities

### New Capabilities
- `STOP HERE injection`: New step injecting STOP HERE blocks
  into spec-phase speckit commands (specify, plan, tasks,
  analyze, checklist, clarify) to prevent premature
  implementation
- `Review-rationale guardrail`: Extends Guardrails injection
  to include the review-rationale sentence in spec-phase
  commands
- `Scaffold comment dedup`: Cleanup step that deduplicates
  `<!-- scaffolded by uf ... -->` comments, keeping only the
  most recent one
- `Legacy directory cleanup`: New step removing `unbound/packs/`
  after verifying `uf/packs/` exists, and improving `command/`
  migration reliability

### Modified Capabilities
- `Branch enforcement (Step 2)`: Tightened idempotency checks
  to distinguish between basic branch checks and enhanced
  variants (dirty tree, commit-before-archive,
  branch-switch confirmation)
- `Guardrails injection (Step 6)`: Split into spec-phase and
  execution-phase variants so the review-rationale sentence
  is only added where appropriate

### Removed Capabilities
- None

## Impact

- **Files modified**: `.opencode/commands/uf-init.md` (sole
  target)
- **Downstream effect**: All repos running `/uf-init` will
  receive the full set of customizations, producing parity
  with the meta repo
- **Risk**: Low. All insertions are idempotent and the command
  instructs users to run `git diff` after completion. No Go
  source or test files are modified.
- **Related files**: The 7 OpenSpec skill/command files and
  9 speckit command files that `/uf-init` targets are not
  directly modified by this change -- only the command that
  modifies them is updated

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change improves artifact consistency across repos by
ensuring `/uf-init` produces identical customizations
regardless of which repo it runs in. All customizations
are file-based (Markdown insertions) with no runtime
coupling between heroes.

### II. Composability First

**Assessment**: PASS

`/uf-init` already handles missing files gracefully (skip
and report). The new steps follow the same pattern: check
for presence, skip if present, insert if absent. No new
mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

The command produces a structured summary report with status
indicators for every file processed. The new steps follow
the existing reporting pattern (checkmark/skip/error per
file).

### IV. Testability

**Assessment**: PASS

Verification is through scaffold asset drift detection
tests (`TestEmbeddedAssets_MatchSource`) and cross-repo
comparison after running the command. The spec scenarios
define observable outcomes that serve as acceptance
criteria.

### V. Security by Default

**Assessment**: N/A

This change modifies a slash command (Markdown
instructions), not Go source code. No supply chain,
input validation, or privilege changes are introduced.
The commit-before-archive step uses explicit file staging
(not `git add -A`) to avoid staging unintended files.
