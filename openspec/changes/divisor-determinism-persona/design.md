## Context

PR #959 shipped the persona and its evidence tooling; this
change documents the design decisions retroactively. The
Determinist is repo-local (not scaffolded by `uf init`):
deployments enroll it by adding the agent file and a
reviewer-capabilities entry.

## Goals / Non-Goals

Goals:

- One review owner per pack: DE-001..DE-008 move from the
  linguist to the Determinist.
- Mechanical evidence first: the persona consumes linter
  deltas, the seam report, and the tool inventory instead of
  re-deriving them (DE-001).
- Entry-point correctness: shared isMainModule entry
  detection because import.meta.main is undefined before
  Node 24.2 -- a CLI that silently no-ops and exits 0 is
  worse than one that fails loudly.

Non-Goals:

- Scaffolding the persona into `uf init` output (repo-local
  by design).
- New determinism rules: DE-001..DE-008 are unchanged; the
  persona owns their enforcement.

## Key Decisions

- Repo-local enrollment: the persona ships in this repository
  only; divisor-guard's deferral is written against the
  enrollment condition, not this repository's instantiation,
  so deployed Guards never defer to a persona they lack.
- Shared CLI plumbing (`.opencode/lib/cli-common.ts`): entry
  detection, crash conversion (exit 2), control-character
  escaping, and null-safe git error detail live in one module
  consumed by both evidence CLIs. `collectClis` detects CLIs
  by the runMain entry-guard idiom on comment-stripped code,
  not by prose mentions.
- Seam credit is pinned by a real-repo discoverability test:
  the extraction runs against the live registration-surface
  suite, so test-layout drift flips credit loudly instead of
  silently (DE-007).
