<!--
  Dependency order controls execution. [P] marks tasks that may run
  concurrently only when every prerequisite task is complete.
-->

## Execution Checklist

- [x] Step 0: Startup Cleanup
- [x] Step 1: Branch Safety Gate
- [x] Step 2: Resumability Detection
- [x] Step 3: Clarify (Step 1)
- [x] Step 4: Plan (Step 2)
- [x] Step 5: Tasks (Step 3)
- [x] Step 6: Spec Review (Step 4) -- iteration: 8/8
- [x] Step 7: Implement (Step 5) -- phase: 6/6, batch: 0/N,
  workers: 0/N
- [x] Step 8: Code Review (Step 6) -- iteration: 2/3 (APPROVED; HIGH-1 reclassification + LOW-1/LOW-2 resolved)
- [x] Step 9: Retrospective (Step 7)
- [x] Step 10: Demo (Step 8)

## 1. Baseline and Policy Contracts

- [x] 1.1 Record and verify the fullsend commit and seven SHA256
  baselines from `design.md`. Re-read sources and record each target
  adaptation before porting. [RD-FR-001, RD-FR-002, DW-FR-003]
- [x] 1.2 Verify OpenCode prompt types, variant placement, child
  session
  APIs, response parts, cancellation, and usage metadata.
  Keep compatibility types narrow at the current-message response and
  prompt-body boundaries.
  [RD-FR-006]
- [x] 1.3 Define and validate review-matrix version 2 with ordered
  runs, nested profile variants, a closed advisor allowlist, `specs` to
  `spec` normalization, augmentation opt-in, limits, and host fallback.
  Add valid and invalid policy fixtures.
  [RD-FR-001, RD-FR-003, RD-FR-004, RD-FR-005]
- [x] 1.4 Define `.uf/reviewer-capabilities.yaml` and its closed
  version 1 schema. Encode the exact ordered scopes for all six review
  and three content personas. Test eligibility, content exclusion,
  Curator pruning, unknown agents, ordering, and malformed manifests.
  Add an OpenCode-load fixture proving agent provider options are
  unchanged. Scaffold and drift-test the manifest, not agent metadata.
  [RD-FR-002, RC-FR-005, RC-FR-007]
- [x] 1.5 Track and scaffold exact `.opencode/package.json` and lock
  inputs before plugin source exists. Pin plugin 1.4.10, Zod 4.1.8,
  Vitest 5.0.3, and coverage-v8 5.0.3. Verify MIT licenses,
  npm integrity, ignores, staged-probe prerequisites, and rollback.
  Record the tracked baseline lock SHA256 from the design and restore
  the prior `.opencode/.gitignore` state in rollback fixtures.
  [SC-FR-003, SC-FR-004]

## 2. Dispatch Runtime and Artifact

- [x] 2.1 Port the canonical and scaffolded `dispatch-advisor` skill
  and add the `review-dispatch` policy plugin registered in `opencode.json`. Expose
  `plan_review_dispatch` for matrix and reviewer-manifest parsing,
  relevance, total tiering, stable plans, limits, opt-in fan-out, and
  fail-closed full-panel behavior. Commands MUST consume this tool
  rather than reimplement policy.
  [RD-FR-002, RD-FR-003, RD-FR-004, RD-FR-005]
- [x] 2.2 After task 1.5, implement `invoke-agent` with bounded
  schemas,
  manifested review agents, optional model override, timeout,
  cancellation, variant support,
  nullable usage, and redaction. Preserve the selected agent's existing
  OpenCode permissions without adding an override. Test that invocation
  neither broadens nor narrows that existing permission contract. Use
  the reference-compatible first-slash model check and test its
  intentional difference from strict matrix validation. Test that
  a host run resolves the current assistant model and active variant,
  replays them explicitly, and retains null requested provenance plus
  resolved parent and reported child provenance. Fail missing current
  message, model, or variant as sanitized availability without default
  substitution.
  [RD-FR-006]
- [x] 2.3 [P] Add the review-dispatch payload schema, envelope sample,
  closed discriminated run and verdict models, positive and negative
  fixtures, immutable PR and local input contexts, registry validation,
  lossless native-to-generic verdict mappings, exact terminal run
  counts, sum invariants, and same-major compatibility tests. Document
  additive dual emission in registry README, samples, and
  producer-consumer tables. Add `review-verdict` 2.0.0, preserve v1 as
  historical data, and define the canonical consumer migration contract
  implemented by task 3.9. Use hero
  identity `the-divisor`. [RD-FR-008, RV-FR-001, RX-FR-005]
- [x] 2.4 Implement `finalize_review_dispatch` in the policy
  plugin. Validate semantic mappings, terminal states, and run-count
  arithmetic.
  Write artifacts atomically with restricted modes and collision-safe
  correlation ids. Retain immutable refs, SHAs, and human-only
  assessments when persistence fails.
  [RD-FR-008]
