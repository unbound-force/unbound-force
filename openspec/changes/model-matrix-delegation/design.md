## Context

`uf_delegate_command` (`.opencode/plugins/uf-workflow/index.ts`,
factory `createDelegateCommandTool`) is the single delegation path
`/uf.unleash` and `/uf.run-pipeline-plan` use for content-agent
work. Today the child model ALWAYS comes from `resolveHostModel()`
(`.opencode/lib/agent-executor.ts`) — parent-session
introspection of the current assistant message; failure aborts
with `HOST_MODEL_UNAVAILABLE`. Meanwhile `executeAgentSession`
already accepts any `ModelIdentity {providerID, modelID,
combined}` plus optional `variant` — review agents run on
non-host models today via `invoke_agent`. Only the selection
layer is missing.

The review council already established the config pattern this
change mirrors: `.uf/review-matrix.yaml` (tracked, tier
structure), `.uf/review-matrix.override.yaml` (gitignored,
per-developer model identities, whitelisted keys only,
`validateOverride()`/`mergeOverride()` precedent in
`.opencode/plugins/review-dispatch/index.ts`), and a JSON Schema
registry maintained as three byte-identical copies (root
`schemas/`, deployed `.uf/schemas/`, embedded scaffold mirror).
Stakeholder intent (issue #887): there is a large capability gap
between models good at meta-prompt engineering and models good
at writing code; the matrix makes that gap addressable per
developer without code changes.

Signals available in-tool at selection time: `command_path`,
`checkpoint_step`, `phase`, `task_classification`, `agent`,
`feature_dir`, `skill`, and the composed prompt. Delegation call
sites: `.opencode/commands/uf.unleash.md` and
`.opencode/commands/uf.run-pipeline-plan.md` only. No
"model-matrix" references exist in `.opencode/`, `.uf/`, or
`internal/` — the namespace is clean.

This design covers issue #887's contract: the tracked matrix +
override pair (MM-FR-001/002), the precedence chain
(MM-FR-003), zero-config no-op (MM-FR-004), the classifier
cascade and LLM engine (MM-FR-005/006), the audit trail
(MM-FR-007), the schema registry (MM-FR-008), the
delegation-tool integration (unleash delta), and the scaffold
asset surface (SC-FR-007). The meta-prompt pack (#888) is
authored inside this change as a distinct task.

## Goals / Non-Goals

### Goals

- Spec/plan/tasks authoring work can run on a different model
  than code implementation, per configuration, with capability
  categories expressible without code changes.
- Deterministic, auditable selection: a fixed precedence chain
  evaluated by tool-layer code, `selection_source` provenance in
  every envelope and checkpoint record.
- Per-developer model identities in a gitignored override with a
  fail-closed whitelist; tracked config carries structure only.
- Zero-config unchanged: no matrix present means host model, no
  errors, byte-identical call behavior (envelope extended
  additively).
- Classifier auto-derivation for `implement` and unknown kinds
  that never blocks delegation and falls back statically.
- The `pack` parameter doing double duty (model selection + a
  tool-rendered pack directive), golden-shape tested.
- Scaffold mirrors synced so the mirrored plugin loads in
  scaffolded hero repos and drift tests stay authoritative.
- The meta-prompt convention pack (#888) exists so the starter
  `packs:` mapping validates.

### Non-Goals

- Hero-repo deployment via `uf init` (pack templates,
  `alwaysDeployedPacks` entries, per-repo matrix provisioning) —
  slice 2, issue #889.
- Modifying the review-matrix trio, `resolveHostModel`, or
  `executeAgentSession` — the executor already supports any
  model; host introspection remains the fallback path.
- Reviewer-agent selection: `invoke_agent`/review-dispatch are
  untouched (the review matrix governs reviewers; the model
  matrix governs content agents).
- A multi-file merge strategy beyond the single nested-path
  override, environment-variable or CLI override mechanisms, or
  matrix hot-reload — the matrix is read per delegation call,
  like `loadPolicies`.
- Teaching the classifier to ask questions or return non-pack
  capabilities — its output contract is a single pack name with
  confidence.

## Decisions

### D1: Selection logic lives in a new `.opencode/lib/model-matrix.ts`

The loader (tracked matrix + override + merge), Zod schema,
precedence resolution, pack validation, classifier engines, and
the result cache live in one shared lib module with an
injectable `dependencies.readText` (the `loadPolicies` pattern).
The uf-workflow plugin imports it; `agent-executor.ts` is
untouched. Rationale per Constitution II/IV: standalone,
testable without the plugin, one source of truth for the plugin
and its scaffold mirror. The mirrored plugin imports the
mirrored lib — the lib mirror under
`internal/scaffold/assets/opencode/lib/` is REQUIRED, not
optional (SC-FR-007).

### D2: Selection hooks in before host introspection; validation is pre-spawn

Inside `createDelegateCommandTool`'s `execute()`, selection runs
after input validation and skill resolution and BEFORE
`resolveHostModel()`. `profile`/`pack` validation failures
(unknown profile, pack file missing, no matrix entry) fail
closed at the same pre-spawn stage as `SKILL_NOT_FOUND` — a
rejected call costs one tool round-trip, zero child sessions.
When the chain resolves a matrix profile model, host
introspection is skipped entirely, so delegations that would
have failed `HOST_MODEL_UNAVAILABLE` now proceed — the
side effect issue #887 explicitly preserves. When the chain
falls through to host, the existing variant-fallback and
structural-failure semantics are preserved verbatim (the
delegation-skill-enforcement contract).

Prompt composition gains exactly one section, at a fixed
position: command text → feature-dir context → skill directive →
pack directive → extra prompt.

### D3: Runtime Zod schema mirrors the canonical JSON Schema

The matrix surface is enforced twice, like review-matrix: a Zod
schema in the lib module (runtime, closed input, unknown keys
rejected) and the canonical `schemas/model-matrix/v1.schema.json`
(`additionalProperties: false` throughout). Sample fixtures are
cross-validated against both in tests, so the two cannot drift.
Adaptation recorded: issue #887's file map names only
`.uf/schemas/model-matrix/`; this change follows the established
registry convention instead — canonical root `schemas/
model-matrix/`, deployed `.uf/schemas/model-matrix/`, and
embedded scaffold mirror, all byte-identical (review-matrix,
review-dispatch precedent). The contract semantics (closed
surface, override whitelist) are unchanged; only the copy set
follows repo convention.

### D4: Override whitelist is `profiles.*.model`,
`profiles.*.variant`, `classifier.llm.*` — fail-closed

Mirroring `validateOverride()`/`mergeOverride()` from
review-dispatch: the override is parsed and walked for
whitelisted nested paths only; each matching leaf replaces its
base counterpart; non-whitelisted keys (including structural
keys like `packs`, `defaults`, `agents`, `classifier.enabled`)
fail validation with an error naming the whitelist. `null`
model values propagate as host fallback. The merged result is
re-validated against the identical schema; empty and
comments-only overrides are no-ops. Distinct failure policy,
stated crisply: STRUCTURAL config failures (invalid matrix,
invalid override) fail closed like `loadPolicies` — the
delegation call errors with a clear message; CLASSIFIER runtime
failures never block (D5). This split is the issue's contract
(AC5 fail-closed override vs. AC7 classifier-never-blocks).

`classifier.llm.*` admits any nested key under `classifier.llm`
(endpoint, model, apiKey, `request_timeout_ms`) because that
subtree is exclusively per-developer and carries no governance
surface; everything else under `classifier` is structural and
tracked. The endpoint MUST use the `https` scheme or `http`
with a loopback host (`localhost`, `127.0.0.1`, `[::1]` —
local endpoints such as Ollama); any other `http` endpoint
fails validation before any request is issued (MM-FR-006).

### D5: Classifier cascade — agent engine, LLM engine, static fallback

Modeled on the auto-question plugin pattern
(investigate-then-decide subagent with a JSON-only response
contract, plus an LLM-engine fallback):

- **Agent engine**: hidden subagent
  `.opencode/agents/model-matrix-classifier.md` (mode
  `subagent`, hidden, description "Internal to the model-matrix
  selection layer"). It investigates the feature dir's
  spec/plan/tasks plus existing code/diff with tools, then
  responds with ONLY `{"pack": "<name>", "confidence":
  <0.0-1.0>, "reasoning": "<optional>"}` — no markdown fences,
  no prose, never questions. Bounded by the configured
  `agent_timeout_ms` (default 45_000). The child session
  inherits the session model hierarchy — no endpoint config.
- **LLM engine**: chat-completions POST, `temperature: 0`,
  `response_format: {type: "json_object"}` with a single
  400-retry without the field, request timeout via AbortSignal
  (the configured `classifier.llm.request_timeout_ms`, default
  10_000), apiKey interpolated from `{env:VAR}` (literal keys
  rejected; a reference naming an unset variable skips the
  engine for that invocation — never an empty credential — with
  the skip reason recorded). The endpoint MUST be `https` or an
  `http` loopback host. Endpoint/model/apiKey config lives only
  in the gitignored override under `classifier.llm.*`;
  unconfigured means skipped.
- **Gating**: first valid JSON naming a matrix-valid pack (a
  `packs` key whose pack file exists) with confidence ≥
  `confidence_threshold` (default 0.6) wins. Below threshold,
  malformed, matrix-invalid, timed out, or errored → next engine
  or static fallback (the `defaults` entry for the task kind,
  else host). Never blocks; the audit trail records the
  attempted engine and fallback reason.
- **Caching**: results cached keyed by SHA-256 of (feature_dir,
  prompt, agent), capped at a bounded entry count (named
  constant, 128 — deliberately not matrix-configurable in v1,
  mirroring the derive-per-run-timeout D1 stance on keeping the
  new config surface minimal). Cache hits skip both engines.
  In-process cache; no cross-session persistence in v1. The
  cache is instantiated per cascade instance — the cascade is
  constructed per delegation-tool plugin instance, so the cache
  is never module-level state shared across instances; tests
  inject their own cache to assert hit/eviction behavior in
  isolation (Constitution IV; TC-002/TC-004 isolation).

A classifier-selected pack performs full pack double duty: its
profile model is used AND the pack directive renders in the
child prompt.

Testability (Constitution IV): the LLM engine's `fetch` and the
agent engine's session spawn are injectable seams; every
classifier behavior — thresholds, timeouts, malformed JSON,
fences-wrapped JSON, matrix-invalid packs, cache hits — is
unit-testable without network or a live subagent.

### D6: Pack directive is tool-rendered, fixed template, golden-tested

`renderPackDirective(pack)` produces a short instruction to load
`.opencode/uf/packs/<pack>.md` before writing — never the pack
content inlined, never agent-reconstructed (mirroring
`renderSkillDirective`). Fixed position after the skill
directive, before the extra prompt; golden-shape test asserts
the exact rendering and position. Rationale: the directive is a
deterministic tool artifact (Constitution VI) — an
LLM-reconstructed directive would be untestable prose.

### D7: Audit trail extends the existing delegation artifacts

`DelegateCommandData` and the checkpoint data gain additive
fields: `selection_source` (the eight-value enum:
`explicit-model`, `explicit-profile`, `explicit-pack`,
`agent-pin`, `default-kind`, `classifier-agent`,
`classifier-llm`, `host`), the selected pack, the resolved
capability, the resolved model identity, and classifier
confidence when applicable. Calls with no matrix record
`selection_source: "host"` — the pre-change envelope shape
extended, never removed (the delegation-skill-enforcement
backward-compatibility precedent: calls omitting new fields keep
the exact pre-change checkpoint shape). The executor's existing
`provenance` block continues to carry the resolved model for
child-session provenance; `requested_model` is populated when an
explicit `model` param was supplied.

### D8: Zero-config short-circuit

Before any selection work, the tool probes for
`.uf/model-matrix.yaml` via the injected filesystem dependency.
Absent → the selection layer is skipped entirely: no override
read, no classifier, host model, `selection_source: "host"`,
identical error semantics (including `HOST_MODEL_UNAVAILABLE`
for structural introspection failures — preserved because the
host path is the only path in that state). Present-but-invalid
matrix or override fails closed with a clear error (D4). A
present file that parses to no keys — empty or comments-only —
is the single documented exception to MM-FR-001's closed v1
surface: it is a valid no-op matrix (every chain step falls
through to host), whereas a present file carrying at least one
key MUST satisfy the closed surface in full — a partial surface
is not a no-op. This
guarantees AC4 (missing matrix = behavior identical to today)
by construction: the new code paths are unreachable without the
file.

### D9: Tracked v1 matrix shape and the starter capability set

The tracked matrix follows the issue's recommended v1 shape:
`version: 1`; `profiles` with the starter capabilities
`spec-authoring`, `code-writing`, `content-writing`,
`mechanical` (empty in tracked config — identities arrive via
override); `packs` mapping `meta-prompt` → spec-authoring, `go`/
`typescript`/`python` → code-writing, `content` →
content-writing; `defaults` mapping `plan`/`tasks` →
spec-authoring and `implement` → code-writing; empty `agents`;
`classifier` structure only. Cross-cutting packs (severity,
determinism, review-voice, default) get no matrix entries —
they are never the primary capability of a task; the schema does
not forbid entries for them, the tracked starter config simply
omits them (a team may add one).

Two starter-matrix facts are deliberate and documented rather
than accidental. First, `mechanical` is explicit-profile-only:
no starter `packs` entry or `defaults` kind reaches it and the
classifier resolves through `packs`, so only an explicit
`profile: mechanical` selects it (MM-FR-001). Second, the
`plan`/`tasks` → spec-authoring mapping is a tracked-file
conformance rule — verified by the task 4.1 tracked-file test,
not by schema validation, since arbitrary conforming matrices
MAY map those kinds to any present profile (MM-FR-001).

The `meta-prompt` pack (#888) is authored inside this change as
a distinct task because the starter `packs:` mapping requires
the pack file to exist for MM-FR-001 validation. Pack scope per
#888: MP-NNN rule ids, `[MUST]/[SHOULD]/[MAY]` tags,
consolidation of existing authoring conventions (RFC 2119,
FR-NNN numbering, Given/When/Then, artifact shapes, `[P]`
markers, checkbox discipline, determinism-first command prose),
no literal clarification-marker syntax in the prose (a
well-formed marker extracts as a question even inside
documentation). It matches the format and voice of `go.md`,
`content.md`, `determinism.md`. Two test-registry ripples are
exemption decisions recorded here: `internal/schemas/
packvalidator_test.go` skips meta-prompt.md in
TestValidateConventionPack_AllPacksValid (mirroring the
review-voice.md precedent for packs the validator's expectations
do not cover), and `internal/scaffold/scaffold_test.go` gains a
knownNonEmbeddedFiles entry for the pack (mirroring the
governance precedent of determinism.md) — the pack is
convention content, not a scaffold-deployed asset, in slice 1.

### D10: Scaffold registration surface

`internal/scaffold/scaffold.go`: the lib asset enumeration (the
two `opencode/lib/*.ts` lists) gains `model-matrix.ts`;
`gitignoreBlock` gains `.uf/model-matrix.override.yaml`; the
embedded `uf/model-matrix.yaml` template and
`uf/schemas/model-matrix/` registry mirror land under
`internal/scaffold/assets/uf/`. The uf-workflow plugin mirror
and the new lib mirror must be byte-identical to canonical (the
existing `TestEmbeddedAssets_MatchSource` drift test is extended
to the new files). Deployment behavior of the new assets to
hero repos is slice 2 (#889); this change only registers the
surface so the mirrored plugin loads and drift stays enforced.

Ownership classification (drift obligations follow from it):
the embedded `uf/model-matrix.yaml` template and the
`.uf/schemas/model-matrix/` registry mirror are tool-owned —
byte-identical to their canonical sources, drift-enforced by
the extended `TestEmbeddedAssets_MatchSource`; the lib and
plugin mirrors are tool-owned the same way; the
`.uf/model-matrix.override.yaml` file is user-owned —
gitignored, never scaffolded beyond the gitignore entry, exempt
from drift checks exactly like `*-custom.md` packs (the
custom-pack ownership precedent); a deployed repo's tracked
`.uf/model-matrix.yaml` is team-owned config once `uf init`
writes it (slice 2), at which point it diverges from the
template by design.

### D11: Command guidance is documentation-only

`.opencode/commands/uf.unleash.md` (and its scaffold mirror)
gains the UF-FR-001 guidance: pass `pack` on implement
delegation when the task clearly names one; the classifier
covers the rest; document the precedence order and the
parameter contract. The tool enforces the contract; the prose
never instructs reconstructing pack content. No behavior change
from the prose itself.

### Constitution alignment (referenced from proposal)

- **Autonomous Collaboration (I)**: matrix + override + audit
  fields are self-describing artifacts; classifier cache is an
  internal artifact keyed by input hash, not live coordination.
- **Composability First (II)**: standalone lib module with
  injected dependencies; matrix, override, and classifier
  config all optional; both classifier engines independently
  skippable.
- **Observable Quality (III)**: `selection_source` + resolved
  model + confidence recorded machine-parseably on every
  delegation; fallbacks auditable, never silent.
- **Testability (IV)**: pure/injectable selection functions;
  fetch and session-spawn seams; mocked-child-session
  integration test proves the non-host-model run (AC1).
- **Security by Default (V)**: identities and endpoints only in
  the gitignored override; `{env:VAR}` apiKey interpolation with
  literal-key rejection; fail-closed whitelist.
- **Deterministic Execution (VI)**: the precedence chain is
  tool-layer code with no LLM judgment on deterministic paths;
  the classifier is bounded, cached, threshold-gated, and
  overridable; the pack directive is tool-rendered.

## Risks / Trade-offs

| Risk | Mitigation |
|-----|-----------|
| Classifier adds latency to implement delegations | Bounded engine timeouts (agent 45 s default; LLM configurable), input-hash cache, and the never-blocks static fallback; explicit `pack`/`profile`/`model` bypasses the classifier entirely |
| Wrong model selected for a task | Explicit params override everything; per-agent pins; `selection_source` audit makes every selection reviewable; zero-config keeps today's behavior |
| Structural config failure blocks delegation | Deliberate: invalid matrix/override fails closed with a clear error naming the path/whitelist (the `loadPolicies` precedent) — silent misconfiguration would be worse; the file can be removed to restore zero-config |
| Prompt-shape change breaks existing golden tests | The pack directive renders only when a pack is selected; zero-config calls compose the identical prompt; golden tests extended, not rewritten |
| Classifier selects a pack the child model handles poorly | The mapping is per-developer config (override); confidence threshold gates weak picks; audit trail exposes the choice for adjustment |
| Override drift/typos in model strings | Merged result re-validated against the schema; invalid model strings fail closed at load |
| New module drags the 94% plugin coverage floor | The module is designed injectable-first; test tasks enumerate loader, merge, precedence, classifier, cache, and integration coverage before verification |
| Scaffold mirror desync breaks hero-repo plugin load | The lib mirror is REQUIRED (SC-FR-007) and drift-enforced in the same change; `make check` runs the drift tests |
| Cache serves stale pack choices after config edits | Cache is in-process and keyed by input hash, not by matrix content; a session's selections re-resolve per matrix load — worst case one delegation per session uses a superseded mapping; acceptable for v1, recorded for revisit |

## Open questions recorded as decisions

Issue #887 leaves two details unspecified; recorded here rather
than blocking:

1. **Classifier cache cap and scope** — bounded in-process cache
   of 128 entries, keyed by SHA-256(feature_dir + prompt +
   agent). Not matrix-configurable in v1 (minimal config
   surface); revisit with evidence.
2. **Static fallback for kinds absent from `defaults`** — the
   host model (the chain's terminal step). `defaults.implement`
   maps to `code-writing` in the tracked starter config, so the
   fallback is reached only for unknown kinds or
   unconfigured matrices. Clarification recorded during
   iteration-2 review: a configured `defaults.implement` never
   short-circuits the cascade — step 5 always proceeds to the
   classifier (step 6), a successful classification wins, and
   `defaults.implement` serves only as the cascade's static
   fallback when both engines fail (MM-FR-003 steps 5–6).
