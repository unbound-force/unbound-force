## 1. OpenSpec Propose — Fix Missing `/unleash`

- [x] 1.1 Update `.opencode/command/opsx-propose.md` line 12: change intro from `run /opsx-apply` to `run /unleash (autonomous) or /opsx-apply (sequential)`
- [x] 1.2 Update `.opencode/command/opsx-propose.md` line ~122: change output prompt to list `/unleash` first: `"Run '/unleash' for autonomous pipeline execution, '/opsx-apply' for sequential implementation, or '/cobalt-crush' for direct coding."`
- [x] 1.3 Update `.opencode/command/opsx-propose.md` lines ~150-153: add `/unleash` to NEVER-run guardrail and user prompt
- [x] 1.4 Update `.opencode/skills/openspec-propose/SKILL.md` line 19: change intro from `run /opsx-apply` to `run /unleash (autonomous) or /opsx-apply (sequential)`
- [x] 1.5 Update `.opencode/skills/openspec-propose/SKILL.md` line ~173: change output prompt to list `/unleash` first
- [x] 1.6 Update `.opencode/skills/openspec-propose/SKILL.md` line ~198: add `/unleash` to NEVER-run guardrail

## 2. Cobalt-Crush Command — Fix Fallback and Typo

- [x] 2.1 Update `.opencode/command/cobalt-crush.md` lines ~88-97: add `/unleash` as first option in the no-context fallback prompt
- [x] 2.2 Fix typo in `.opencode/command/cobalt-crush.md`: change `/opsx:apply` to `/opsx-apply` (2 occurrences)

## 3. uf-init Command — Fix Template and Next Steps

- [x] 3.1 Update `.opencode/command/uf-init.md` lines ~503-518: add `/unleash` to the OpenSpec guardrails template block (NEVER-run list and user prompt)
- [x] 3.2 Update `.opencode/command/uf-init.md` lines ~980-985: add `/unleash` to Next Steps section as primary recommendation

## 4. Speckit Spec-Phase Commands — Reorder STOP Blocks

- [x] 4.1 Update `.opencode/command/speckit.specify.md`: reorder STOP block from `(/opsx-apply, /cobalt-crush, or /unleash)` to `(/unleash, /cobalt-crush, or /opsx-apply)`
- [x] 4.2 Update `.opencode/command/speckit.clarify.md`: same reorder
- [x] 4.3 Update `.opencode/command/speckit.plan.md`: same reorder
- [x] 4.4 Update `.opencode/command/speckit.tasks.md`: same reorder
- [x] 4.5 Update `.opencode/command/speckit.analyze.md`: same reorder
- [x] 4.6 Update `.opencode/command/speckit.checklist.md`: same reorder
- [x] 4.7 Update `.opencode/command/speckit.testreview.md`: same reorder

## 5. Explore Command — Add `/unleash` to Proposal Flow

- [x] 5.1 Update `.opencode/command/opsx-explore.md` lines ~148-157: add `/unleash` mention to the "Flow into a proposal" ending bullet

## 6. Speckit Workflow Skill — Add Entry Point

- [x] 6.1 Add Entry Point section to `.opencode/skill/speckit-workflow/SKILL.md` identifying `/unleash` as the primary command

## 7. Scaffold Asset Sync

- [x] 7.1 Copy `.opencode/command/cobalt-crush.md` to `internal/scaffold/assets/opencode/command/cobalt-crush.md`
- [x] 7.2 Copy `.opencode/command/uf-init.md` to `internal/scaffold/assets/opencode/command/uf-init.md`
- [x] 7.3 Copy `.opencode/skill/speckit-workflow/SKILL.md` to `internal/scaffold/assets/opencode/skill/speckit-workflow/SKILL.md`

## 8. Verification

- [x] 8.1 Grep all modified files for remaining instances of `(/opsx-apply, /cobalt-crush, or /unleash)` — should be zero (all reordered)
- [x] 8.2 Grep all `.opencode/command/` and `.opencode/skills/` files for "Run `/opsx-apply`" without `/unleash` — should be zero
- [x] 8.3 Run `go test ./internal/scaffold/...` to verify scaffold drift tests pass
<!-- spec-review: passed -->
<!-- code-review: passed -->
