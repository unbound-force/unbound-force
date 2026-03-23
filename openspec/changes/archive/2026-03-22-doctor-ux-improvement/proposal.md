## Why

The `uf doctor` output currently uses minimal ASCII
indicators (`✓`, `✗`, `!`, `○`) with muted ANSI colors
(colors 1-3, 8). While functional, the output is hard
to scan quickly and lacks the visual polish expected of
a developer tool. The gcal-organizer project demonstrates
a more effective UX pattern using emoji status indicators,
lipgloss-styled color, and boxed summaries that make
diagnostic output immediately scannable.

## What Changes

Upgrade the doctor text output formatting in
`internal/doctor/format.go` to use:

- **Emoji indicators** instead of ASCII symbols:
  `✅` pass, `⚠️` warn, `❌` fail, `⊘` skip/info
- **Richer lipgloss colors**: bright green (10), bright
  yellow (11), bright red (9), dim gray (241), bold
  pink (212) for titles
- **Styled title** with emoji: `🩺 Unbound Force Doctor`
- **Boxed summary** using lipgloss rounded border with
  purple accent (color 63)
- **Styled fix hints** in gray with `Fix:` prefix
- **Contextual completion message**: `🎉 Everything
  looks good!` when all checks pass

## Capabilities

### New Capabilities
- `emoji-status-indicators`: Doctor output uses emoji
  for instant visual scanning of check states
- `boxed-summary`: Summary displayed in a bordered box
  with colored counters
- `styled-fix-hints`: Fix hints rendered in subtle gray
  with consistent indentation

### Modified Capabilities
- `doctor-text-output`: Visual formatting changes only.
  Same data, same check logic, same JSON output. Only
  the text rendering changes.

### Removed Capabilities
- None

## Impact

- `internal/doctor/format.go` -- all formatting changes
  are isolated to this single file
- `internal/doctor/format_test.go` -- test assertions
  for new emoji indicators and formatting
- No changes to check logic, JSON output, doctor.go,
  checks.go, models.go, or environ.go

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

No artifact communication changes. This is a
presentation-layer change only.

### II. Composability First

**Assessment**: N/A

No dependency changes. The doctor command continues
to work independently.

### III. Observable Quality

**Assessment**: PASS

The doctor output remains machine-parseable (JSON format
is unchanged). The text output is enhanced for human
readability while maintaining the same data structure.
The `NO_COLOR` / pipe detection fallback to plain text
indicators is preserved.

### IV. Testability

**Assessment**: PASS

All formatting changes are in `format.go` which is
tested via `format_test.go`. Tests will be updated to
assert the new emoji indicators. The `hasColor` branching
ensures plain-text fallback is still testable.
