## Context

The doctor text output in `internal/doctor/format.go`
uses minimal ASCII indicators and muted ANSI colors.
The gcal-organizer project's doctor command demonstrates
a more effective pattern with emoji indicators, lipgloss
styling, and boxed summaries.

All changes are isolated to `format.go` and its tests.
No check logic, models, or JSON output changes.

## Goals / Non-Goals

### Goals
- Replace ASCII indicators with emoji indicators
- Use brighter, more distinctive lipgloss colors
- Add styled title with emoji prefix
- Add boxed summary with counters
- Style fix hints in subtle gray
- Add contextual completion message
- Preserve `NO_COLOR` / pipe detection plain-text
  fallback

### Non-Goals
- Changing check logic or severity assignments
- Changing JSON output format
- Adding spinners, progress bars, or animation
- Adding interactive prompts
- Changing doctor.go, checks.go, models.go, or
  environ.go
- Flattening the group structure (keep groups)

## Decisions

### Lipgloss Style Definitions

```go
// Title: bold pink (matches gcal-organizer pattern)
titleStyle = renderer.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("212"))

// Pass: bright green
passStyle = renderer.NewStyle().
    Foreground(lipgloss.Color("10"))

// Warn: bright yellow
warnStyle = renderer.NewStyle().
    Foreground(lipgloss.Color("11"))

// Fail: bright red
failStyle = renderer.NewStyle().
    Foreground(lipgloss.Color("9"))

// Subtle: dim gray (fix hints, details)
dimStyle = renderer.NewStyle().
    Foreground(lipgloss.Color("241"))

// Box: rounded border, purple accent
boxStyle = renderer.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    Padding(0, 1)
```

### Indicator Mapping

| State | Colored | Plain Text |
|-------|---------|------------|
| Pass | `✅` (green) | `[PASS]` |
| Warn | `⚠️ ` (yellow) | `[WARN]` |
| Fail | `❌` (red) | `[FAIL]` |
| Info (optional absent) | `⊘` (gray) | `[INFO]` |

### Output Format

Title:
```
🩺 Unbound Force Doctor
```

Group header (bold, no emoji):
```
Core Tools
```

Check line (2-space indent + emoji + name + message):
```
  ✅ opencode           installed
  ❌ dewey              not found
     Fix: brew install unbound-force/tap/dewey
```

Fix hints (5-space indent + `Fix:` in gray):
```
     Fix: brew install unbound-force/tap/dewey
```

Boxed summary:
```
╭─────────────────────────────────────────────╮
│  ✅ 12 passed  ⚠️  2 warnings  ❌ 1 failed │
╰─────────────────────────────────────────────╯
```

Contextual message after box:
- All pass: green `🎉 Everything looks good!`
- Failures: gray `Run 'uf doctor' after fixes.`
- Warnings only: gray `All critical checks passed.`

## Risks / Trade-offs

**Emoji rendering**: Some terminals or CI environments
may not render emoji well. The `NO_COLOR` / pipe
detection fallback uses plain `[PASS]`/`[FAIL]` text,
avoiding emoji entirely when color is disabled.

**Box width**: The boxed summary width depends on the
counter values. Lipgloss handles this automatically
with the `Padding` setting.

**Test assertions**: Existing format tests assert
specific indicator characters (`✓`, `✗`). These will
need updating to the new emoji characters. The test
structure (checking for indicator presence in output)
remains the same.
