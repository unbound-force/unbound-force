<!-- capability: delegation-timeouts
     change: per-task-delegation-budgets
     originating_issue: 1080 -->

## MODIFIED Requirements

### Requirement: Hierarchical Delegation Budgets

**ID: DT-FR-001** Extended — registry applicability for per-task children.

The system MUST preserve the existing `delegation-budget.ts` `globalThis`
symbol-keyed deadline registry, `resolveInheritedDeadline` (at most 8 parent
hops), `clampToBudget` (60 s reserve / 30 s minimum), and
`budget_exhausted` fail-fast semantics unchanged. Per-task
`uf_delegate_command` children running inside a budgeted command-path
`/uf.unleash` parent MUST resolve and clamp via the same registry entry. No
new global idle default SHALL be introduced.

See `per-task-delegation-budgets` PTB-FR-002 for idle-inheritance behavior
which reuses this registry.

#### Scenario: Per-task child clamps to parent deadline

- **GIVEN** a per-task child inside a budgeted `/uf.unleash` parent
- **WHEN** the executor resolves its inherited deadline
- **THEN** the existing `resolveInheritedDeadline` + `clampToBudget` path MUST
  be used

### Requirement: Idle Timeout

**ID: DT-FR-002** Extended — idle watchdog inherited by per-task children.

The existing `idle_timeout` / `probeSubtreeActivity` mechanism MUST remain
unchanged. When a per-task child runs under a parent carrying an
`idle_timeout`, the watchdog MUST be inherited via the parent's registry
entry with the same DFS, cap, and confirming-probe rules. See PTB-FR-002.

#### Scenario: Per-task child inherits idle watchdog

- **GIVEN** a per-task child under a parent carrying `idle_timeout`
- **WHEN** the idle watchdog is evaluated
- **THEN** the child's `probeSubtreeActivity` MUST be polled via the inherited
  parent registry entry
