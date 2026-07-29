## Why

The `/finale` command's secrets check gate (Step 2, lines 63-80)
warns the user about potential secret files but relies on
prose-only confirmation: "Ask for confirmation. If the user
declines, stop and let them handle it manually."

This is a T3 weakness (confirmation gate is inline text only).
Under context compression or fast-path reasoning, an agent can
treat this prose as informational and proceed to `git add .`
without explicit user confirmation, potentially staging files
that contain secrets.

The sibling change `askuser-tool-confirmations` (issue #346)
addresses this same T3 pattern in `/review-pr`,
`/address-feedback`, and `/triage-issue`. This change applies
the same fix to `/finale`, which was identified as a separate
issue (#347) because it involves a different command with
distinct safety semantics (staging files vs. posting reviews).

## What Changes

Replace the prose-only confirmation at lines 79-80 of
`.opencode/commands/finale.md` with an explicit AskUserQuestion
tool call that presents structured options. The user must
deliberately select whether to proceed with staging or stop.

## Capabilities

### New Capabilities
- None (no new functionality)

### Modified Capabilities
- `/finale`: The secrets check confirmation in Step 2 is
  converted from prose instructions ("Ask for confirmation")
  to an explicit AskUserQuestion tool call with structured
  options

### Removed Capabilities
- None

## Impact

**Files modified** (2 files -- command + scaffold asset):
- `.opencode/commands/finale.md`
- `internal/scaffold/assets/opencode/commands/finale.md`

**Behavioral change**: The secrets check gate retains its
safety semantics (never stage secret files without user
confirmation). Only the confirmation mechanism changes from
prose text to a structured AskUserQuestion tool call with
predefined options.

**Scaffold drift**: The command file and its scaffold asset
must be kept byte-identical. Both copies must be updated
together. The existing drift detection test
(`TestEmbeddedAssets_MatchSource`) enforces this.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent prompt instructions (a slash command
markdown file), not inter-hero artifact exchange. No artifact
formats, envelopes, or hero interfaces are affected.

### II. Composability First

**Assessment**: N/A

No hero dependencies are introduced or modified. The `/finale`
command remains independently usable. The AskUserQuestion tool
is a built-in OpenCode capability, not a hero dependency.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable outputs or provenance metadata are
affected. The commit, push, and PR creation behaviors remain
unchanged.

### IV. Testability

**Assessment**: PASS

The scaffold drift detection test
(`TestEmbeddedAssets_MatchSource`) will verify that the command
file and its scaffold asset remain synchronized. No new test
infrastructure is needed.

### V. Security by Default

**Assessment**: PASS

This change strengthens a security-relevant confirmation gate
by converting a prose-only instruction (which can be skipped
under context compression) to a structured AskUserQuestion
tool call that requires explicit user selection. This makes
the secrets check gate a structural property of the command
rather than a review-time afterthought, directly supporting
the principle's intent.
<!-- scaffolded by uf vdev -->