- [x] 2.5 Add consolidation fixtures for partial failure, no success,
  duplicate findings, advisory-only results, budget skips, and model
  identity mismatch. Cover every native verdict mapping, availability-
  only and mixed-cause precedence, and semantic rejection of mismatched
  run counts. Run fixtures through schema and semantic validators.
  [RD-FR-007, RD-FR-008, RC-FR-003, RC-FR-004, RC-FR-008]

## 3. Context and Direct Workflows

- [x] 3.1 [P] Add the closed sibling schema, live hero declaration, and
  empty scaffold template. Implement field, URL, origin, clean-state,
  commit, canonical path, symlink, glob, size, count, uniqueness, hash,
  ordering, and untrusted-delimiter checks. Implement local, clone, and
  cache precedence, fetch modes, offline fallback, and unavailable
  behavior. Add positive and negative schema and acquisition fixtures.
  [RX-FR-001, RX-FR-002, RX-FR-003]
- [x] 3.2 Implement the exact delimited JSON lesson wire format after
  sibling evidence contracts. Add its closed schema and fixtures,
  deterministic CRLF grounding and NFC information normalization,
  fixed secret and tool detectors, generated tags, dedupe hashes,
  stable provenance encoding, existing-Dewey learning payloads, bounded
  parent-supplied dedupe identities, duplicate detection, and recorded
  skips.
  [RX-FR-004]
- [x] 3.3 Update canonical and scaffolded `/uf.review-council` for the
  bounded ASCII argument grammar, immutable PR or local input context,
  advisor plans,
  invocation, provenance,
  artifacts, deduplication, verdicts, human fix gate, advisory output,
  no-success behavior, and iterative reruns.
  [RC-FR-001, RC-FR-002, RC-FR-003, RC-FR-004, RC-FR-005,
  RC-FR-006, RC-FR-007, RC-FR-008, RC-FR-009, RX-FR-005]
- [x] 3.4 Update canonical and scaffolded `/uf.triage-issue` for the
  selected persona universe, deterministic title/body/comment
  normalization, versioned length-prefixed framing, keyword rules,
  content hashing, issue tiering,
  two-stage majority, model self-report, artifact output, and
  no-success result. Add rule, ordering, threshold, and boundary
  fixtures. Include fixed SHA256 vectors for the base fixture and exact
  4096, 4097, 32768, and 32769 text-byte boundaries.
  [RD-FR-002, RD-FR-003, DW-FR-001, DW-FR-004, DW-FR-005]
- [x] 3.5 Update canonical and scaffolded `/uf.address-feedback` for
  advised Tier 2 runs, strictest recommendation, Tier 1 preservation,
  model self-report, artifact output, and no-success result.
  [DW-FR-002, DW-FR-004, DW-FR-006]
- [x] 3.6 Update live-only `/speckit.testreview` for testing-only
  plans,
  explicit-first runs, model self-report, artifact output, and
  no-success blocking behavior. [DW-FR-003, DW-FR-004]
- [x] 3.7 Add table-driven command tests for parsing and plan use,
  workflow verdicts, advisory and fix gates, context boundaries,
  lesson proposals, self-reporting, and artifact requirements.
  [RD-FR-007, DW-FR-001, DW-FR-002, DW-FR-003, DW-FR-004,
  RC-FR-001, RC-FR-008, RX-FR-001, RX-FR-002, RX-FR-004]
- [x] 3.8 Update `specs/005-the-divisor-architecture/spec.md` and
  `specs/026-documentation-curation/spec.md`. Record six review and
  three content personas, rerun included plan runs, allow Curator
  pruning outside documentation or user-facing changes, and retain
  `review-verdict` as the canonical downstream artifact.
  [RD-FR-002, RD-FR-008, RC-FR-007, RC-FR-009]
- [x] 3.9 Migrate Mx F, Cobalt-Crush, and Muti-Mind consumers to
  `review-verdict` 2.0.0. Verify approval, advisory, blocking,
  inconclusive, and unavailable decisions plus v1 historical reads and
  major-version rejection by unmigrated consumers. [RV-FR-001]
- [x] 3.10 Update Specs 008 and 009 for the version 2 migration.
  Require orchestration to block `INCONCLUSIVE` and `UNAVAILABLE`.
  Register the five-value enum, historical v1 reads, and
  major-version compatibility.
  Add contract tests for both no-success decisions. [RV-FR-001]

## 4. Reproducible Scaffold Activation

- [x] 4.1 Verify scaffolded manifest and lock assets match task 1.5,
  remain reproducible in fresh targets, and keep both plugin sources
  outside `.opencode/plugins/` until installation and both probes pass.
  [SC-FR-003]
- [x] 4.2 Extend asset prefixes, `uf/` mapping, expected paths,
  directory checks, counts, canonical mapping, and byte-drift tests for
  every new or changed asset. [SC-FR-001, SC-FR-002]
- [x] 4.3 Implement target-directory npm installation with lifecycle
  scripts disabled, Node 20-24 and npm 10-11 prerequisite validation,
  external staging, provider-free probes, atomic source activation, and
  project-directory deployment with explicit `opencode.json` plugin
  registration. Validate anchored ASCII version syntax, CRLF, leading
  zeroes, Unicode, uint32 overflow, an explicit rejected npm
  `v10.11.0` fixture, and extra-line failures. Register both plugins in
  `opencode.json`. Use
  `npm ci --ignore-scripts --omit=dev` in generated targets. Repository
  plugin tests MUST use the full locked development install.
  [SC-FR-004]
