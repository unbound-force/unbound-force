<!-- originating_issue: 887 is recorded in .openspec.yaml; /uf.finale
     emits `Closes #887` in the PR body. The meta-prompt pack
     (What Changes 6) closes #888 in the same PR, and the
     documentation issue filed per the AGENTS.md documentation
     gate (task 7.3) is #893: the PR body carries `Closes #887`,
     `Closes #888`, and `Closes #893` (task 7.3 bookkeeping). -->

## Why

`uf_delegate_command` — the tool `/uf.unleash` and
`/uf.run-pipeline-plan` use to delegate work to content agents —
always replays the parent session's model. Every delegated task,
from spec/plan/tasks authoring to Go implementation to docs edits,
runs on whichever model the orchestrating session happens to use.

There is a large, stakeholder-observed capability gap between models
good at meta-prompt engineering (spec/plan/tasks/command authoring)
and models good at writing code; other capability categories likely
exist. The review council already solved this shape of problem for
reviewer agents: `.uf/review-matrix.yaml` + a gitignored
per-developer override + JSON Schemas route each Divisor run to a
model suited to its tier. Content delegation has no equivalent —
`executeAgentSession` already accepts any `ModelIdentity` plus
`variant` (review agents run on non-host models today via
`invoke_agent`), so only the selection layer is missing.

