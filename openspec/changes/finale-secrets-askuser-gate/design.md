## Context

The `/finale` command (`.opencode/commands/finale.md`)
handles branch finalization: staging, committing, pushing,
and PR creation. Step 2 includes a secrets check that scans
for files matching common secret patterns (.env, *.key,
*.pem, credentials.json) before running `git add .`.

The current confirmation gate at lines 79-81 is prose-only:
"Ask for confirmation. If the user declines, stop and let
them handle it manually." This is a T3 weakness -- the same
pattern fixed in `/review-pr`, `/address-feedback`, and
`/triage-issue` by the `askuser-tool-confirmations` change.

The proposal's constitution alignment confirmed this change
is N/A for Principles I-III and PASS for Principles IV-V.

## Goals / Non-Goals

### Goals
- Replace the prose confirmation at lines 79-81 with an
  explicit AskUserQuestion tool call with structured options
- Match the established convention from the
  `askuser-tool-confirmations` change (bold formatting,
  action-descriptive option labels)
- Sync the scaffold asset copy to maintain byte-identity

### Non-Goals
- Scanning for additional secret file patterns beyond what
  is already defined
- Modifying other confirmation points in `/finale` (e.g.,
  commit message approval at line 149)
- Changing Go source code logic
- Modifying the AskUserQuestion tool itself

## Decisions

### D1: Use structured selection with action-descriptive labels

Replace the prose "Ask for confirmation" with:

```
Use the **AskUserQuestion tool** with options:
  ["Yes -- stage all files including flagged ones",
   "No -- stop and let me handle it manually"]

If the user selects "No", STOP. Do not run `git add .`.
```

**Rationale**: Matches the convention established by D2
in the `askuser-tool-confirmations` design -- option
labels include action context, not bare yes/no. The
"including flagged ones" phrasing makes the consequence
explicit.

### D2: Bold formatting convention preserved

The AskUserQuestion tool is referenced as
`**AskUserQuestion tool**` (bold, PascalCase), matching
the convention in `opsx-propose.md`, `opsx-apply.md`,
`review-pr.md`, `address-feedback.md`, and
`triage-issue.md`.

**Rationale**: Consistency with established command
formatting.

### D3: Scaffold asset updated in lockstep

The command file `.opencode/commands/finale.md` has a
byte-identical copy at
`internal/scaffold/assets/opencode/commands/finale.md`.
Both must be updated together. The drift detection test
(`TestEmbeddedAssets_MatchSource`) enforces this.

**Rationale**: Required by the scaffold pattern. Existing
test infrastructure validates automatically.

### D4: Minimal diff -- replace only the prose gate

Only lines 79-81 change. The surrounding warning message
text (lines 70-78), the secrets pattern list (lines 64-68),
and the `git add .` command (line 83) remain unchanged.

**Rationale**: Smallest possible change reduces risk and
keeps the review focused.

## Risks / Trade-offs

### R1: No behavioral regression

This change is purely instructional (markdown edits). No
Go code, API, or schema changes. The scaffold drift test
provides automated verification. Risk is minimal.

### R2: Single confirmation point

Unlike the `askuser-tool-confirmations` change (15
interaction points across 3 commands), this change
addresses exactly 1 interaction point in 1 command. The
scope is intentionally narrow.
<!-- scaffolded by uf vdev -->
