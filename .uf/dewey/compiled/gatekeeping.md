---
tier: validated
compiled_at: 2026-10-25T16:35:32Z
compiled_by: rits/zai-org/glm-5-3
sources:
  - governance-20261014T123131-jay-flowers
  - governance-20261014T151521-jay-flowers
  - governance-20261014T151535-jay-flowers
  - security-gate-delegation-20261014T220250-jay-flowers
  - principle-vi-evidence-20261015T020323-jay-flowers
  - gatekeeping-20261016T195913-jay-flowers
  - gatekeeping-20261016T195917-jay-flowers
  - gatekeeping-20261019T022432-jay-flowers
  - gatekeeping-20261019T111057-jay-flowers
  - gatekeeping-20261019T153511-jay-flowers
  - gatekeeping-20261019T161132-jay-flowers
  - gatekeeping-20261020T170535-jay-flowers
  - gatekeeping-20261020T170544-jay-flowers
  - ci-governance-20261020T170549-jay-flowers
  - coverage-governance-20261022T192847-jay-flowers
  - gatekeeping-20261022T192853-jay-flowers
topic: gatekeeping
---

# Gatekeeping

Compiled 2026-10-25 from 16 learnings (2026-10-14 to 2026-10-22); source: AGENTS.md Behavioral Rules, Gatekeeping.

## Current State (October 2026)

### The Protected Set

- AGENTS.md Gatekeeping rule: MUST NOT modify quality/governance gates — coverage thresholds, CRAP scores,
  severity definitions, CI flags, agent settings, constitution MUST rules, review limits, workflow markers, and pinned
  dependency versions. Violations are CRITICAL severity.
- Agents MUST report and stop when exception conditions are missing or a change would weaken, bypass, or disable a
  gate; the human is the final gate authority (agents report and stop; the human decides).
- Concrete protected values: npm runtime pins (@opencode-ai/plugin), GitHub Action `uses:` SHA pins, coverage floors
  (.opencode/vitest.config.ts, coverage-gate.json), review-matrix values (per_run_timeout_seconds).

### Lockstep Records

