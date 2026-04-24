## Why

AGENTS.md is the API contract between humans and AI agents -- it is the single
document where agents learn project context, build commands, coding conventions,
behavioral constraints, and governance rules. Every agent in the uf ecosystem
reads it at session start, and cross-tool bridge files (CLAUDE.md, .cursorrules)
route other tools to it.

Despite this critical role, `uf` has a **bootstrap gap**: nothing creates
AGENTS.md from scratch. The scaffold engine (`uf init`) only appends a
Convention Packs section to an existing file and skips entirely if the file is
absent. The `/uf-init` command injects 8 guidance blocks but also skips if
the file is missing. The `uf doctor` check only verifies file existence and
its install hint ("Run: uf init") is misleading because `uf init` will not
create the file.

This leaves new adopters stuck: they must manually author AGENTS.md before any
uf tooling can enhance it. There is no guidance on what sections to include,
no validation of content quality, and no detection of stale or missing sections.

## What Changes

Introduce the `/agent-brief` slash command and enhanced `uf doctor` checks to
close the AGENTS.md bootstrap gap:

1. **New `/agent-brief` OpenCode command** -- an LLM-driven command that
   auto-detects mode (create vs audit) and handles AGENTS.md lifecycle:
   - **Create mode**: Analyzes the project (language, build system, CI config,
     directory structure, spec framework presence) and generates AGENTS.md with
     Tier 1 sections filled from analysis, Tier 1C sections for detected
     governance/spec artifacts, and Tier 2 sections stubbed with guidance.
   - **Audit mode**: Scores existing AGENTS.md against a context-sensitive
     section taxonomy, reports quality metrics (line count, section coverage,
     code block presence), detects staleness and missing cross-framework
     governance references, and offers improvement suggestions.
   - **Bridge file handling**: Ensures CLAUDE.md and .cursorrules exist with
     proper AGENTS.md references after create or improve operations.

2. **New "Agent Context" doctor check group** -- replaces the single existence
   check with a comprehensive 10-12 check group covering section presence
   (Tier 1 headers), build command code blocks, line count, context-sensitive
   governance checks (constitution reference when `.specify/` exists, spec
   framework description when `specs/` or `openspec/` exist), and bridge file
   validation.

3. **Scaffold asset distribution** -- the command is embedded as a scaffold
   asset so every project that runs `uf init` receives `/agent-brief`.

## Capabilities

### New Capabilities
- `/agent-brief`: OpenCode slash command for AGENTS.md creation, auditing,
  and improvement with context-sensitive section taxonomy
- `/agent-brief create`: Force create mode (generates AGENTS.md from project
  analysis even if file exists)
- `/agent-brief audit`: Force audit mode (read-only quality assessment)
- `uf doctor` "Agent Context" check group: Structural validation of AGENTS.md
  with 10-12 deterministic checks including spec-framework awareness

### Modified Capabilities
- `uf doctor`: Existing AGENTS.md existence check in `checkScaffoldedFiles()`
  moves to dedicated `checkAgentContext()` group with deeper validation. Install
  hint changes from "Run: uf init" to "Run: /agent-brief in OpenCode"
- `uf init`: Now deploys the `/agent-brief` command as a scaffold asset

### Removed Capabilities
- None

## Impact

### Files Created
- `.opencode/command/agent-brief.md` -- the slash command (~500 lines)
- `internal/scaffold/assets/opencode/command/agent-brief.md` -- scaffold
  asset copy

### Files Modified
- `internal/doctor/checks.go` -- new `checkAgentContext()` function, remove
  AGENTS.md check from `checkScaffoldedFiles()`
- `internal/doctor/checks_test.go` -- tests for new check group
- `internal/scaffold/scaffold_test.go` -- updated `expectedAssetPaths`,
  new asset count

### Systems Affected
- `uf doctor` output gains a new "Agent Context" check group
- `uf init` deploys one additional scaffold asset
- Projects without AGENTS.md get a clear creation path

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The `/agent-brief` command produces a self-describing artifact (AGENTS.md)
that all agents consume independently without runtime coupling. The command
itself operates autonomously -- it reads project files, generates content,
and writes output without requiring synchronous interaction with any agent.
The doctor checks produce machine-readable results (JSON format via
`--format=json`) following the established `CheckGroup`/`CheckResult`
artifact pattern.

### II. Composability First

**Assessment**: PASS

The command is independently useful: any project can benefit from AGENTS.md
creation/auditing regardless of whether other uf heroes (Gaze, Divisor,
Mx F) are deployed. The doctor checks follow the existing injectable
`Options` pattern for standalone testability. The command does not introduce
mandatory dependencies on any other hero or tool. When spec-driven artifacts
are absent, governance sections are simply skipped rather than failing.

### III. Observable Quality

**Assessment**: PASS

The audit mode produces a structured quality report with per-section status,
tier completion metrics, and an overall effectiveness label (Excellent /
Strong / Adequate / Weak / Missing). The doctor check group produces
`CheckResult` items with `Pass`/`Warn`/`Fail` severities and actionable
`InstallHint` strings, consumable in both human-readable and JSON formats.
Quality claims (section coverage, line count) are backed by deterministic
checks reproducible across runs.

### IV. Testability

**Assessment**: PASS

The doctor check group follows the established injectable function pattern
(`opts.ReadFile`, `opts.LookPath`) for isolation testing without filesystem
access. Each check in the group is independently verifiable. The slash
command is an LLM-driven Markdown file (tested via scaffold drift detection
and asset presence tests). Coverage strategy: unit tests for all doctor
check functions (section detection regex, line counting, bridge file
validation), scaffold tests for asset registration and drift detection.
