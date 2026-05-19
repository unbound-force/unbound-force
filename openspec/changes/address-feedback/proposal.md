## Why

After a PR is filed via `/finale`, external reviewers
(humans, bots, AI-assisted tools) provide feedback that
the author must address. Today, this is entirely manual:
the author reads each comment, writes an ad-hoc prompt
for the agent, hopes nothing is missed, and pushes fixes
without structured validation. This gap breaks the
otherwise automated lifecycle that `/unleash` and
`/review-council` establish before the PR, and
`/review-pr` provides for one-shot analysis after.

Without standardization:
- Authors may miss feedback items across long threads
- Agent prompts lack project context (constitution,
  convention packs), leading to fixes that deviate
  from guidelines
- There is no structured record of how feedback was
  addressed, preventing learning from review patterns
- The same types of feedback recur across PRs with no
  mechanism to detect or prevent this
- Multiple review rounds require re-reading the entire
  PR conversation each time

## What Changes

Introduce an `/address-feedback` slash command that
fetches PR review feedback from GitHub, assesses each
item against project standards, walks the author through
a structured triage, and executes the resulting actions
(code fixes, reply comments) as a batch. Introduce a
`feedback-triage` schema to capture triage decisions as
a structured artifact for consumption by other heroes.

## Capabilities

### New Capabilities

- `address-feedback-command`: Slash command that ingests
  PR review feedback from GitHub, classifies each item
  as data-driven or subjective, assesses against project
  conventions and constitution, presents items one-by-one
  to the author for decision (accept, modify, reject,
  ask), then executes all decisions as a batch (code
  changes committed per-fix, reply comments posted to
  the PR). Runs `/review-council` on code changes before
  pushing.
- `tiered-assessment`: Two-tier assessment engine. Tier 1
  handles simple items (formatting, naming, clear
  convention matches) directly. Tier 2 escalates complex
  items (security, architecture, multi-file, ambiguous)
  to the relevant Divisor agent for specialized analysis.
- `feedback-triage-schema`: JSON Schema for the
  feedback-triage artifact, capturing per-item
  classification, evidence, recommendation, author
  decision, and reasoning. Includes round tracking for
  multi-round review cycles.
- `pr-scoped-cache`: Local cache under `.uf/feedback/`
  keyed by PR number, accelerating re-assessment on
  subsequent rounds. Cache is disposable -- the command
  reconstructs from GitHub if cache is missing.
- `authority-matrix`: Reviewer role detection (maintainer,
  collaborator, member, bot) with priority weighting.
  Data-driven findings from maintainers are MUST-fix.
  Subjective suggestions follow the author-decides model
  with priority influenced by reviewer authority.

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

### Files Created
- `.opencode/commands/address-feedback.md` -- slash command
- `internal/scaffold/assets/opencode/commands/address-feedback.md`
  -- scaffold mirror
- `schemas/feedback-triage/v1.0.0.schema.json` -- schema
- `schemas/feedback-triage/samples/sample-feedback-triage.json`
  -- sample artifact
- `schemas/feedback-triage/README.md` -- schema docs

### Files Modified
- `AGENTS.md` -- add `/address-feedback` to PR Review
  Commands table, add `feedback-triage` to schema
  registry if documented
- `CHANGELOG.md` -- add entry for new command and
  schema
- `.gitignore` -- add `.uf/feedback/` cache directory
- `docs/usage.md` -- add feedback workflow section
- `docs/architecture.md` -- reference new command and
  schema

### Related Issues
- #207 (Close learning feedback loop): feedback-triage
  artifact provides a new learning signal source. Accept
  and reject patterns across PRs reveal what the pre-PR
  review consistently misses.
- #177 (Enrich /review-council with /review-pr
  capabilities): `/address-feedback` is the inbound
  counterpart to #177's outbound GitHub posting. Together
  they close the full review conversation loop.
- #176 (Create review-context convention pack):
  `/address-feedback` should consume the review-context
  pack for spec discovery, issue linking, and path
  classification when available. Soft dependency -- works
  without it by inlining context discovery.
- #175 (Extract shared pre-flight skill): consumed
  transitively via `/review-council` in Phase 4.
- #142 (Improve PR description in finale): feedback-triage
  artifact data could enable a "Review Feedback Status"
  section in PR descriptions.
- #199 (Measurement baseline): accept/reject ratios and
  escalation frequency from feedback-triage artifacts
  provide measurement data for harness effectiveness.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The command produces a self-describing artifact
(feedback-triage) with full provenance metadata
(producer, version, timestamp, PR context). Other
heroes (Mx F, Muti-Mind, Cobalt-Crush) can consume
this artifact without synchronous interaction. The
artifact envelope format follows the Hero Interface
Contract. The command itself consumes GitHub API data
asynchronously -- it does not require any other hero
to be present or running.

### II. Composability First

**Assessment**: PASS

The command is independently useful without any other
hero deployed. It requires only the `gh` CLI and access
to convention packs (which are local files). Tier 2
escalation to Divisor agents is additive -- if no
Divisor agents are present, all items are assessed at
Tier 1. The `review-context` pack (#176) is a soft
dependency -- the command inlines context discovery
when the pack is absent.

### III. Observable Quality

**Assessment**: PASS

The feedback-triage artifact is machine-parseable JSON
with a registered schema. It includes provenance
(producer identity, version, timestamp, PR reference),
per-item classification with evidence references, and
aggregate statistics. The schema will be versioned in
`schemas/feedback-triage/` alongside existing schemas.
Human-readable terminal output accompanies the JSON
artifact.

### IV. Testability

**Assessment**: PASS

The command is a slash command (agent instructions),
not compiled Go code. The Go-testable surface consists
of: (1) JSON Schema validation -- positive validation
of sample artifacts against the schema, negative
validation of invalid inputs (missing required fields,
out-of-enum values, extra properties); (2) scaffold
embedding tests -- `TestAssetPaths_MatchExpected`
verifies the new command asset is in the expected list,
`TestEmbeddedAssets_MatchSource` ensures drift
detection between live copy and scaffold asset;
(3) the `TestRun_CreatesFiles` test verifies the file
count includes the new asset. Coverage strategy is
defined in the spec's Test Strategy section.
<!-- scaffolded by uf vdev -->