Done means (issue #887): spec/plan/tasks authoring work
("meta-prompting") can run on a different model than code
implementation, per configuration; selection is deterministic,
auditable, and overridable per call; a gitignored override carries
model identities while tracked config carries structure only; and
zero-config behavior is unchanged — with no matrix present,
delegation behaves exactly as today (host model).

## What Changes

1. **Model matrix configuration pair** — a new tracked
   `.uf/model-matrix.yaml` declares capability-named profiles
   (`spec-authoring`, `code-writing`, `content-writing`,
   `mechanical`), a `packs:` mapping (UF convention pack →
   capability), `defaults:` by task kind, `agents:` pins, and
   `classifier:` structure — model strings optional everywhere
   (host fallback). A gitignored `.uf/model-matrix.override.yaml`
   carries per-developer model identities via a fail-closed
   whitelist: `profiles.*.model`, `profiles.*.variant`,
   `classifier.llm.*` only. This deliberately mirrors the
   review-matrix trio (tracked config + override + JSON Schema
   registry).

2. **Deterministic selection precedence** — `uf_delegate_command`
   gains `model` (+`variant`), `profile`, and `pack` parameters and
   resolves the child model through a fixed chain: explicit
   `model` → explicit `profile` → explicit `pack` (validated:
   pack file exists AND has a matrix entry) → `agents:` pin →
   static `defaults:` by task kind (`plan`/`tasks` →
   spec-authoring) → classifier cascade (`implement` and unknown
   kinds) → host model. When any matrix profile model resolves,
   host introspection is skipped entirely — delegation no longer
   fails on `HOST_MODEL_UNAVAILABLE`.

3. **Classifier cascade (auto-derivation)** — for `implement` and
   unknown task kinds: a hidden subagent agent engine
   (`.opencode/agents/model-matrix-classifier.md`,
   investigate-then-decide, JSON-only response) then an LLM engine
   (chat-completions POST, `temperature: 0`, JSON response format,
   `{env:VAR}` apiKey interpolation, endpoint config only in the
   gitignored override), gated by a confidence threshold, cached by
   input hash, with static fallback. Classifier failure, timeout,
   or low confidence never blocks delegation.

4. **Pack param double duty** — an explicit or classifier-selected
   pack both selects the model AND renders a tool-generated pack
   directive in the child prompt (fixed position after the skill
   directive, golden-shape tested, never agent-reconstructed).

5. **Audit trail** — the delegation success envelope and checkpoint
   data record `selection_source` (`explicit-model`,
   `explicit-profile`, `explicit-pack`, `agent-pin`,
   `default-kind`, `classifier-agent`, `classifier-llm`, `host`),
   the selected pack/capability/model, and classifier confidence
   when applicable.

6. **Meta-prompt convention pack (#888)** — a new
   `.opencode/uf/packs/meta-prompt.md` (MP-NNN rule ids)
   codifies existing spec/plan/tasks/command authoring conventions
   so the starter `packs:` mapping validates (pack existence
   check). Authored inside this change as a distinct task.

7. **Scaffold mirrors** — the new `.opencode/lib/model-matrix.ts`
   module and the modified uf-workflow plugin are mirrored
   byte-identical under `internal/scaffold/assets/opencode/` (the
   mirrored plugin imports the mirrored lib — the lib mirror is
   required for plugin load in scaffolded hero repos); the tracked
   matrix and its JSON Schema registry gain scaffold mirrors;
   `internal/scaffold/scaffold.go` registers the new lib asset and
   appends `.uf/model-matrix.override.yaml` to the managed
   gitignore block; root `.gitignore` gains the override entry.

8. **Docs and guidance** — `docs/configuration.md` documents the
   config pair; `CHANGELOG.md` gains an entry carrying the
   plugin-suite coverage result and constitution-alignment
   outcome in its footer; `AGENTS.md` gains structural updates
   (project tree, agent count, convention-pack list — task
   7.4); `.opencode/commands/uf.unleash.md` recommends passing
   `pack` on implement delegation when the task clearly names
   one (the classifier covers the rest); a documentation issue
   is filed against this repo (task 7.3) and a cross-repo
   documentation issue against `unbound-force/website` per the
   org constitution's Cross-Repo Documentation rule (task 7.5).

Hero-repo deployment via `uf init` (#889) is slice 2 and out of
scope; this change is the repo-first slice.

## Capabilities

### New Capabilities

- `model-matrix`: capability-driven model selection for delegated
  content-agent work — tracked matrix + gitignored override +
  fail-closed whitelist, deterministic precedence chain, classifier
  cascade with static fallback, selection audit trail.

### Modified Capabilities

- `unleash`: `uf_delegate_command` gains `model`/`variant`,
  `profile`, and `pack` parameters; prompt composition gains the
  tool-rendered pack directive; the success envelope and checkpoint
  data gain selection provenance; the implement-phase delegation
  guidance recommends the `pack` parameter.
- `scaffold`: the tool-owned asset surface gains the
  `opencode/lib/model-matrix.ts` lib mirror, the uf-workflow plugin
  mirror stays drift-enforced, the tracked matrix + schema registry
  gain embedded copies, and the managed gitignore block gains the
  override file.

### Removed Capabilities

- None. Zero-config delegation (no matrix file) is preserved
  verbatim; every existing `uf_delegate_command` call shape
  remains valid.

## Impact

- **New files**: `.uf/model-matrix.yaml`;
  `.uf/schemas/model-matrix/` (v1.schema.json, README.md,
  samples/); `.opencode/agents/model-matrix-classifier.md`;
  `.opencode/lib/model-matrix.ts` (loader, schema validation,
  override merge, precedence resolution, classifier engines;
  injectable `dependencies.readText`); `.opencode/uf/packs/
  meta-prompt.md`; `.opencode/test/model-matrix*.test.ts`.
- **Modified**: `.opencode/plugins/uf-workflow/index.ts` (params,
  selection integration, pack directive, audit fields) and its
  byte-identical scaffold mirror; `internal/scaffold/scaffold.go`
  (lib asset enumeration, gitignore block);
  `internal/scaffold/assets/opencode/lib/` (new mirror);
  `.gitignore`; `docs/configuration.md`; `CHANGELOG.md`;
  `AGENTS.md` (structural updates — project tree, agent count,
  convention-pack list); `internal/schemas/packvalidator_test.go`
  (meta-prompt.md validation exemption); `internal/scaffold/
  scaffold_test.go` (knownNonEmbeddedFiles entry + UF-FR-001
  prose pin); `.opencode/commands/uf.unleash.md` (documentation-only
  guidance); `.opencode/test/unleash-delegate.test.ts` (extended).
- **Unchanged**: `resolveHostModel` and `executeAgentSession` in
  `.opencode/lib/agent-executor.ts` (the executor already accepts
  any model identity; host introspection remains the fallback
  path). The review-matrix trio is untouched.
- **Call sites**: `.opencode/commands/uf.unleash.md` and
  `.opencode/commands/uf.run-pipeline-plan.md` are the only
  `uf_delegate_command` callers; behavior with no matrix present
  is identical to today.
- **Security**: model identities live only in the gitignored
  override; the LLM engine interpolates the apiKey from
  `{env:VAR}` (established secrets pattern — never a literal key);
  tracked config carries structure only.
- **Gates**: no coverage floor, review limit, CI flag, or pinned
  dependency changes. The `.opencode/` suite must hold its
  governance-controlled 94% statement coverage floor.
- **Documentation gate**: user-facing change — two issues must
  be filed before PR merge: the repo documentation issue
  (AGENTS.md rule; task 7.3, closed by the PR alongside #887
  and #888) and the cross-repo `unbound-force/website`
  documentation issue (org constitution v1.4.0 Cross-Repo
  Documentation rule; task 7.5).
- **Content coverage**: the model matrix is a configuration
  reference for an existing audience (agent-workflow operators
  already reading `docs/configuration.md`) — below the
  multi-step-feature threshold that warrants a blog post; the
  tutorial bar is met by documenting it in
  `docs/configuration.md` (task 7.1), with the website issue
  (task 7.5) recording the tutorial disposition (configuration
  reference, citable from existing delegation docs, no new
  tutorial).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0,
amended 2026-10-14; six active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The matrix is a self-describing configuration artifact consumed by
the delegation tool; no runtime coupling between heroes is
introduced. The selection audit trail (`selection_source`, resolved
pack/capability/model, classifier confidence) extends the existing
delegation envelope and checkpoint artifacts, so post-hoc auditing
requires no session reconstruction. Classifier results are cached
artifacts keyed by input hash, not live coordination.

### II. Composability First

**Assessment**: PASS

The matrix is entirely optional — absent file, absent override,
and absent classifier config are all valid no-ops (host model,
exactly today's behavior). The selection logic lives in a
standalone `.opencode/lib/model-matrix.ts` module with injected
filesystem dependencies, usable and testable without the plugin.
The classifier's two engines are independently skippable. No new
mandatory dependencies; the LLM engine's endpoint config is
optional and lives only in the gitignored override.

### III. Observable Quality

**Assessment**: PASS

Every delegation records `selection_source`, the resolved
pack/capability/model, and classifier confidence in
machine-parseable form (success envelope + checkpoint data) — the
selection equivalent of the review council's
`resolved_parent_model` provenance. Classifier failures and
fallbacks are recorded, never silent. The tracked/override split
makes team structure vs. per-developer identity inspectable in
version control.

### IV. Testability

**Assessment**: PASS

The loader, override merge, precedence resolution, and pack
validation are pure or injectable (`dependencies.readText`)
functions testable in isolation. The classifier engines are
testable with injected fetch/agent-executor seams; the acceptance
criterion of a non-host-model delegation run is provable with a
mocked child session. No external services are required to verify
any requirement (the LLM engine's network call is behind an
injectable seam; the agent engine reuses the tested
executor/subagent path).

### V. Security by Default

**Assessment**: PASS

Model identities and the LLM endpoint/apiKey config live only in
the gitignored override — nothing credential-shaped is tracked. The
LLM engine interpolates the apiKey from `{env:VAR}` (never a
literal key) and enforces a request timeout via AbortSignal. The
override whitelist fails closed: any key outside
`profiles.*.model`, `profiles.*.variant`, `classifier.llm.*`
fails validation with an error naming the whitelist, preventing
silent governance changes. The classifier child session inherits
the session model hierarchy (no endpoint config needed).

### VI. Deterministic Execution

**Assessment**: PASS

Selection is a documented precedence chain evaluated by tool-layer
code — identical matrix + call inputs produce identical model
choices, with no LLM judgment on the deterministic paths (explicit
params, agent pins, static defaults). The classifier — where
judgment genuinely participates — is a bounded, JSON-contract
subagent/LLM call whose result is cached by input hash, gated by a
confidence threshold, and overridable by any explicit parameter;
its failure modes all fall back to a static default rather than
blocking. The pack directive is tool-rendered (golden-shape
tested), never agent-reconstructed.
