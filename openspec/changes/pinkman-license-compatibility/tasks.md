## 1. Add Compatibility Tier Classification

- [x] 1.1 In `.opencode/agents/pinkman.md`, add a
  "License Compatibility" section after "License
  Classification". Define the three-tier mapping table
  (permissive, weak-copyleft, strong-copyleft) with all
  SPDX identifiers per the delta spec. Include the
  `unknown` tier for unrecognized licenses. Add a step
  to the License Classification procedure: "After
  assigning the OSI verdict, assign a compatibility
  tier from the License Compatibility table."
- [x] 1.2 Add the compatibility verdict mapping
  (compatible, caution, incompatible) based on the tier.
  Include the rules for non-OSI (`incompatible`) and
  manual_review (`caution`) verdicts.
- [x] 1.3 Add the dual-license compatibility rule:
  evaluate each option independently, use the most
  favorable tier. Include the tier ordering
  (permissive > weak-copyleft > strong-copyleft >
  unknown).

## 2. Gate Recommendation Verdict

- [x] 2.1 In the Report Mode "Recommendation verdict"
  step (current step 9), add the compatibility gate:
  `compatible` allows any recommendation, `caution`
  caps at `evaluate`, `incompatible` forces `avoid`.
  Update each verdict bullet to reference the
  compatibility gate.
- [x] 2.2 Update the `avoid` verdict to include
  "incompatible license compatibility verdict" as a
  trigger alongside the existing "License is not
  OSI-approved" trigger.

## 3. Update Fallback License List

- [x] 3.1 Annotate each license in the fallback list
  with its compatibility tier in parentheses (e.g.,
  `MIT (permissive)`, `GPL-3.0-only (strong-copyleft)`,
  `LGPL-3.0-only (weak-copyleft)`). Group by tier for
  readability.

## 4. Update Output Formats

- [x] 4.1 In the Discover/Trend Result List template,
  add `- **Compatibility**: <tier> (<verdict>)` after
  the License line in each project entry.
- [x] 4.2 In the Audit Result Table template, add a
  `Compatibility` column after `License Changed?`.
- [x] 4.3 In the Recommendation Report template, add
  `- **Compatibility**: <tier> (<verdict>)` in the
  License Analysis section.

## 5. Update Dewey Learning Templates

- [x] 5.1 In the Dewey Integration "After Scouting"
  structured prose templates, update each mode's
  required content to include the compatibility verdict
  per project (e.g., "testify (MIT,
  permissive/compatible, adopt)").

## 6. Synchronize Scaffold Copy

- [x] 6.1 Synchronize the scaffold asset copy at
  `internal/scaffold/assets/opencode/agents/pinkman.md`
  with the updated live agent file.

## 7. Verify Scaffold Drift Detection

- [x] 7.1 Run `go test ./internal/scaffold/...` to
  confirm the scaffold drift detection test passes with
  the synchronized files.
- [x] 7.2 Run `go test ./cmd/unbound-force/...` to
  confirm the file count assertion still passes (no
  files added or removed).

## 8. Constitution Alignment Verification

- [x] 8.1 Verify no hero agent files were modified
  (Composability First).
- [x] 8.2 Verify no schema registry entries were added
  or modified under `schemas/` (Autonomous
  Collaboration).
- [x] 8.3 Verify that the Report Persistence section
  and its YAML frontmatter schema are unchanged
  (Observable Quality — provenance metadata preserved).

## 9. Documentation Impact Assessment

- [x] 9.1 Update AGENTS.md "Recent Changes" section
  with a summary of the pinkman-license-compatibility
  change.
- [x] 9.2 Verify no other documentation files require
  updates. Website documentation gate: exempt —
  internal agent behavioral instruction change with no
  user-facing CLI, workflow, or output format changes.
