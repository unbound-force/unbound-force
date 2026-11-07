## Why

The agent ecosystem is driven by prose: agents, commands, skills,
convention packs, and AGENTS.md. Instruction text is executed by
agents, but nothing checked it mechanically — reference rot
(slash commands that no longer exist), RFC 2119 drift, vague
qualifiers in normative sentences, token-budget overruns, and
copy-paste instruction drift across files were all invisible
until an agent mis-executed them. The ancestor changes
(determinism-codification, command-prose-pilot) deferred this
work explicitly.

This change is filed retroactively: the system shipped first on
the feature branch (PR #958) and this change bounds the intent
and acceptance criteria after the fact, per review feedback.

## What Changes

- Add `.opencode/lib/instruction-lint.ts`, a deterministic
  linter for instruction prose (rules IL-001..IL-050: frontmatter
  shape, slash-command / agent / skill / path references, pack
  rule markers and RFC 2119 force, vague qualifiers, file and
  load-set token budgets, near-duplicate paragraphs), with a
  0/1/2 exit contract (0 clean, 1 findings, 2 could-not-run) and
  JSON output.
- Add `.uf/scripts/instruction-lint`, a dependency-checking
  wrapper that verifies the linter entry and the yaml/zod
  packages before starting node, so a fresh or partial checkout
  exits 2 instead of an ESM crash exiting 1.
- Add the `make lint-instructions` target, routed through the
  wrapper.
- Declare an advisory, instruction-scoped `instruction-lint`
  check in `.uf/preflight.yaml`.
- Add the `divisor-linguist` persona: linter evidence first,
  then DE-001..DE-008 determinism review, then IQ-1..IQ-11
  clarity review; bash allows only the wrapper, `git diff`, and
  `git status`, behind terminal metacharacter denies.
- Declare `yaml` as an exact-pinned devDependency of the plugin
  suite (it was imported but resolved only through transitive
  hoisting), with scaffold mirrors and pin tests updated.

## Capabilities

- `instruction-lint` (new): the linter, wrapper, make target,
  pre-flight declaration, and persona described above.

## Impact

- Code: `.opencode/lib/instruction-lint.ts` (983 lines),
  `.uf/scripts/instruction-lint` (42 lines), Makefile,
  `.uf/preflight.yaml`, `.uf/reviewer-capabilities.yaml`.
- Tests: `.opencode/test/instruction-lint.test.ts` (529 lines),
  `.opencode/test/instruction-lint-preflight.test.ts` (235
  lines); `internal/scaffold/scaffold_test.go` extends-canonical
  checks.
- Instruction delta (DE-004): `divisor-linguist.md` is new, 262
  lines (~4.2k self tokens, ~13.6k loaded); `AGENTS.md` grows
  by 9 lines net (+13/-4: lint-instructions command,
  `.uf/scripts/` structure entry, agent count, yaml
  devDependency note, zod pin, npm security overrides).
  Baseline lint findings are unchanged: 1 error, 31
  warnings, 23 info before and after.
- Docs: CHANGELOG entry; AGENTS.md build-command and structure
  updates; docs/heroes.md and docs/usage.md repo-local Linguist
  notes; documentation issue filed for user-facing coverage.

## Constitution Alignment

- Principle VI (Deterministic Execution): every check in the
  linter is mechanical; the persona consumes its output instead
  of re-deriving it (DE-001).
- Principle V (least privilege): the persona's bash allowlist is
  three read-only prefixes behind terminal metacharacter denies,
  with edit and webfetch denied.
