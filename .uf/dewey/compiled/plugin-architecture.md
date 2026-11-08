---
tier: validated
compiled_at: 2026-10-25T16:32:58Z
compiled_by: rits/zai-org/glm-5-3
sources:
  - plugin-scaffold-20261007T185850-jay-flowers
  - opencode-plugins-20261008T223708-jay-flowers
  - typescript-immutability-20261012T005044-jay-flowers
  - typescript-immutability-20261012T005400-jay-flowers
  - typescript-immutability-20261012T011814-jay-flowers
  - plugin-architecture-20261012T160738-jay-flowers
  - plugin-architecture-20261012T173159-jay-flowers
  - plugin-architecture-20261013T021611-jay-flowers
  - opencode-plugins-20261014T165544-jay-flowers
  - opencode-plugins-20261014T184609-jay-flowers
  - opencode-plugin-sdk-20261015T234420-jay-flowers
  - opencode-plugins-20261016T132709-jay-flowers
  - plugin-architecture-20261016T135615-jay-flowers
  - plugin-activation-20261016T152641-jay-flowers
  - plugin-staging-20261016T162806-jay-flowers
  - plugin-staging-20261016T223739-jay-flowers
  - plugin-architecture-20261017T123523-jay-flowers
  - plugin-activation-20261017T145231-jay-flowers
  - plugin-architecture-20261017T181137-jay-flowers
  - plugin-architecture-20261020T172441-jay-flowers
  - plugin-architecture-20261022T212018-jay-flowers
topic: plugin-architecture
---

# Plugin Architecture

## Current State (October 2026)

### Factory & DI Conventions

- Tool factories follow `export function createXxxTool(exec: GitExec, cwd: string)` returning
  `tool({...})` — ~30 in uf-workflow, more in uf-finale; injection enables hermetic testing
  with mock executors and temp directories. `GitExec` is the args-array
  `(args: string[], cwd: string) => string` via `execFileSync`, no shell.
- `tool()` returns a complex framework-defined generic type; explicit return-type annotation
  is impractical, so inference is the sanctioned convention (CS-005 exception).
- The SDK recognizes only the `args` key (zod ZodRawShape); a `parameters` key is silently
  ignored — empty schemas, stripped runtime arguments. New tools MUST declare zod `args:`;
  parameterless tools omit the key; parameters arriving empty indicate wrong declaration style.
- Handler output must flow through the structured JSON result envelope — never
  `console.log`/`console.error`: handlers run in the opencode server process sharing the
  TUI's alternate-screen terminal; subprocess stdio piped and captured, never inherited.
- TypeScript immutability: prefer readonly/frozen inputs in tool schemas; extend factories
  (e.g., `makeProvenance()` for `InvocationProvenance`) to accept new fields rather than
  mutating readonly fields post-construction — TS readonly is compile-time only.
- Extract per-tool validation glue at the third verbatim copy (CS-004 threshold);
  `resolveRepoRelative` (repo-relative traversal-failure envelope) is the precedent —
  grep new sibling factories for the shared failure-message prefix.

### Shared Library & Envelope Types

- Shared envelope types (`ToolSuccess`/`ToolFailure`/`ToolResult`) live in
  `.opencode/lib/uf-workflow-types.ts`, re-exported from the uf-workflow and uf-finale
  barrels; invoke-agent and review-dispatch do not re-export them.
- Enriching a success envelope moves three documentation surfaces together: the factory's
  `description` string (machine-facing API doc), command templates under
  `.opencode/commands/` plus their scaffold mirrors, and unit tests asserting the field.

### Plugin Split

- Shared primitives live in uf-workflow (git ops, model normalization, slug sanitization);
  command-specific composites in uf-finale (startup, stage, PR context, CI poll); both re-export.
- The uf.finale command template dropped 1,104 -> 333 lines (~70%) at extraction while
  preserving all LLM decision points; it has since regrown to 391 lines (base-sync step).
- `uf-workflow/index.ts` is now 4,902 lines — it grew rather than split, so the
  mirror-sync constraints below remain live.

### Dependency Pins

- `.opencode/package.json` pins @opencode-ai/plugin 1.19.0, zod 4.6.6 (with a zod
  override at 4.6.6 forcing nested copies); vitest 5.0.3 stays in this repo's
  devDependencies only — the scaffolded mirror ships none (lockless install, D2a).
- Four plugins are registered in `opencode.json` (invoke-agent, review-dispatch, uf-workflow,
  uf-finale); registration is by directory, so `index.ts` must remain the barrel entry point.

### Cross-Plugin Contracts

