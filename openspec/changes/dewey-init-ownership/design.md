## Context

Spec 017 established the principle: `uf setup` handles
system-level installs, `uf init` handles repo-level
configuration. That spec moved `opencode.json` from
setup to init. This change applies the same principle
to dewey init/index operations.

Currently `uf setup` has 15 steps. Steps 13-14 are
`initDewey()` and `indexDewey()`. Step 15 runs
`uf init`, which also attempts dewey init/index via
`initSubTools()` but skips because `.dewey/` already
exists from steps 13-14.

## Goals / Non-Goals

### Goals
- Remove dewey init/index from `uf setup`
- Renumber setup steps from 15 to 13
- Add `dewey index` re-run on `uf init --force`
- Simplify setup tests (fewer mocked dependencies)

### Non-Goals
- Changing dewey install in setup (step 12 stays)
- Changing the dewey init/index logic itself
- Adding new dewey features
- Changing how `uf doctor` checks dewey

## Decisions

### D1: Remove both initDewey and indexDewey from setup

Remove both functions and their call sites. The comment
at `setup.go:787-790` that documents the duplication
is also removed (the duplication it describes will no
longer exist).

### D2: Renumber to 13 steps

Steps after the removed ones shift down:
- Old step 15 (runUnboundInit) becomes step 13
- Progress messages update from `[N/15]` to `[N/13]`

### D3: Force re-index in initSubTools

When `.dewey/` already exists AND `opts.Force` is true,
run `dewey index` with a progress message:
`"  Re-indexing Dewey sources..."`. This gives users
a way to refresh the index after adding new files.

When `.dewey/` already exists AND `opts.Force` is
false, skip silently (current behavior, preserving
idempotency).

### D4: Preserve existing initSubTools behavior

The existing dewey init/index block in `initSubTools()`
(lines 666-699) is unchanged for the first-run case.
The only addition is the force re-index path for the
`.dewey/`-already-exists case.

## Risks / Trade-offs

### Risk: Setup users lose direct dewey init/index

Users running `uf setup` will no longer see explicit
`[13/15] Dewey workspace...` and `[14/15] Dewey index`
steps. Instead, dewey init/index happens inside
`uf init` at the final step with different progress
messages.

**Mitigation**: `uf init` already has progress messages
("Initializing Dewey workspace...", "Indexing Dewey
sources (this may take a moment)...") that provide
equivalent UX. The user sees the same information,
just under the `uf init` umbrella.

### Trade-off: Force re-index adds latency

`uf init --force` will re-index even if the index is
already current. For large repos with many sibling
sources, indexing can take 10-60 seconds.

**Acceptance**: This is opt-in (only with `--force`).
The default `uf init` remains fast. Users who want a
fresh index explicitly accept the latency.
