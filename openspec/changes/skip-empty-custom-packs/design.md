## Context

`uf init` generates a managed block in `CLAUDE.md` via
`buildCLAUDEmdBlock` (`scaffold.go:1055`), which calls
`collectDeployedPacks` (`scaffold.go:1259`) to assemble the
list of `@`-import lines. `collectDeployedPacks` returns all
custom packs unconditionally — it has no knowledge of whether
the files on disk contain real content.

Three custom pack stubs (`default-custom.md`, `go-custom.md`,
`content-custom.md`) are thus always imported, even though
they contain only boilerplate scaffolding and the placeholder
comment `<!-- Add project-specific rules below this line -->`.
No rules are present in any of these files across the five
repos in the org.

The proposal establishes that this is a pure scaffold-engine
concern, with no changes to embedded asset templates or CLI
interface. Constitution alignment was assessed as PASS on all
four principles.

## Goals / Non-Goals

### Goals
- Custom pack `@` imports are omitted from `CLAUDE.md` when
  the corresponding file on disk contains no rules below the
  placeholder sentinel.
- `uf init --reinit` re-evaluates emptiness on each run, so
  a pack that gains content is automatically re-added.
- The existing `collectDeployedPacks` test surface is
  preserved (function signature is backward-compatible).
- The new helper `hasRuleContent` is unit-testable in
  isolation with `t.TempDir()` fixture files.
- The existing `CLAUDE.md` in this repo has its three empty
  custom pack `@` imports removed as part of this change.

### Non-Goals
- Modifying the embedded scaffold asset templates — stubs
  continue to be written to disk on first `uf init`.
- Emitting a user-visible warning or log line when a pack is
  skipped (may be added later; not required for correctness).
- Changing the typescript-custom.md behaviour — it follows
  the same code path and benefits automatically.
- Any change to non-Go repos or cross-repo scaffolding.

## Decisions

### D1 — Sentinel-based emptiness detection

**Decision**: A pack is "empty" if it contains no
non-whitespace content after the last occurrence of the HTML
comment `<!-- Add project-specific rules below this line -->`.

**Rationale**: The sentinel is stable (defined in the embedded
asset template), unambiguous (it appears exactly once in every
stub), and robust against whitespace-only trailing content.
Alternative: check overall file size — rejected because it
would misfire if a user inadvertently saves trailing
whitespace, and it doesn't pinpoint the rule section.

### D2 — `hasRuleContent(path string) bool` helper

**Decision**: Introduce a new unexported helper that takes a
file path, reads the file, finds the sentinel, and returns
`true` iff there is non-whitespace content after it. Returns
`true` on any I/O error (fail-open: if we can't read the file,
include it — don't silently drop a pack the user may have
populated).

**Rationale**: Fail-open is safer. If `CLAUDE.md` is being
generated in a context where the custom pack is inaccessible
(e.g., a filesystem permission edge case), including the
import is harmless; omitting it could silently drop rules.

### D3 — `collectDeployedPacks` gains optional root parameter

**Decision**: Add a second parameter `root string` to
`collectDeployedPacks`. When `root == ""`, the function
behaves identically to before (all packs included). When
`root != ""`, each custom pack filename is checked via
`hasRuleContent(filepath.Join(root, ".opencode/uf/packs/",
name))`. Non-custom packs are never filtered.

**Rationale**: Keeping the root optional preserves all
existing call sites in tests that pass no directory. It also
makes the function usable in the existing `shouldDeployPack`
divisor path without requiring a filesystem.

### D4 — `buildCLAUDEmdBlock` passes `opts.TargetDir`

**Decision**: `buildCLAUDEmdBlock` gains a `root string`
parameter and passes it straight through to
`collectDeployedPacks`. `ensureCLAUDEmd` passes
`opts.TargetDir` (already available at the call site,
`scaffold.go:1093`).

**Rationale**: Minimal change surface. No other callers of
`buildCLAUDEmdBlock` exist; they are both internal and in
tests. Updating the signature is safe.

### D5 — Remove three `@` imports from CLAUDE.md in this repo

**Decision**: As part of this PR, manually remove the three
empty custom pack `@` import lines from `CLAUDE.md`. The next
`uf init --reinit` would do this automatically, but doing it
now keeps the repo immediately consistent.

**Rationale**: This is the most visible and immediate fix for
the issue reporter. It also serves as a concrete regression
target in tests.

## Risks / Trade-offs

**Risk**: A user populates a custom pack, runs `uf init
--reinit`, and expects the import to be added back — this now
works correctly (D2 fail-open + D3 conditional check), but it
requires the file to exist at `opts.TargetDir` at the time
of `uf init`. If the file is populated but the tool is run
from a different working directory, the file won't be found
and the pack will be omitted. Accepted: this is an edge case;
the normal flow is to run `uf init` from the repo root.

**Trade-off**: `hasRuleContent` reads each custom pack file
on every `uf init` run. For three small (~20-line) files this
is negligible. Even in a pathological multi-pack scenario the
cost is trivially small.

**No backward-incompatible interface change**: The
`collectDeployedPacks` signature change is internal. No
exported API is altered.
