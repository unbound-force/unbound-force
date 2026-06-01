<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when: different files, no deps, safe to
  parallelize. Do NOT add [P] when tasks modify the
  same file.
-->

## 1. Severity Pack Enhancement

- [x] 1.1 Add "Compound Severity Escalation" section
  to `.opencode/uf/packs/severity.md` with escalation
  rule (three-part test), example patterns table, and
  anti-consolidation guard

## 2. Adversary Persona Updates

- [x] 2.1 Add dependency necessity check to Audit
  Checklist item 2 in
  `.opencode/agents/divisor-adversary.md` — "Is each
  external tool justified?" before integrity checks
- [x] 2.2 Add content-integrity verification guidance
  to Audit Checklist item 2 — SHA256 vs name-addressed,
  severity depends on execution context
- [x] 2.3 Add CI bot corroboration instruction to
  Audit Checklist item 2 — treat Scorecard/Trivy/GHAS
  findings as corroborating evidence
- [x] 2.4 Add adversarial input enumeration section to
  Audit Checklist (new item 5 between existing items 4
  and old item 5) — per-input threat analysis for each
  new input/parameter/secret. Renumbered subsequent
  items (old 5 → 6, old 6 → 7).

## 3. Review-PR Command Enhancements

- [x] 3.1 Add Step 8e (CI Bot Annotation
  Cross-referencing) to
  `.opencode/commands/review-pr.md` — cross-reference
  Step 7.5b inline comments against Steps 8a–8d
  findings
- [x] 3.2 Add Step 8f (Finding Consolidation) to
  `.opencode/commands/review-pr.md` — group findings
  by root cause, apply compound severity escalation,
  preserve category attribution
- [x] 3.3 Add Step 8g (Severity Calibration) to
  `.opencode/commands/review-pr.md` — for each finding,
  quote matching `severity.md` definition and confirm
  or adjust severity assignment
- [x] 3.4 Add adversarial input enumeration substep to
  Step 8b (Security Review) in
  `.opencode/commands/review-pr.md`
- [x] 3.5 Add issue suggestion gap detection substep
  to Step 8a (Alignment Check) in
  `.opencode/commands/review-pr.md`

## 4. Review-Council Command Enhancements

- [x] 4.1 [P] Add cross-persona finding consolidation
  to Code Review Mode Step 3 in
  `.opencode/commands/review-council.md`
- [x] 4.2 [P] Add cross-persona finding consolidation
  to Spec Review Mode Step 2 in
  `.opencode/commands/review-council.md`

## 5. Constitution Amendment

- [x] 5.1 Add Principle V (Security by Default) to
  `.specify/memory/constitution.md` — supply chain
  integrity, input validation, least privilege,
  dependency necessity
- [x] 5.2 Update constitution version from 1.1.0 to
  1.2.0 and amend date
- [x] 5.3 Update Sync Impact Report comment block at
  top of constitution file
- [x] 5.4 Update Hero Constitution Alignment section
  — note that hero repos need Principle V alignment
  review

## 6. Verification

- [x] 6.1 Verify `severity.md` compound escalation
  section includes both escalation examples and
  anti-consolidation guard
- [x] 6.2 Verify `/review-pr` Steps 8e → 8f → 8g
  are in correct order (CI cross-ref → consolidation
  → calibration)
- [x] 6.3 Verify adversarial input enumeration is
  consistent between `divisor-adversary.md` and
  `review-pr.md` Step 8b
- [x] 6.4 Verify constitution Principle V does not
  contradict existing principles I–IV