- [x] 4.4 Add failure cleanup and idempotent retry. An install or load
  failure leaves no source in `.opencode/plugins/` but retains repairable
  assets. Require exit zero, top-level `partial`, one added failed
  subtool, retained file counts, continued independent subtools, an
  inactive review-plugins result, remediation, and safe retry.
  [SC-FR-004]
- [x] 4.5 Extend `uf doctor` for manifest-lock consistency, dependency
  presence, exact Node/npm parsing, the registration state, plugin
  loads, and repairable versus broken activation states. [SC-FR-005]
- [x] 4.6 Add isolated scaffold and doctor tests for path mapping,
  inventory, drift, working directory, install command, staging,
  exact version boundaries, atomic activation, top-level partial
  result, failure cleanup, independent continuation, and rerun.
  [SC-FR-001, SC-FR-002, SC-FR-003, SC-FR-004, SC-FR-005]

## 5. Measurable Verification

- [x] 5.1 Configure Vitest and automated coverage failure at 90 percent
  statements and 85 percent branches. Unit tests MUST cover matrix and
  reviewer-manifest and plan fixtures, both plugin tools, requests,
  cancellation, redaction, semantic validation, artifact persistence,
  response extraction, current-message host resolution, both narrow
  compatibility boundaries, and all error boundaries.
  [RD-FR-001, RD-FR-002, RD-FR-003, RD-FR-004, RD-FR-006]
- [x] 5.2 Add fake-client integration tests and a provider-free scratch
  repository smoke test for both plugin imports, tool registration,
  planning, finalization, current-message lookup, explicit host replay,
  and child request construction. Cover missing model or variant and
  prove no default lookup occurs. No test may use provider, GitHub,
  Dewey, npm
  registry, or other external network access. [RD-FR-006, SC-FR-004]
- [x] 5.3 Add a checked-in Go coverage-gate manifest and additive Make
  target. Enforce new helper scopes at 80 percent and pure path,
  mapping, config, artifact, and schema scopes at 90 percent. Fail
  missing, multiply matched, or non-numeric values. Derive changed
  production functions from the CI base or local merge base and add an
  omitted-scope negative fixture. Wire the target into `make check` and
  Local CI without changing global 80, backlog 90, or race ratchets.
  [RD-FR-008, SC-FR-001, SC-FR-002, SC-FR-004, SC-FR-005,
  SC-FR-006]
- [x] 5.4 Run schema fixtures, command contracts, asset drift, package
  consistency, plugin unit/integration/smoke tests, and targeted Go
  tests. Verify every requirement has positive and boundary coverage.
- [x] 5.5 Add a `plugin-test` Make target for locked npm install,
  TypeScript unit, integration, smoke, and coverage. Make `check`
  depend on it. Add the same blocking gate to `ci_local.yml` using Node
  22 and an upstream-verified immutable `actions/setup-node` SHA.
  [RD-FR-006, SC-FR-003, SC-FR-004]
- [x] 5.6 Run `gofmt` and `goimports`, then exact CI parity:

  ```bash
  make check
  make plugin-test
  make coverage-gate
  go build ./...
  go test -race -count=1 -coverprofile=coverage.out ./...
  bash council-review-action/test/test-pipeline.sh
  make lint
  make crapload-check
  ```

  Exercise the Local CI workflow path or its repository-supported local
  runner and verify it invokes both new blocking targets.

  Existing CRAP and GazeCRAP MUST not increase. New function CRAP MUST
  remain at or below 30.
- [x] 5.7 Run `openspec validate
  review-council-multi-model-fanout` and verify all requirement IDs map
  to completed tasks with no orphaned asset, schema, or policy.

## 6. Governance and Documentation

- [x] 6.1 Reassess all five constitution principles after delivery,
  including envelope completeness, no-success safety, isolated tests,
  staged activation, and sibling trust boundaries.
- [x] 6.2 Assess and update `CHANGELOG.md`, `README.md`, and
  `AGENTS.md`
  for user-visible behavior, configuration, dependencies, project
  structure, validation commands, and the approved narrow commit-scope
  amendment. Preserve explicit staging and separate unrelated scaffold
  refreshes.
- [x] 6.3 File the required website documentation issue and retain its
  URL as merge evidence. https://github.com/unbound-force/website/issues/295
- [x] 6.4 Assess blog and tutorial value for explicit-first matrices,
  provenance artifacts, and safe sibling context. File a content issue
  when the assessment finds user value. https://github.com/unbound-force/unbound-force/issues/640
- [x] 6.5 Run `/uf.review-council` before PR submission, resolve every
  REQUEST CHANGES result, and make no code changes after approval.

<!-- scaffolded by uf vdev -->

<!-- spec-review: passed -->

<!-- code-review: passed -->
