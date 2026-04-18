## 1. OpenSpec Propose (Command + Skill)

- [x] 1.1 Add inline STOP after Step 5 artifact
  creation loop in `.opencode/command/opsx-propose.md`
- [x] 1.2 Add "why" reasoning to the existing
  `## Guardrails` section in `opsx-propose.md`
- [x] 1.3 Add implementation prevention guardrails +
  inline STOP to
  `.opencode/skills/openspec-propose/SKILL.md`

## 2. Speckit Commands

- [x] 2.1 Add inline STOP after spec is written +
  "why" to guardrails in `speckit.specify.md`
- [x] 2.2 Add inline STOP after clarifications
  resolved + "why" to guardrails in `speckit.clarify.md`
- [x] 2.3 Add inline STOP after plan.md created +
  "why" to guardrails in `speckit.plan.md`
- [x] 2.4 Add inline STOP after tasks.md created +
  "why" to guardrails in `speckit.tasks.md`
- [x] 2.5 Add inline STOP after analysis report +
  "why" to guardrails in `speckit.analyze.md`
- [x] 2.6 Add inline STOP after checklist created +
  "why" to guardrails in `speckit.checklist.md`
- [x] 2.7 Add inline STOP after testability report +
  "why" to guardrails in `speckit.testreview.md`

## 3. Verification

- [x] 3.1 Verify all 9 files have inline STOP
  (grep for "STOP HERE")
- [x] 3.2 Verify all 9 files have "why" in guardrails
  (grep for "review the plan before")
- [x] 3.3 Run `go build ./...` to verify no
  compilation issues from scaffold drift

<!-- spec-review: passed -->
<!-- code-review: passed -->
