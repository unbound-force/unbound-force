<!--
  All tasks modify overlapping files, so no tasks are
  parallel-eligible.
-->

## 1. Convention Pack Rules

- [x] 1.1 Add Content Portability section with CP-001,
  CP-002, CP-003 rules to
  `.opencode/uf/packs/default.md` and its canonical
  scaffold copy. (D1)

## 2. Targeted Guardrails

- [x] 2.1 Add content portability guardrail to
  `.opencode/commands/uf.triage-issue.md` child issue
  creation section. (D2)

- [x] 2.2 Add content portability guardrail to
  `.opencode/commands/speckit.taskstoissues.md`. (D2)

- [x] 2.3 Add content portability guardrail to
  `.opencode/commands/uf.init.md` taskstoissues
  template block. (D2)

- [x] 2.4 Add Content Portability section to
  `.opencode/agents/divisor-curator.md` audit
  checklist. (D2)

- [x] 2.5 Sync all scaffold canonical copies under
  `internal/scaffold/assets/opencode/`. (D3)

## 3. Regression Tests

- [x] 3.1 Add `TestContentPortability_DefaultPack`
  verifying CP-001/CP-002/CP-003 markers in the
  embedded default convention pack.

- [x] 3.2 Add `TestContentPortability_TriageIssueCommand`
  verifying content portability in the triage-issue
  command's section 4.4.

- [x] 3.3 Add content portability assertions to the
  existing `TestGuardrailTemplates_CommandSpecificContent`
  for the taskstoissues guardrails block.

- [x] 3.4 Add `TestContentPortability_CuratorAgent`
  verifying Content Portability section in the
  divisor-curator agent.

## 4. Documentation

- [x] 4.1 Add CHANGELOG.md entry under Unreleased > Fixed.
