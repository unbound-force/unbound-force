## 1. README.md Fixes

- [x] 1.1 Replace the "Knowledge Graph" section
  (lines 64-67): remove all graphthulhu references,
  rewrite to describe Dewey as the semantic knowledge
  layer with MCP-based search. Reference
  `unbound-force/dewey` repo and
  `specs/014-dewey-architecture/` for the full spec.

- [x] 1.2 Fix `opencode.json` description (line 60):
  change "knowledge graph via graphthulhu" to
  "Dewey semantic knowledge layer".

- [x] 1.3 Fix spec count (line 55): change
  "10 architectural specifications" to
  "16 architectural specifications organized in
  four phases".

- [x] 1.4 Fix scaffold file count (line 47): change
  "47 files" to "50 files".

## 2. AGENTS.md Fixes

- [x] 2.1 Fix hero table: change "Embedded in
  `unbound` binary" to "Embedded in `unbound-force`
  binary" for both Cobalt-Crush and The Divisor rows.

- [x] 2.2 Fix hero status text below the table:
  replace "Gaze is the only hero with a functional
  implementation. The Divisor has a prototype
  deployment (reviewer agents) inside the Gaze repo."
  with an accurate description of all five heroes'
  implementation status.

- [x] 2.3 Fix Specification Framework intro: change
  "distributed via the `unbound` CLI binary" to
  "distributed via the `unbound-force` CLI binary
  (alias: `uf`)".

- [x] 2.4 Add specs 014 and 015 to the project
  structure tree under `specs/`:
  `014-dewey-architecture/` and
  `015-dewey-integration/` between 013 and 016.

- [x] 2.5 Fix Phase 0 description: change
  "Three core principles" to "Four core principles"
  for spec 001.

- [x] 2.6 Fix Phase 1 description: change
  "Three-persona review protocol" to "Five-persona
  review council" for spec 005.

- [x] 2.7 Add Dewey to the Sibling Repositories
  table: `unbound-force/dewey`, purpose "Semantic
  knowledge layer (MCP server)", status "Active".

## 3. Spec Frontmatter Status Updates

- [x] 3.1 Update `specs/012-swarm-delegation/spec.md`
  frontmatter: `status: draft` → `status: complete`

- [x] 3.2 Update `specs/013-binary-rename/spec.md`
  frontmatter: `status: draft` → `status: complete`

- [x] 3.3 Update `specs/014-dewey-architecture/spec.md`
  frontmatter: `status: draft` → `status: complete`

- [x] 3.4 Update `specs/015-dewey-integration/spec.md`
  frontmatter: `status: draft` → `status: complete`

- [x] 3.5 Update `specs/016-autonomous-define/spec.md`
  frontmatter: `status: draft` → `status: complete`

## 4. Verification

- [x] 4.1 Run `grep -rn 'graphthulhu' README.md` and
  verify zero matches.

- [x] 4.2 Run `grep -n 'Embedded in .unbound. binary'
  AGENTS.md` and verify zero matches for bare
  `unbound` (without `-force` suffix).

- [x] 4.3 Verify `go build ./...` still succeeds
  (no code changes, but sanity check).

- [x] 4.4 Verify constitution alignment: this change
  is documentation-only with no impact on Principles
  I, II, or IV. Principle III (Observable Quality) is
  satisfied by improving accuracy of provenance
  metadata (spec status fields).
