---
tier: validated
compiled_at: 2026-10-25T16:35:12Z
compiled_by: rits/zai-org/glm-5-3
sources:
  - dependency-governance-20261019T022428-jay-flowers
  - dependency-governance-20261019T022438-jay-flowers
  - dependency-management-20261019T022501-jay-flowers
  - dependency-governance-20261019T111053-jay-flowers
  - dependency-governance-20261019T111059-jay-flowers
  - dependency-governance-20261019T111106-jay-flowers
  - dependency-governance-20261019T111108-jay-flowers
  - dependency-management-20261019T125106-jay-flowers
  - dependency-governance-20261019T171727-jay-flowers
  - dependency-management-20261019T182352-jay-flowers
  - dependency-pins-20261019T182358-jay-flowers
  - dependency-management-20261019T211511-jay-flowers
  - dependency-management-20261019T211518-jay-flowers
  - dependency-management-20261019T211521-jay-flowers
topic: dependency-governance
---

# Dependency Governance

## Current State (October 2026)

### The pin-update exception path

Pinned dependencies (e.g., `.opencode/package.json` plugin-suite pins, currently `@opencode-ai/plugin`
1.19.0) are triple-protected governance — AGENTS.md Gatekeeping, spec MUSTs, and the Commit Scope rule
— not convenience metadata. A pin MAY be updated only for a documented security or CI-reliability
fix, and only when all hold: verified upstream immutable target commit SHA; existing gates remain
enforced; documented rollback to the previous SHA; reproducible validation evidence; normal review
and CI approval. Pins MUST NOT become mutable tags or other name-addressed refs. The exception is
never satisfied by bundling manifest changes into an unrelated feature branch: undeclared changes
to `package.json` or `go.mod` are gatekeeping violations even when the
declared spec work is clean. Reviewers always diff a proposal's declared impact list against the
actual working-tree diff; spec review mode does not exempt dependency changes from scrutiny.

### Lockstep record sets

A pin is a multi-record lockstep set: manifest; scaffold/embedded mirror copies
(inherited by `uf init`); the provenance records (plugin-package-provenance.json); test constants;
AGENTS.md Technology Stack; and the mandating spec requirement text. The tracked lockfile left the
set when stop-shipping-plugin-lock stopped shipping it. A bump that migrates manifest, tests, and
docs but not the
spec record archives a governance MUST contradicting the shipped pin — reconcile the spec
requirement in the same change via a MODIFIED requirement or explicit supersession. Enumeration
claims and upstream-verification lists derive from the throwaway lock the CI scan generates
(`npx -y npm@11.6.2 install --package-lock-only`): newly added transitive
packages plus major-version changes of already-present ones are the supply-chain delta, not just
visible major bumps; license claims come from that generated lock's license field so evidence
cannot contradict the records it certifies. The same lockstep rule governs value corrections
inside a
set (e.g., an engines value): enumerate every surface repeating the value — canonical manifest,
scaffold mirror, README, AGENTS.md, CHANGELOG, in-flight spec delta text, value-pinning tests —
grep the repo for the value string, diff every hit against the authoritative source (the
manifest pin, cross-checked against the generated lock's entry for the binding transitive) and
every other hit, then land all corrections in
one lockstep commit staged by explicit path. Values have no independent pin: `MirrorsMatchCanonical`
enforces only canonical/mirror byte parity, so the same wrong value in both files passes
`make check` — value and assertion must land atomically.

### Staged override posture

When a maintainer overrides a gatekeeping rule with an honestly declared bump, review posture is
staged: HIGH (not CRITICAL) at spec-review time pre-disposition — transparent declaration plus
maintained gates distinguish it from a silent bypass — then MEDIUM after the human override;
standing overrides make re-review loops waste. The council verdict is the remaining gate. Reviewer
duties: (1) transitive surface fully enumerated in the design, (2) rollback documented, (3) all
pin-enforcement gates (drift tests, exact-version assertions, baseline SHA, AGENTS.md Technology
Stack) moved to the new pin by tasks.

### Engine floors

Transitive engine floors are part of the bump. Declare at the manifest level — an `engines` field
mirrored to scaffold assets so `uf init` consumers inherit it — not documentation-only: npm
surfaces engines violations as warnings naming the deep transitive (ini@7.0.0), not the repo's
requirement, giving contributors an unactionable signal.
- The binding constraint is the transitive's disjunctive range (`^22.22.2 || ^24.15.0 || >=26.0.0`
  for ini 7.0.0), declared verbatim in manifest and mirror; a single-value `>=min` floor is a
  strict superset admitting rejected bands (Node 23.x, 25.x, 24.0-24.14).
- Verify the local runtime against the transitive's declared range, never a self-authored manifest
  floor (circular; npm's advisory enforcement catches neither).
- CI `node-version: 22` resolves to v22.23.3 (in range, verified 2026-10-19); older Node sees
  EBADENGINE warnings, not hard failures.
- The engines value requires an exact-value assertion in
  `TestPluginPackageInputs_ExactVersions` (internal/scaffold), the same way
  dependency versions are pinned; no lockfile ships anymore (stop-shipping-plugin-lock),
  so the floor is verified against the live runtime, not a lock.

### Workflow SHA discipline

Workflow SHA bumps are constitutional gate modifications: API-verified SHA, documented rollback,
validation evidence — or a separate CI-reliability change. The retained actions/upload-artifact
v7.0.2 bump is now properly SHA-pinned in `ci_local.yml`.

## History

### The 1.19.0 adoption case

The @opencode-ai/plugin 1.4.10 -> 1.19.0 bump (opsx/clarification-marker-superset, 2026-10-19)
began as a flagged violation: a diff bundling dependency changes with unrelated spec artifacts was
flagged as commit-scope violation, and the pin was known risky — this toolchain broke consumers on
unvalidated jumps before (council-review-action pins opencode-ai because 1.16.0+ removed the
`part.text` JSONL field its parser depends on), and the bump drags the effect beta channel and
ini 6->7. It was then maintainer-authorized (design D7) and honestly declared: staged override
posture applied (HIGH at spec-review time), council verdict the remaining gate. Reconciliation
completed: SC-FR-003 ("Reproducible Plugin Dependencies", which had pinned 1.4.10 with a recorded
baseline lock SHA256) carries a SUPERSEDED-BY note attributing the update; dedicated commit
02b593e2; manifests, tests, and AGENTS.md all at 1.19.0. Review of the declared enumeration caught
a false claim: D7 named only ini/msgpackr/effect while the lock diff also introduced
@ai-sdk/provider 3.0.8 and json-schema 0.4.0 and refreshed @opencode-ai/sdk, fast-check,
msgpackr-extract, multipasta, pure-rand, and yaml — establishing the derive-enumeration-from-
lock-diff rule.

### Engines-field landing

Initial handling documented the ini 7 floor in AGENTS.md only — an engine watch-item with no owning
task is an unmitigated environment regression. The manifest `engines` field then went through
approximation cycles: `>=22.22.2` was a superset of ini 7's range, appearing on five-plus surfaces
with three different values for one constraint inside a single change (PR 761,
fix-triage-issue-dispatch-docs), while the same PR's AGENTS.md edit claimed the full disjunctive
range was declared. The final state declares the verbatim OR-range in manifest and scaffold mirror,
with the exact-value test assertion landing in the same lockstep commit as the corrected value.

### upload-artifact retention

The actions/upload-artifact v7.0.2 bump in ci_local.yml was retained and is now properly
SHA-pinned, satisfying the constitutional requirement for API-verified SHAs with documented
rollback and validation evidence.
