## Why

The determinism pack (DE-001..DE-008) had no single review
owner. divisor-linguist carried lint evidence, DE review, and
instruction-clarity review in one persona -- the widest load
set in the council -- and DE findings competed with clarity
findings for the same reviewer's attention. The ancestor
change (instruction-lint) left DE review inside the linguist
persona explicitly as a stepping stone.

This change is filed retroactively: the system shipped first
on the feature branch (PR #959) and this change bounds the
intent and acceptance criteria after the fact, per review
feedback.

## What Changes

- Add `.opencode/agents/divisor-determinism.md`, a repo-local
  Divisor persona (The Determinist) that owns determinism pack
  rules DE-001 through DE-008, enrolled in
  `.uf/reviewer-capabilities.yaml` with scopes `documentation`
  and `standard`.
- Re-scope `divisor-linguist`: it no longer runs a DE review --
  it defers DE findings and the IL-040..IL-042 hints to the
  new persona and keeps lint evidence and instruction clarity.
- Make `divisor-guard` defer DE-008 to the enrolled determinism
  persona in deployment-neutral wording (the enrollment
  condition, not this repository's instantiation).
- Add `.opencode/lib/tool-inventory.ts`, a plain-node CLI that
  lists registered plugin tools and repo scripts (`--match
  WORDS` for owner lookup) and maps added, modified, or
  removed tools to their tests by seam layer (`--seams --base
  REF`, DE-007).
- Add `.opencode/lib/cli-common.ts`, shared CLI plumbing:
  entry detection via isMainModule (import.meta.main is
  undefined before Node 24.2), crash-to-exit-2 conversion,
  control-character escaping, and null-safe spawn detail.
- Add instruction-lint rule IL-042: delegation call blocks
  that omit `task_classification` or the skill contract
  (DE-008), and `--base REF` per-file line, token, and
  load-set deltas (DE-004).

## Capabilities

- `divisor-determinism-persona` (new): the persona, its
  enrollment, the re-scopes, and the evidence CLIs described
  above.

## Impact

- Code: `.opencode/agents/divisor-determinism.md` (246
  lines), `.opencode/lib/tool-inventory.ts`,
  `.opencode/lib/cli-common.ts`,
  `.opencode/lib/instruction-lint.ts` (--base deltas, IL-042,
  --changed), `.uf/reviewer-capabilities.yaml`.
- Tests: `.opencode/test/tool-inventory.test.ts`,
  `.opencode/test/cli-common.test.ts`,
  `.opencode/test/instruction-lint.test.ts`,
  `.opencode/test/instruction-lint-preflight.test.ts`.
- Instruction delta (DE-004): divisor-determinism is new, 246
  lines; divisor-linguist shrinks 236 -> 203 lines (load set
  -3880 tokens); divisor-guard +2 lines (DE-008 deferral).
- Docs: CHANGELOG entry; AGENTS.md agent-count update;
  docs/heroes.md and docs/usage.md 11-persona council notes.

## Constitution Alignment

- Principle VI (Deterministic Execution): the persona consumes
  mechanical evidence (linter deltas, seam report, inventory)
  instead of re-deriving it (DE-001); its own instruction
  surface is lint-covered.
- Principle V (least privilege): the persona's bash allowlist
  is the linter wrapper and the tool-inventory entry, behind
  terminal metacharacter denies; all inputs are treated as
  untrusted data.
- Principle III (provenance): evidence CLIs stamp JSON output
  with producer provenance (tool, version, generatedAt).
