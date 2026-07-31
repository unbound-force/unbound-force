## Why

The `/finale` command's secrets check gate (lines 63-81 in
`finale.md`) warns users about potential secret files before
staging, but the confirmation is prose-only: "Ask for
confirmation. If the user declines, stop and let them handle
it manually." No `AskUserQuestion` tool call is specified.

Under context compression or fast-path reasoning, an agent
can treat this prose as informational and proceed to `git
add .` without explicit user confirmation, potentially
staging files that contain secrets (.env, credentials.json,
*.key, *.pem).

This is a T3 weakness (confirmation gate is inline text
only, not enforced by a tool call) -- the same pattern
addressed in `/review-pr`, `/address-feedback`, and
`/triage-issue` by the `askuser-tool-confirmations` change.

Fixes: https://github.com/unbound-force/unbound-force/issues/347

## What Changes

Replace the prose confirmation at lines 79-81 of
`.opencode/commands/finale.md` with an explicit
**AskUserQuestion tool** call with structured options.
Sync the identical change to the scaffold asset copy.

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `/finale` secrets check: Prose confirmation replaced
  with explicit AskUserQuestion tool call, ensuring the
  gate cannot be skipped under context compression

### Removed Capabilities
- None

## Impact

- `.opencode/commands/finale.md` -- secrets check gate
  section (lines ~63-81)
- `internal/scaffold/assets/opencode/commands/finale.md`
  -- byte-identical scaffold copy
- No Go source code changes
- No schema, API, or artifact format changes

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instruction text only. No
artifact formats, inter-hero communication, or envelope
structures are affected.

### II. Composability First

**Assessment**: N/A

No new dependencies introduced. The `/finale` command
remains independently usable. The AskUserQuestion tool
is already available in the agent runtime.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable outputs, provenance metadata, or
scoring thresholds are modified. This is purely an
instructional text change.

### IV. Testability

**Assessment**: PASS

The scaffold drift test
(`TestEmbeddedAssets_MatchSource`) validates that command
files and scaffold assets remain byte-identical. This
existing test infrastructure covers the change without
requiring new tests.

### V. Security by Default

**Assessment**: PASS

This change directly improves security posture by
hardening the secrets confirmation gate from a
skippable prose instruction to an enforceable tool
call. It prevents agents from bypassing the user
confirmation and staging secret files.
<!-- scaffolded by uf vdev -->
