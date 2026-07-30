## Why

The `/finale` command's Step 4 ("Push to Remote") at lines
156-165 executes `git push` immediately after a divergence
check with no AskUserQuestion gate before the push command.

Pushing to a remote is an irreversible external action. Per
the policy established by issue #346 (root cause analysis of
irreversible actions without confirmation gates): every
irreversible external action requires a mandatory
AskUserQuestion gate immediately before execution.

This is the same T1 weakness pattern (irreversible external
action without mandatory AskUserQuestion) that was fixed in
`/review-pr`, `/address-feedback`, and `/triage-issue` by the
`askuser-tool-confirmations` change. The `/finale` command's
push step was not included in that change.

## What Changes

Add an AskUserQuestion confirmation gate immediately before the
`git push` execution in Step 4 of `/finale`. The gate presents
the user with a clear choice to proceed or abort before the
irreversible push occurs.

## Capabilities

### New Capabilities
- None (no new functionality)

### Modified Capabilities
- `/finale` Step 4: Push to Remote now requires explicit user
  confirmation via AskUserQuestion before executing `git push`.
  The confirmation shows the target remote and branch.

### Removed Capabilities
- None

## Impact

**Files modified** (1 command file + 1 scaffold asset):
- `.opencode/commands/finale.md`
- `internal/scaffold/assets/opencode/commands/finale.md`

**Behavioral change**: The push step now pauses for user
confirmation before executing. The safety semantics are
strengthened -- a user can no longer accidentally push to a
remote without explicit opt-in. All other `/finale` steps
remain unchanged.

**Scaffold drift**: The command file has a byte-identical
scaffold asset that must be kept synchronized. Both copies
must be updated together.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent prompt instructions (a slash
command markdown file), not inter-hero artifact exchange.
No artifact formats, envelopes, or hero interfaces are
affected.

### II. Composability First

**Assessment**: N/A

No hero dependencies are introduced or modified. The
`/finale` command remains independently usable. The
AskUserQuestion tool is a built-in OpenCode capability,
not a hero dependency.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable outputs or provenance metadata are
affected. The commit messages, PR bodies, and CI check
outputs remain unchanged.

### IV. Testability

**Assessment**: PASS

The scaffold drift detection tests
(`TestEmbeddedAssets_MatchSource`) will verify that the
command file and its scaffold asset remain synchronized.
No new test infrastructure is needed.

### V. Security by Default

**Assessment**: PASS

This change strengthens security posture by adding a
mandatory confirmation gate before an irreversible
external action (`git push`). No supply chain, input
validation, or privilege changes are involved. The
AskUserQuestion gate is a UX confirmation that prevents
accidental irreversible actions, aligning with the
principle's safety-by-design intent.
<!-- scaffolded by uf vdev -->
