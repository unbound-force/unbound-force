## Why

Issue #256: `/uf.init` Step 6 injects identical guardrails into
all 9 speckit command files. The guardrails are correct for the
6 spec-authoring commands but factually wrong for the 3
execution/utility commands:

| Command | Guardrail Says | Command Actually Does |
|---------|---------------|----------------------|
| `speckit.implement.md` | "NEVER modify source code" | Entire purpose is writing source code |
| `speckit.constitution.md` | May only write to `FEATURE_SPEC`/`FEATURE_DIR` | Writes to `.specify/memory/` and `.specify/templates/` |
| `speckit.taskstoissues.md` | May only write to `FEATURE_SPEC`/`FEATURE_DIR` | Creates GitHub issues via MCP API; writes no local files |

The current workaround (Note at lines 627–632) acknowledges the
`implement` contradiction but relies on "instructions override
guardrails" — an anti-pattern that trains agents to distrust
guardrails generally ("guardrail habituation"). The `constitution`
and `taskstoissues` contradictions are unacknowledged.

**Related issues**: #257 (strategic preset migration), #258
(upstream proposal). This fix is tactical and independent of both.

## What Changes

Modify Step 6 of `uf.init.md` to use command-specific guardrail
blocks for the 3 execution/utility commands instead of a single
shared template. Remove the workaround Note (lines 627–632).
Update both copies in lockstep:

1. `.opencode/commands/uf.init.md` (deployed — affects current users)
2. `internal/scaffold/assets/opencode/commands/uf.init.md` (canonical — affects new scaffolds)

Add idempotency logic to detect and replace incorrect guardrails
in existing command files (not just inject when absent).

## Capabilities

### New Capabilities
- `Command-specific guardrail variants`: Step 6 produces
  3 distinct guardrail blocks tailored to each execution/utility
  command's actual operational model

### Modified Capabilities
- `Guardrails injection (Step 6)`: Extended from 2 variants
  (spec-phase, execution/utility) to 5 variants (spec-phase,
  implement, constitution, taskstoissues, and the shared
  execution/utility variant is removed)
- `Guardrail idempotency check`: Enhanced to detect and
  replace incorrect guardrails, not just skip when a
  `## Guardrails` heading exists

### Removed Capabilities
- `Workaround Note (lines 627–632)`: The Note documenting the
  implement override conflict is removed — the fix eliminates
  the conflict it worked around

## Impact

- **Files modified**: `.opencode/commands/uf.init.md` and
  `internal/scaffold/assets/opencode/commands/uf.init.md`
  (both must stay byte-identical per scaffold engine contract)
- **Downstream effect**: All repos running `/uf.init` will
  receive correct, command-specific guardrails for the 3
  execution/utility commands
- **Risk**: Low. Guardrail content changes only — no Go source,
  no test files, no config files. Existing drift detection test
  (`TestEmbeddedAssets_MatchSource`) enforces file sync.
- **Blast radius**: 3 of 9 speckit commands across all projects
  using the Unbound Force scaffold

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change modifies a slash command (Markdown instructions)
that produces file-based artifacts. No runtime coupling or
synchronous interaction is introduced. The corrected guardrails
make each command's behavioral constraints self-describing —
an improvement over the current state where guardrails
contradict the command's own instructions.

### II. Composability First

**Assessment**: PASS

`/uf.init` already handles missing files gracefully. The new
command-specific variants follow the same idempotent pattern.
No new mandatory dependencies are introduced. Each command
remains independently usable.

### III. Observable Quality

**Assessment**: PASS

The command produces a structured summary report with status
indicators for every file processed. The new variants follow
the existing reporting pattern. No output format changes.

### IV. Testability

**Assessment**: PASS

The existing drift detection test
(`TestEmbeddedAssets_MatchSource`) ensures both copies stay in
sync. The fix should include guardrail content regression tests
to verify each command type receives appropriate guardrails —
this closes a gap identified by the triage panel (zero guardrail
content validation tests exist today).

### V. Security by Default

**Assessment**: PASS

Incorrect guardrails for `speckit.taskstoissues.md` currently
fail to mention the command's most security-sensitive operation
(creating GitHub issues via external API) while granting
irrelevant local file write permissions. The fix corrects this
by describing the actual operation model, improving security
awareness for agents executing the command.
