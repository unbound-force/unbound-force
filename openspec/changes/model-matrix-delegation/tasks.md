<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
  Phase checkpoints (MP-018): every phase below is terminated
  by the /uf.unleash implement-phase checkpoint — a hard-gate
  pre-flight run of the full detected CI/local suite
  (.opencode/commands/uf.unleash.md Step 5) — and the next
  phase MUST NOT begin while that checkpoint fails. Tasks
  whose text carries an explicit verification command express
  the same gate in-plan.
-->

## 1. Meta-prompt convention pack (#888)

- [x] 1.1 Author `.opencode/uf/packs/meta-prompt.md`: MP-NNN
  rule ids, `[MUST]/[SHOULD]/[MAY]` tags, matching the format
  and voice of `go.md`, `content.md`, and `determinism.md`.
  Content is consolidation, not invention: RFC 2119 language,
  FR-NNN requirement numbering, Given/When/Then scenarios,
  line length < 72; Speckit vs OpenSpec artifact shapes and the
  OpenSpec delta format (`## ADDED Requirements`, requirement
  numbering allocation headers); `plan.md`/`tasks.md`
  decomposition conventions (`[P]` markers, checkbox discipline,
  phase numbering); command-file determinism-first framing;
  clarification-marker grammar awareness described WITHOUT any
  literal, well-formed marker syntax appearing in the prose.
  Run `scan_clarifications`-equivalent verification (a marker
  scan over the pack file) and confirm zero extracted
  questions. This task precedes the tracked matrix (task 4.1)
  because the starter `packs:` mapping names `meta-prompt` and
  pack-existence validation (MM-FR-001) requires the file.
  Landing the pack also carries its two test-registry ripples
  (design D9 records the exemption decisions):
  `internal/schemas/packvalidator_test.go` (skip meta-prompt.md
  in TestValidateConventionPack_AllPacksValid, mirroring the
  review-voice.md precedent) and
  `internal/scaffold/scaffold_test.go` (knownNonEmbeddedFiles
  entry, mirroring the determinism.md governance precedent).

## 2. Model-matrix lib module — schema, loader, override, precedence

- [x] 2.1 Create `.opencode/lib/model-matrix.ts` (design D1):
  closed Zod schema for the v1 surface (`version`, `profiles`
  with optional `model`/`variant`, `packs`, `defaults` with the
  closed task-kind key set, `agents`, `classifier` structure
  only — unknown keys rejected everywhere); `loadModelMatrix`
  with an injectable `dependencies.readText` (the `loadPolicies`
  pattern) returning a discriminated
  `{status: "absent" | "ok" | "invalid", ...}` result; pack
  existence validation against `.opencode/uf/packs/` via an
  injected `fileExists`; and the `SelectionResult` type
  (`selection_source`, pack, capability, model, variant,
  confidence, fallback reason).
- [x] 2.2 In the same module, implement the override handling
  (design D4): `validateModelMatrixOverride` constrained to the
  fail-closed whitelist `profiles.*.model`,
  `profiles.*.variant`, `classifier.llm.*` (any nested key under
  `classifier.llm`; everything else under `classifier` is
  structural and rejected); `mergeOverride` nested-path merge
  with null-model propagation; merged-result re-validation
  against the identical schema; empty and comments-only override
  no-ops; error messages name the whitelist.
- [x] 2.3 In the same module, implement `resolveSelection`
  (design D2, MM-FR-003): the fixed precedence chain — explicit
  `model` → explicit `profile` → explicit `pack` (validated:
  pack file exists AND is a `packs` key) → `agents:` pin →
  static `defaults:` by `checkpoint_step` task kind
  (`plan`/`tasks` static; `implement`/unknown to the
  classifier) → classifier (injected callback) → host. A pure
  function of (matrix, call inputs, classifier result) — no
  I/O. Unknown explicit profile, invalid explicit pack, or
  explicit profile/pack with no matrix configured fail closed
  with non-retryable errors naming the remediation.
- [x] 2.4 In `.opencode/lib/model-matrix.ts`, enforce the
  tracked-role population rule (MM-FR-001) in `loadModelMatrix`:
  a load of the tracked `.uf/model-matrix.yaml` whose
  `classifier.llm` carries any populated value (a non-empty
  `endpoint`, `model`, `apiKey`, or `request_timeout_ms` entry)
  MUST return `status: "invalid"` with an error naming
  `.uf/model-matrix.override.yaml` as the remediation. The
  check runs against the tracked file BEFORE the override merge
  (merged views legitimately populate those paths, MM-FR-002),
  separately from Zod schema validation, and the
  empty/comments-only no-op file stays valid (MM-FR-004). Add
  the loader rejection rows to
  `.opencode/test/model-matrix.test.ts`: tracked populated
  endpoint/model/apiKey/request_timeout_ms each fail load
  naming the override file, while the same values arriving via
  the override still load. Not [P]: extends the task 2.1
  module and its test file (MP-017).
