## 1. Update Pinkman Agent Dewey Integration

- [x] 1.1 In `.opencode/agents/pinkman.md`, replace the
  "After Scouting" subsection under "Dewey Integration"
  with the enriched tag taxonomy and content conventions.
  Update the tag from flat `pinkman` to mode-specific
  `pinkman-<mode>` (discover, trend, audit, report).
  Add content prefix convention (scouting-report:,
  trend-report:, dependency-audit:, adoption-report:).
  Replace "2-3 sentence summary" with mode-specific
  structured prose templates per the Structured Prose
  Content requirement in `specs/dewey-integration.md`.
- [x] 1.2 In each of the four mode sections (Discover,
  Trend, Audit, Report), update the existing one-liner
  Dewey reference (e.g., "Dewey: Store summary per
  Dewey Integration") to include the correct mode-specific
  tag and content prefix inline (e.g., "Dewey: Store per
  Dewey Integration using tag `pinkman-discover` and
  prefix `scouting-report:`"). Do not duplicate the full
  Dewey integration instructions — keep the centralized
  section as the authoritative reference.

## 2. Synchronize Scaffold Copy

- [x] 2.1 Synchronize the scaffold asset copy at
  `internal/scaffold/assets/opencode/agents/pinkman.md`
  with the updated live agent file.

## 3. Verify Scaffold Drift Detection

- [x] 3.1 Run `go test ./internal/scaffold/...` to
  confirm the scaffold drift detection test passes with
  the synchronized files.
- [x] 3.2 Run `go test ./cmd/unbound-force/...` to
  confirm the file count assertion still passes (no
  files added or removed).

## 4. Constitution Alignment Verification

- [x] 4.1 Verify no hero agent files were modified
  (Composability First -- no mandatory dependencies
  introduced).
- [x] 4.2 Verify no schema registry entries were added
  or modified under `schemas/` (Autonomous Collaboration
  -- no ad-hoc exchange format).
- [x] 4.3 Verify that the Output Formatting, Report
  Persistence, and Recommendation Report sections in
  `pinkman.md` are unchanged — only the Dewey Integration
  section is modified (Observable Quality -- provenance
  metadata preserved).

## 5. Documentation Impact Assessment

- [x] 5.1 Update AGENTS.md "Recent Changes" section with
  a summary of the pinkman-dewey-enrichment change.
- [x] 5.2 Verify no other documentation files require
  updates (README.md, unbound-force.md are unlikely
  to need changes for a convention-only agent update).
  Website documentation gate: exempt — internal agent
  behavioral instruction change with no user-facing CLI,
  workflow, or output format changes.
