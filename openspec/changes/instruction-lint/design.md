## Context

PR #958 shipped the system; this change documents the design
decisions retroactively. The linter runs wherever devDependencies
are installed (make target, advisory pre-flight check, persona
invocation, vitest).

## Goals / Non-Goals

Goals:

- Deterministic checks only (same input, same output), per the
  determinism pack; judgment stays in review packs.
- Findings carry their Divisor review severity so reviewers never
  re-map it (DE-006).
- An exit contract that separates findings (1) from
  could-not-run (2) at every entry point.

Non-Goals:

- Judgment checks (clarity, correctness of intent) — owned by
  the divisor-linguist persona's Steps 2-3, not the tool.
- Style or voice linting — owned by the content and
  review-voice packs.

## Decisions

1. **Exit 0/1/2 contract.** 0 clean, 1 findings (errors, or
   warnings with --strict), 2 could-not-run (bad arguments, bad
   config, unreadable tree, crash). The contract is only as
   strong as its weakest entry point: the Makefile target and
   the persona both invoke the wrapper, never node on the
   library directly, because static ESM imports crash before
   main() can catch them. The wrapper pre-checks the linter
   entry file, the node_modules directory, and the yaml and zod
   package dirs.
2. **--only is a usage-error surface.** A --only path matching
   no lintable file exits 2 before analysis; a typo must never
   filter findings away and report clean.
3. **Config is untrusted input, validated strictly.** Zod
   schema, `.strict()` (unknown keys are rejected, not
   stripped); ignorePaths entries must be non-empty (an empty
   prefix would startsWith-match every file and suppress all
   findings).
4. **Suppression and dedup semantics.** IL-050 caps shingle
   posting lists at 64 during candidate generation (verification
   still scores every shingle); byte-identical copies cluster
   via an O(n) exact pass, and the near-identical-beyond-cap
   recall cliff is accepted boilerplate filtering, pinned by
   test.
5. **devDependencies placement.** `yaml` is declared as a
   pinned devDependency: repo-local tooling runs only where
   devDependencies are installed, and scaffolded targets install
   with --omit=dev and must not carry it. Pin surfaces
   (manifest, lockfile, embedded mirrors, provenance testdata,
   exact-version test maps) all updated in the same change.
6. **Persona hardening.** The persona consumes untrusted diff,
   pre-flight, and linter output; its Review Scope requires
   treating all inputs as data (IQ-11), and its bash permission
   block places terminal metacharacter denies (redirection,
   pipes, separators, command and process substitution) after
   every allow, so no allow prefix can chain, redirect, or
   substitute.

## Risks / Trade-offs

- The IL-050 posting-list cap trades recall for bounded runtime
  at scale; the cliff is narrow (near- but not byte-identical
  text replicated beyond 64 files) and pinned by test.
- The persona's token load (~13.4k) consumes most of the 20k
  budget; further persona growth needs pruning, not budget
  raises.
- Advisory gating: instruction-lint findings do not fail
  pre-flight (the repo baseline carries one pre-existing error);
  tightening to required is deferred until the baseline is
  clean.
- Accepted residuals, tracked in issue #971: the metacharacter
  deny set does not cover newline command separation or
  backslash-newline continuation (the permission engine's
  multi-line matching semantics are themselves unverified);
  the wrapper's package checks verify manifest presence, not
  loadability, so a corrupted-but-manifested package still
  exits 1; and node-binary absence exits 127 at the make and
  persona lanes (the pre-flight lane guards it with
  `requires: [node]`).
