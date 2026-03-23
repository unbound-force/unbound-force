## ADDED Requirements

### Requirement: emoji-status-indicators

The doctor text output MUST use emoji indicators for
check states: `✅` for pass, `⚠️` for warn, `❌` for
fail, `⊘` for info/skip.

#### Scenario: pass check shows green checkmark

- **GIVEN** a check result with severity Pass
- **WHEN** the doctor output is rendered in a color
  terminal
- **THEN** the indicator is `✅` rendered in bright
  green (ANSI color 10)

#### Scenario: fail check shows red cross

- **GIVEN** a check result with severity Fail
- **WHEN** the doctor output is rendered in a color
  terminal
- **THEN** the indicator is `❌` rendered in bright
  red (ANSI color 9)

### Requirement: styled-title

The doctor text output MUST display a styled title
with emoji: `🩺 Unbound Force Doctor` in bold pink
(ANSI color 212).

#### Scenario: title renders with stethoscope emoji

- **GIVEN** the doctor command is run
- **WHEN** the text output is rendered
- **THEN** the first line contains `🩺` and
  `Unbound Force Doctor` in bold

### Requirement: boxed-summary

The doctor text output MUST display the summary in a
bordered box using lipgloss rounded borders with purple
accent (ANSI color 63).

#### Scenario: summary shows boxed counters

- **GIVEN** a completed doctor run with 10 pass, 2
  warn, 1 fail
- **WHEN** the summary is rendered
- **THEN** the output contains `✅ 10 passed`,
  `⚠️  2 warnings`, and `❌ 1 failed` inside a
  bordered box

### Requirement: contextual-completion

The doctor text output MUST display a contextual
message after the summary box: `🎉 Everything looks
good!` when all checks pass.

#### Scenario: all checks pass

- **GIVEN** all doctor checks pass with zero warns and
  fails
- **WHEN** the summary is rendered
- **THEN** the output contains `🎉 Everything looks
  good!` in green

### Requirement: styled-fix-hints

Fix hints MUST be rendered in dim gray (ANSI color 241)
with 5-space indentation and `Fix:` prefix.

#### Scenario: fix hint renders in gray

- **GIVEN** a check result with an InstallHint
- **WHEN** the output is rendered
- **THEN** the hint appears as `     Fix: <hint>` in
  gray text

## MODIFIED Requirements

### Requirement: plain-text-fallback

The `NO_COLOR` / pipe detection plain-text fallback
MUST continue to use bracket indicators (`[PASS]`,
`[WARN]`, `[FAIL]`, `[INFO]`) instead of emoji.

Previously: Used `[PASS]`, `[WARN]`, `[FAIL]`, `[INFO]`
with ASCII `✓`, `✗`, `!`, `○` for colored mode.

Now: Colored mode uses emoji; plain text mode unchanged.

## REMOVED Requirements

None.