- [x] 2.5 Reconcile `.opencode/test/model-matrix.test.ts` with
  the plan of record: the file was created with tasks 2.1-2.3
  and already carries most of the enumerated coverage — this
  task verifies every enumerated case exists, corrects the
  file-header task attribution to name task 2.5 (the classifier
  tests live in task 3.3's separate file), and adds the missing
  rows: boundary rows for the profile-count bound (0-profile
  and 33-profile matrices rejected; 1 and 32 accepted),
  tracked-role credential-population rejection (a tracked file
  with populated `classifier.llm` values fails load naming the
  override file), the comments-only tracked file as a valid
  no-op, plus the enumerated coverage — schema closed-input
  rejection (unknown root/nested keys, unknown `defaults` kind,
  `packs` key naming a nonexistent pack, `packs` value naming
  an absent profile); loader absent/ok/invalid tristate;
  override whitelist violations fail closed naming the
  whitelist (structural `classifier.enabled` rejected,
  `classifier.llm.*` accepted); null-model propagation;
  empty/comments-only override no-op; schema bound rows for
  the agent-pin count (0 and 33 pins rejected, 32 accepted),
  `agent_timeout_ms` above 1_800_000 rejected, and
  `classifier.llm.request_timeout_ms` above 180_000 rejected;
  precedence order scenarios incl. explicit-model-wins,
  profile-beats-pack with the supplied pack retained for
  directive rendering, explicit-model-with-valid-pack (pack
  retained), agent-pin-beats-defaults, plan/tasks static
  resolution, classifier-wins-over-defaults.implement,
  defaults.implement-as-fallback, profile-without-model → host,
  unknown profile fail-closed, explicit profile with no matrix
  fail-closed, and the unconditional-validation combined rows —
  model+invalid-pack, model+unknown-profile, profile+invalid-
  pack, matrix-less model+pack, matrix-less model+profile
  (MM-FR-001 through MM-FR-004 scenarios). Run the
  file's tests via `make plugin-test`.
- [x] 2.6 Add the JSON-Schema cross-validation harness
  `.opencode/test/model-matrix-schema.test.ts`: a test that
  validates every `schemas/model-matrix/
  samples/` fixture against `v1.schema.json` AND against the
  Zod schema, asserting identical verdicts (design D3). The
  fixtures themselves land with task 4.2; this harness may land
  first with the valid-matrix fixture only. Not [P]: it
  cross-validates against the Zod schema authored by task 2.1
  in the same phase (MP-017 rule (b)).

## 3. Classifier cascade

- [x] 3.1 Create `.opencode/agents/model-matrix-classifier.md`:
  frontmatter `name: model-matrix-classifier`, `mode:
  subagent`, `hidden: true`, description "Internal to the
  model-matrix selection layer. Investigates the task context
  and responds with a JSON pack classification.", and a
  least-privilege permission block (`edit: deny`, `bash:
  deny`, `webfetch: deny` — the divisor reviewer shape,
  MM-FR-005). Body mandates
  the investigate-then-decide protocol: read the feature dir's
  spec/plan/tasks and relevant code, then respond with ONLY
  `{"pack": "<name>", "confidence": <0.0-1.0>, "reasoning":
  "<optional>"}` — no markdown fences, no prose, never ask
  questions (MM-FR-005 agent-engine contract).
- [x] 3.2 In `.opencode/lib/model-matrix.ts`, implement the
  classifier cascade (design D5): an agent engine that spawns
  the hidden subagent through an injected session-spawn seam
  (reusing the executor path), bounded by `agent_timeout_ms`
  (default 45_000); an LLM engine with an injectable `fetch`
  issuing the chat-completions POST (`temperature: 0`,
  `response_format: {type: "json_object"}` with one 400-retry
  without the field, AbortSignal request timeout
  (`classifier.llm.request_timeout_ms`, default 10_000), a
  50 MiB response-body cap (an oversized body is malformed —
  skipped unbuffered, cascade falls back statically,
  MM-FR-006), apiKey
  interpolated from `{env:VAR}` with literal-key rejection and
  unset-variable skip (engine skipped, never an empty
  credential), https-or-loopback endpoint validation,
  endpoint/model/apiKey read only from the override's
  `classifier.llm.*`, unconfigured = skipped); confidence
  threshold gating (default 0.6); matrix-validity check of the
  returned pack; first valid above-threshold result wins;
  below-threshold/timeout/error/malformed (including
  fences-wrapped JSON) → next engine or static fallback; and
  the in-process result cache keyed by SHA-256 of
  (feature_dir + prompt + agent), capped at a 128-entry named
  constant, cache hit skips both engines, instantiated per
  cascade instance (never module-level state) and injectable
  for tests. Classifier failure
  never throws to the caller.
- [x] 3.3 Create `.opencode/test/model-matrix-classifier.test.ts`
  unit tests with mocked seams: agent engine success (valid
  JSON, above threshold, matrix-valid pack) and
  `selection_source: "classifier-agent"`; LLM engine success →
  `classifier-llm`; below-threshold fallback; agent timeout →
  LLM engine; both engines fail → static default, no error to
  caller; malformed and fences-wrapped JSON rejected;
  oversized response body (above 50 MiB) rejected unbuffered
  with static fallback; matrix-invalid pack rejected; `{env:VAR}` interpolation and
  literal-key rejection; 400-retry without `response_format`;
  unconfigured `classifier.llm` skips the network entirely;
  cache hit skips engines; cache cap eviction — cache tests
  construct isolated cascade instances sharing no module-level
  state (TC-002/TC-004) (MM-FR-005/006
  scenarios).

## 4. Tracked matrix and schema registry

- [x] 4.1 Create the tracked `.uf/model-matrix.yaml` per design
  D9: `version: 1`; the four starter capabilities with empty
  profiles (comment noting identities arrive via the override);
  `packs` mapping `meta-prompt`/`go`/`typescript`/`python`/
  `content`; `defaults` mapping `plan`/`tasks` →
  spec-authoring and `implement` → code-writing; empty
  `agents`; `classifier` structure only. Verify it loads through
  `loadModelMatrix` (requires task 1.1's pack file) and assert
  the tracked-file conformance rule — `defaults` maps `plan`
  and `tasks` to `spec-authoring` (MM-FR-001 starter-mapping
  scenario; a drifted mapping fails the check while passing
  schema validation).
- [x] 4.2 Create the canonical registry `schemas/model-matrix/`
  (design D3, MM-FR-008): `v1.schema.json` (closed input,
  `additionalProperties: false` throughout, matching the Zod
  surface), `README.md` documenting the surface, the override
  whitelist, the version history, and the `mechanical`
  capability's explicit-profile-only reachability (no implicit
  selection path in the starter matrix), and `samples/`
  fixtures — valid starter matrix, valid minimal matrix, and
  invalid fixtures (unknown root field, unknown `defaults`
  kind, nonexistent pack, absent profile value, 33-profile
  count violation, whitelist-violating override with a
  structural key, whitelist-violating override with
  `classifier.enabled`) plus one loader-rejection sample — a
  tracked matrix with populated `classifier.llm` values —
  which is schema-VALID by design (the tracked-role rule is a
  loader check, MM-FR-001, not a schema rule): the harness
  asserts it passes `v1.schema.json` and fails
  `loadModelMatrix`. All fixtures must pass the task 2.6
  cross-validation harness.
- [x] 4.3 Deploy the byte-identical registry copy at
  `.uf/schemas/model-matrix/` and the embedded scaffold mirror
  at `internal/scaffold/assets/uf/schemas/model-matrix/`
  (verify with `diff -rq`). Not [P]: deploys content created by
  task 4.2 in the same phase (MP-017 rule (b)).
- [x] 4.4 [P] Add `.uf/model-matrix.override.yaml` to the root
  `.gitignore` under the Unbound Force section.

## 5. Delegation-tool integration

- [x] 5.1 In `.opencode/plugins/uf-workflow/index.ts`, extend
  `DelegateCommandInputSchema` and the tool args with
  `model` (provider/model identity, non-empty), `variant`,
  `profile` (1–128 chars), and `pack` (1–128 chars) — the
  validated shape is defined by the unleash delta (UF-FR-002);
  this schema mirrors it —
  each with descriptions documenting the precedence order; add
  the mutual-independence refinement notes (selection params
  are orthogonal to classification params).
- [x] 5.2 Wire selection into `createDelegateCommandTool`'s
  `execute()` (design D2): after skill resolution and BEFORE
  host introspection, probe for the matrix via the injected
  filesystem dependency (zero-config short-circuit, design
  D8); resolve the selection through `resolveSelection` with
  the classifier cascade for `implement`/unknown kinds; on a
  resolved matrix model, skip `resolveHostModel()` entirely;
  on host fallback, preserve the existing variant-fallback and
  structural-failure semantics verbatim. Pack/profile
  validation failures return non-retryable envelopes at the
  pre-spawn stage (zero child sessions).
- [x] 5.3 Add `renderPackDirective(pack)` — the fixed-template,
  tool-rendered instruction to load
  `.opencode/uf/packs/<pack>.md` before writing (never inlined
  content) — and insert it in prompt composition at the fixed
  position: command text → feature-dir context → skill
  directive → pack directive → extra prompt (MM-FR-005
  double duty; unleash delta).
- [x] 5.4 Record the audit trail (design D7, MM-FR-007):
  `selection_source`, selected pack, resolved capability,
  resolved model, and classifier confidence (when applicable)
  in `DelegateCommandData` and the checkpoint data — additive
  fields only; zero-config calls record `selection_source:
  "host"` and keep the pre-change envelope shape.
- [x] 5.5 Extend `.opencode/test/unleash-delegate.test.ts` and
  add `.opencode/test/model-matrix-integration.test.ts`:
  integration test proving a delegated task runs on a non-host
  model configured via matrix + override with a mocked child
  session (AC1); plan/tasks → spec-authoring resolution and
  implement/unknown → classifier with static fallback (AC2);
  explicit model/profile/pack override precedence (AC3);
  missing matrix → behavior identical to today, host model, no
  errors (AC4); non-whitelisted override key fails closed
  naming the whitelist (AC5); golden-shape test asserting the
  exact pack-directive rendering and position, skill+pack
  coexistence order (AC6); classifier failure/timeout/low
  confidence never blocks, falls back statically, audit
  records it (AC7); every selection records `selection_source`
  and the resolved model (AC8); matrix-model resolution
  proceeds when host introspection would fail (no
  `HOST_MODEL_UNAVAILABLE`); input-shape boundary rows per
  UF-FR-002 — `profile`/`pack` at exactly 1 and 128 chars
  accepted, empty-string and 129-char values rejected
  non-retryably pre-spawn.
- [x] 5.6 Seam verification (DE-007): write a test proving the
  three layer surfaces stay in sync — (a) the registration
  surface: `createDelegateCommandTool` advertises
  model/profile/pack in its args and `DelegateCommandInputSchema`
  accepts them; (b) the documented output shape: the audit
  fields recorded by task 5.4 appear in the checkpoint data
  exactly as documented in MM-FR-007; (c) write→read round-trips:
  a matrix written by `schemas/model-matrix/samples/`-shaped
  content loads through `loadModelMatrix` and round-trips its
  selected capability into `resolveSelection` output.

## 6. Scaffold registration and mirrors

- [x] 6.1 In `internal/scaffold/scaffold.go`, add
  `opencode/lib/model-matrix.ts` to both lib asset enumeration
  lists (the ~line 660 and ~line 733 blocks) and append
  `.uf/model-matrix.override.yaml` to `gitignoreBlock` (SC-FR-007).
- [x] 6.2 Create the embedded scaffold mirror
  `internal/scaffold/assets/opencode/lib/model-matrix.ts`
  byte-identical to the canonical lib module (REQUIRED — the
  mirrored plugin imports it; design D1), and sync the modified
  `.opencode/plugins/uf-workflow/index.ts` byte-identically to
  `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`.
- [x] 6.3 Create the embedded template
  `internal/scaffold/assets/uf/model-matrix.yaml` (the tracked
  starter matrix) and confirm the task 4.3 schema mirror; both
  must be drift-enforced. Extend the canonical-assets drift
  table/test (`TestEmbeddedAssets_MatchSource` and the
  `task42CanonicalAssets` enumeration) to cover the new lib
  file, the matrix template, and the schema registry entries.
- [x] 6.4 Update `.opencode/commands/uf.unleash.md` with
  the UF-FR-001 guidance — pass `pack` on implement delegation
  when the task clearly names one; the classifier covers the
  rest; document the selection parameter contract and
  precedence order; never instruct reconstructing pack content —
  and sync the command file byte-identically to its scaffold
  mirror under `internal/scaffold/assets/opencode/commands/`.
- [x] 6.5 Run the scaffold drift tests and `go test
  ./internal/scaffold/...`; fix any mirror or enumeration
  mismatch.
- [x] 6.6 Add the UF-FR-001 prose pin to
  `internal/scaffold/scaffold_test.go`: a between()-span
  assertion over `.opencode/commands/uf.unleash.md` (and its
  mirror via the drift test) verifying the model-selection
  guidance span exists verbatim — recommend `pack` when the
  task clearly names one, the classifier covers the rest, the
  selection parameter contract and precedence order
  documented, no pack-content reconstruction — so removing the
  guidance fails the suite (unleash delta, UF-FR-001 "Guidance
  span is pin-tested" scenario).

## 7. Documentation

- [ ] 7.1 [P] Update `docs/configuration.md`: add a
  "Delegation Model Matrix" section (or extend the Review
  Council Configuration pattern) documenting
  `.uf/model-matrix.yaml` and `.uf/model-matrix.override.yaml`
  — capability profiles, the `packs` mapping, `defaults`,
  `agents` pins, the classifier (agent + LLM engines,
  confidence threshold), the precedence chain, the override
  whitelist, and the gitignore semantics (AC10).
- [ ] 7.2 Add the `CHANGELOG.md` entry under `Unreleased →
  Added`: capability-driven model selection for
  `uf_delegate_command`, the config pair, the classifier
  cascade, the pack directive, the audit trail, and the
  meta-prompt pack. The entry footer MUST carry the verification
  result: the `.opencode/` plugin suite's statements-coverage
  figure under the unchanged governance-controlled 94 percent
  floor (from task 8.1) and the constitution-alignment outcome
  recorded by task 8.2 — follow the established footer
  convention of prior entries (plugin-suite result under the
  unchanged 94 percent statements coverage floor).
- [x] 7.3 File the documentation issue for this user-facing
  change against the current repo before PR merge (AGENTS.md
  documentation gate); record the issue number in the CHANGELOG
  entry or the change notes, and record the bookkeeping in
  `proposal.md` (top comment) and the PR body — the PR closes
  #887 (model matrix), #888 (meta-prompt pack), and the filed
  documentation issue, via `Closes #887`, `Closes #888`, and
  `Closes #<doc-issue>` (all three recorded in the proposal
  top comment).
- [ ] 7.4 Update `AGENTS.md` (structure section): the new
  `.uf/model-matrix.yaml` + `.uf/model-matrix.override.yaml`
  config pair and `schemas/model-matrix/` registry in the
  project structure tree; `.opencode/agents/` count 16 → 17
  (model-matrix-classifier, hidden); `.opencode/` shared library
  line naming `model-matrix.ts`; and a Convention Packs list
  entry for `meta-prompt.md` (repo-local governance pack, no
  `-custom.md` companion — the determinism.md precedent). Keep
  the edits structural (structure,
  counts, lists) — no behavioral-rule changes.
- [ ] 7.5 File the cross-repo documentation issue against
  `unbound-force/website` covering the model matrix: the config
  pair, capability profiles, precedence chain, classifier
  engines, override whitelist, and credential story — per the
  org constitution v1.4.0 Cross-Repo Documentation rule
  (user-facing hero-repo features get website docs; PR #744
  precedent). Include the tutorial disposition in the issue:
  the model matrix is configuration reference, not tutorial
  material — it MAY be cited from existing delegation docs but
  warrants no new tutorial below the multi-step feature
  threshold. Record the issue link in the change notes.

## 8. Verification

- [ ] 8.1 Run `make check` (lint, Go tests, build,
  coverage-gate, plugin-test) and confirm the `.opencode/`
  plugin suite stays at or above the governance-controlled 94
  percent statements coverage floor — CI parity with
  `ci_local.yml` (AC9).
- [ ] 8.2 Verify constitution alignment and record the outcome
  in the CHANGELOG entry from task 7.2: standalone injectable
  lib module (II, IV), `selection_source` provenance on every
  delegation (I, III), credentials only in the gitignored
  override with `{env:VAR}` interpolation and a fail-closed
  whitelist (V), deterministic precedence with the classifier
  bounded/cached/overridable and the pack directive
  tool-rendered (VI), and every behavior testable without
  external services (IV).
- [ ] 8.3 Verify gatekeeping: `git diff` shows no coverage
  floor, CRAP, severity, CI-flag, review-limit, or
  pinned-dependency changes; the review-matrix trio,
  `agent-executor.ts`, and `resolveHostModel` are value-identical
  to the base branch. Stop and report per the gatekeeping rule
  if any check fails.
- [ ] 8.4 Verify the acceptance criteria walkthrough against
  issue #887's ten criteria (AC1–AC10), citing the test or
  artifact that proves each; record gaps as findings, not
  silent omissions.
