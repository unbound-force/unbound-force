<!--
  Retroactive change: the system shipped on PR #958's branch
  before this change was filed; tasks are recorded as completed.
-->

## 1. Linter

- [x] 1.1 Implement .opencode/lib/instruction-lint.ts: artifact
  discovery, frontmatter, references, pack rules, prose,
  budgets, duplicates (IL-001..IL-050), JSON output, exit 0/1/2
- [x] 1.2 JSDoc every exported symbol (typescript pack DR-001)
- [x] 1.3 Exit 2 for --only paths matching no lintable file;
  document it in --help
- [x] 1.4 Strict config schema: unknown keys rejected,
  ignorePaths entries non-empty
- [x] 1.5 Memoize proseUnits and parsePackRules (each is
  consumed by two checks)

## 2. Wrapper and entry points

- [x] 2.1 .uf/scripts/instruction-lint: dependency pre-checks
  (entry file, node_modules, yaml and zod package dirs) before
  node starts
- [x] 2.2 Makefile lint-instructions routes through the wrapper
- [x] 2.3 .uf/preflight.yaml declares the advisory
  instruction-lint check
- [x] 2.4 divisor-linguist persona invokes the wrapper, not node

## 3. Persona

- [x] 3.1 Linter-evidence-first review flow with DE and IQ
  tables
- [x] 3.2 Bash permissions: wrapper / git diff / git status
  allows, narrow flag denies, terminal metacharacter denies
- [x] 3.3 Treat-as-data directive; exit 2 documented in the
  contract and graceful degradation

## 4. Dependencies

- [x] 4.1 Declare yaml as an exact-pinned devDependency with the
  lockfile regenerated and scaffold mirrors synced
- [x] 4.2 Extend pin and provenance tests; replace the
  review-matrix exemption with extends-canonical divergence
  checks

## 5. Tests

- [x] 5.1 Unit suite for every rule, suppression, config, CLI
  (instruction-lint.test.ts)
- [x] 5.2 Seam tests: declaration, round-trip, no-deps, entry
  and partial installs, split-root
  (instruction-lint-preflight.test.ts)
- [x] 5.3 Posting-list cap: exact-duplicate clustering beyond
  the cap; near-identical cliff pinned