- A protected value is a record set, not a number. A pin bump must move every record together: manifests
  (package.json), scaffold mirrors, exact-version test constants (scaffold_test.go), AGENTS.md Technology
  Stack (plus Node engine floor), and mandating spec MUSTs (e.g., SC-FR-003 "Reproducible Plugin
  Dependencies"). The lockfile and its rollback baseline hashes left the record set when
  stop-shipping-plugin-lock stopped shipping the lock; provenance now rides the manifest pins and
  plugin-package-provenance.json.
- Spec mandates are reconciled via a MODIFIED requirement anchored to the unarchived predecessor's text
  (supersession reconciliation on archive). Moving only test constants and AGENTS.md leaves the spec MUST contradicted
  by manifests and testdata; the stale pin flows into baseline specs when the change archives.
- Pins stay exact and immutable — never a mutable tag or name-addressed ref. The process conditions (rationale,
  verification, rollback, validation, commit scope), not the pin format, are the gate.

### Staged Override Posture

- HIGH at spec-review time, pre-disposition (the council verdict is the remaining gate); MEDIUM post-human-override
  with documented authorization (approver/date/rationale/rollback/scope).
- Before accepting an override, verify: (1) human authorization documented in-repo (design/proposal/CHANGELOG plus
  the overridden council verdict); (2) every enforcement gate still enforces at the new value — all records moved
  together; (3) the value remains exact and immutable; (4) rollback plus reproducible validation evidence exist; then
  classify the unsatisfied exception condition MEDIUM, naming PR review and CI approval as completing conditions.
- Re-flagging HIGH post-override contradicts the governance model; silence erases the audit trail. Standing
  overrides make re-review loops waste: the maintainer records REQUEST CHANGES as the TERMINAL verdict for spec and
  code review and proceeds — re-review cannot change the verdict while the override stands. The per-finding human
  decision gate (recommendations before fixing) keeps the override legitimate: every other finding is fixed; only the
  declared item is overridden.
- Maintainer posture: honest declaration plus moved gates, not stealth — declare scope in proposal/design (derived
  from the lock diff, new packages called out), move every record, document rollback, run the local verification suite
  (audit, engine floor).

### Authorized Strengthening Path

- Weakening a gate without documented human authorization is a violation to flag; strengthening through a full
  spec-workflow governance change is the authorized path — noted, not flagged (assess against the archived
  gatekeeping-protection scenarios, archive/2026-04-28-gatekeeping-protection/).
- Exemplar — ts-plugin-coverage-gate (issue #780): plugin statements floor 90→94 in .opencode/vitest.config.ts
  with design decisions D1–D7, a provenance comment naming recorded baselines and the governing issue,
  AGENTS.md/CHANGELOG documentation, and executable proof of both gate modes (passing: 95.18% ≥ 94; failing: CLI
  override --coverage.thresholds.statements=96 exits 1 naming metric, measured value, threshold). The 94 percent floor
  is governance-controlled.
- Floor methodology: set below the measured baseline with headroom for cross-platform instrumentation divergence
  (macOS local vs ubuntu-latest CI — the class that forced the CRAP-load ratchet's 0.5 epsilon); integer floors
  only; re-measure fresh per raise.
- In-session maintainer authorization satisfies traceability once recorded in change artifacts: design decision
  (approver/date/rationale/rollback/scope), proposal Impact entries for the canonical file plus scaffold mirror,
  CHANGELOG note, task record. Durable closure: an explicit maintainer ratification comment on the implementing PR
  before merge. On a ratification request: assess valid and DATA-DRIVEN, recommend ACCEPT plus the comment — not
  revert/re-apply or a split chore change (that applies only with no in-artifact authorization); never perform the
  GitHub mutation or reply in the maintainer's stead.

### Review Hygiene

- Diff changed files against the declared Impact list before approval: undeclared pin bumps are violations even in
  spec review mode (1.4.10→1.19.0 plus transitive cascade rode opsx/clarification-marker-superset with no covering
  task while proposal/design claimed "no new dependencies").
- Workflow `uses:` SHA bumps everywhere require documented rationale, upstream API-verified SHA (CI-002: resolve the
  tag via the API, never from memory), rollback to the previous SHA, and validation evidence — inside the carrying
  change, or as a separate CI-reliability change / dependabot PR under its own commit scope.
- Cross-repo: artifact consumers in other repos' reusable workflows (gaze's reusable_gaze_ai_report.yml downloading
  coverage-out) may run only on push to main, so an upload/download major-version incompatibility is invisible in PR
  CI until post-merge — check every artifact consumer before bumping an artifact-producing action.
- Multi-layer codification (constitution, AGENTS.md rule, pack, skill): pin RFC 2119 classification per rule per
  layer and keep MUST wording identical — a pack-level SHOULD cannot relax a constitution-level MUST, and a
  derived-layer softening qualifier creates citable strictness divergence. Restatement is the mechanism, not a DRY
  violation (each layer loads in a different context); the cost is wording drift, so cross-layer wording diffing is an
  explicit spec-review step.
- Evidence hygiene: compute growth percentages from the same units as the reported counts (a token delta over a word
  baseline inflated 44% where the reported words compute to ~34%).

### Security-Gate Delegation

- Constitution amendments and gate changes route through the constitution-check agent, not ad-hoc edits.
- A parent command delegating to a child agent that owns a security-relevant gate (council spec auto-fix consent)
  must not contradict the child's gate; the fix pattern moves the gate into the parent with the same three fixed
  options plus a session-resume guard invalidating stale consent. Trace every security gate in delegation chains.

## History

- Oct 14 — multi-layer codification ground rules from determinism-codification (per-layer RFC 2119 pinning,
  identical MUST wording, restatement-as-mechanism verified live: the new pack appears in Divisor review prompts with
  zero plugin changes; make plugin-test, 572 tests, no golden-fixture breaks); same day, the delegation-chain
  contradiction (parent "auto-fix them" vs council human confirmation) was fixed in unleash-direct-review-dispatch.
- Oct 16 — timeout ask resolved: per_run_timeout_seconds 600→1200 (PR #745, design decision D11) so a council
  could complete on a large diff; in-session authorization plus the artifact trail resolved the council's HIGH
  finding, and the residual ask — a single on-PR ratification comment — became the durable-closure pattern.
- Oct 19 — the @opencode-ai/plugin 1.4.10→1.19.0 pin-bump case: 02:24 spec-branch hygiene flags the undeclared bump
  as a gatekeeping and commit-scope violation; 11:10 the dual mandate is discovered (test constants AND the SC-FR-003
  spec MUST with rollback lock SHA), requiring MODIFIED-delta reconciliation; 15:35 the post-override posture is
  codified (four verifications, MEDIUM, human as final authority); 16:11 the maintainer records REQUEST CHANGES as
  TERMINAL verdict for spec and code review and proceeds.
- Oct 20 — GitHub Actions pin hygiene: actions/upload-artifact v4.6.2→v7.0.2 rode fix-review-dispatch-recovery (PR
  #776) with no in-change authorization; established that process conditions, not pin format, are the gate.
- Oct 22 — coverage-floor strengthening exemplar: ts-plugin-coverage-gate raises the plugin statements floor 90→94
  (issue #780), establishing strengthening-via-governance-change as noted-not-flagged and headroom-based floor
  methodology.