- Filesystem/state contracts between plugins (one writes under `.uf/`, another reads) belong in
  `.opencode/lib/` shared modules — single-sourced path constants, type discriminators, typed
  accessors — drift-test-pinned; precedents: `uf-workflow-types.ts`, `branch-classification.ts`.
- Duplicated contracts fail open silently: a producer-side relocation or envelope change
  makes the consumer report "not done" with no error, re-running expensive workflows (full
  review councils); register consumers in the artifact-type registry (schemas/README.md).
- OPEN — progress-state pair: uf-workflow writes `.uf/progress/<sanitized-branch>/uf.unleash.json`
  while uf-finale independently reconstructs the path and re-implements the
  completed_steps / positive-integer guards (uf-finale/index.ts:560); test-pinned
  end-to-end (issue #707 precedent), but the shared-module extraction has not happened.
- When an extraction touches scaffold mirrors, live and mirrored copies must remain
  byte-identical, with drift tests verifying identity in the same increment.

### Mirror-Sync Constraints

- Adding a new plugin requires six scaffold updates: scaffold.go constant,
  `isActivationGatedAsset()`, plugin JSON registration loop and hasPlugins check,
  `isToolOwned()`/`isDivisorAsset()`, plugin_activation.go staging and activation
  lists, and tests (expectedAssetPaths, task42CanonicalAssets, TestRunInit_FreshDir).
- A new `.opencode/lib/` module imported by the activation-gated plugins requires three
  lockstep Go registrations: (1) `stageReviewPluginSources` asset list — the probe imports
  staged plugins via file:// URLs and ESM resolves `../../lib/*.js` strictly inside the
  stage root; (2) `isToolOwned` — unregistered modules are created once but never
  refreshed (version-skew import failures at startup); (3) `isDivisorAsset` —
  unregistered modules drop from DivisorOnly deployments.
- Existing guards miss staging drift: faked-probe unit tests never execute staged imports,
  and TestRunInit_FreshDir's count range plus D12's partial-run-exits-0 mask activation
  failures — `make check` proves neither probe success nor registration completeness.
- Closure tests must walk the import graph transitively, derive the expected set from
  observable staging output (not re-asserted copied lists), and assert every sibling
  registration list — not just staging.
- Splitting a mirrored plugin file requires lockstep updates to four surfaces: the canonical
  `.opencode/` tree, the byte-identical `internal/scaffold/assets/` mirror,
  `expectedAssetPaths`, and `task42CanonicalAssets`; the closure test's import grammar
  understands only the flat `opencode/lib` layout — extend it before any plugin-internal split.

## History

- **Oct 7 — scaffold gating.** Six-location activation checklist; misses surface as SCFR003 test failures.
- **Oct 8 — inference exception.** Framework-wrapped `tool()` returns: inference sanctioned (CS-005).
- **Oct 12 — DI hardening.** Injected-GitExec factories established; `GitExec` moved from
  shell-mode `(cmd: string, cwd: string)` to args-array, no shell (commit f2d9143);
  `makeProvenance()` readonly-mutation lessons (advisory MEDIUM).
- **Oct 13 — two-plugin split.** finale-token-reduction moved primitives to uf-workflow,
  composites to uf-finale; template cut ~70% (1,104 -> 333 lines), LLM decision points kept.
- **Oct 14 — output discipline.** TUI corruption from console.log fixed (18 + 32 calls,
  duplicated in scaffold tree); envelope-field doc surfaces codified (`files` gap).
- **Oct 15-16 — SDK defect.** `parameters:` vs `args:` broke checkpoint/get_progress; "14
  legacy tools" corrected to 8 + 3 parameterized factories (6 parameterless, correct) —
  grep-verify enumerations; detect_resume's duplicated store motivated lib/ contracts.
- **Oct 16-17 — staging closure.** Import-closure tests hardened (transitive, derived,
  lockstep); PR 745's 11 factories exposed glue duplication (envelope x8, CRLF x3),
  yielding the third-copy heuristic and `resolveRepoRelative`.
- **Oct 17 — mirror-sync enumeration.** Four-surface split lockstep enumerated at
  uf-workflow/index.ts 3,860 lines; barrel and flat-grammar constraints recorded.
- **Oct 20-22 — state contracts.** Progress-state writer-reader pair flagged as silent-drift
  risk; test-pinned end-to-end, extraction deferred pending spec ceremony.
- **Oct 25 — audit.** uf-workflow/index.ts at 4,902 lines (grew, not split); uf.finale at
  391 lines; progress-state extraction still OPEN.
