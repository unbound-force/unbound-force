## MODIFIED Requirements

### Requirement: Inline STOP in spec-phase commands

Each spec-phase command MUST have an inline STOP
instruction at the exact point where the command's
work is complete. The STOP MUST appear before any
output/report section, not in a separate Guardrails
section at the end.

Target files:
1. `.opencode/command/opsx-propose.md`
2. `.opencode/command/speckit.specify.md`
3. `.opencode/command/speckit.clarify.md`
4. `.opencode/command/speckit.plan.md`
5. `.opencode/command/speckit.tasks.md`
6. `.opencode/command/speckit.analyze.md`
7. `.opencode/command/speckit.checklist.md`
8. `.opencode/command/speckit.testreview.md`
9. `.opencode/skills/openspec-propose/SKILL.md`

### Acceptance Criterion

Given any spec-phase command or skill file (items
1-9 above), when the agent completes artifact
creation, then a STOP instruction reading "STOP
HERE. Do NOT proceed to implementation." appears
inline before any output or report section, and the
`## Guardrails` section includes reasoning for why
implementation is prohibited.

### Requirement: "Why" in Guardrails

Each file's existing `## Guardrails` section MUST
include reasoning for why implementation is prohibited
(user needs to review before implementation).

### Requirement: Skill guardrails

`.opencode/skills/openspec-propose/SKILL.md` MUST
include implementation prevention guardrails matching
the command file's guardrails.

## REMOVED Requirements

None.
