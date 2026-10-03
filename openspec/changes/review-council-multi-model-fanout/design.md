## Context

Divisor commands currently dispatch each selected persona once through
the host-model Task tool. This hides provenance and cannot express
ordered, independent model assessments.

The fullsend reference proves the Option B split: a shared advisor owns
policy and a thin plugin owns explicit-model execution. This design
retains that split, but corrects target-review findings before porting.
See `proposal.md` for motivation and constitution alignment.

The reference baseline is fullsend commit
`b37dbc597baf0180008d3e4a5a764521450b0996`. Its SHA256 digests are:

- `.uf/review-matrix.yaml`
  `21820c2bdd28f59b0730fcb883684bde4759a0c1b2a08fc614fac3f1e4fbb9b2`
- `.opencode/skills/dispatch-advisor/SKILL.md`
  `1bc5da7756782721ac9b649e27c4716449cc1e7a466bf34211f4f909c90a94cd`
- `.opencode/plugins/review-matrix/index.ts`
  `f68d6f5021ccc190529de49b3cb5f5e29efd05cda90bf5a4efd2997bd830799d`
- `.opencode/commands/uf.review-council.md`
  `e58c6228062cc19c27d64165bec36c2095cbd98709b17db8b4e4fa78964afd1e`
- `.opencode/commands/uf.triage-issue.md`
  `ce147d32de05b4785377084614334c382c0cd0c03714f3767716376a5dd64662`
- `.opencode/commands/uf.address-feedback.md`
  `93347aa8dd4f767e9ebd747f892fe565de6bea0a7b6c301d1d20bf27444d06f6`
- `.uf/sibling-repos.yaml`
  `9c015b7fa2d517972cb2106c7c53a03632a8747954cc5c3d67af1b48f8b63f64`

Task 1.1 re-verification on 2026-10-01 used a clean working
checkout of `/Users/jflowers/Projects/github/fullsend-ai/fullsend`.
`git rev-parse HEAD` returned the commit above. `shasum -a 256`
returned the seven digests above without a mismatch.

The source-to-target adaptation ledger is:

1. `.uf/review-matrix.yaml` becomes the canonical target matrix and
   the byte-identical
   `internal/scaffold/assets/uf/review-matrix.yaml` mirror. Preserve
   version 2 and the tier names, but replace scalar profiles with
   closed model and optional-variant objects. Add explicit ordered
   runs, the advisor allowlist, opt-in augmentation, bounded limits,
   strict validation, `specs`-to-`spec` normalization, and host-model
   fallback.
2. `.opencode/skills/dispatch-advisor/SKILL.md` becomes the canonical
   target skill and its byte-identical scaffold mirror. Remove
   `divisor-fullsend` and separate review-capable personas from content
   personas through the closed reviewer manifest. Delegate
   deterministic parsing, selection, tiering, limits, and plan
   serialization to `plan_review_dispatch`; the skill does not
   independently recreate policy or persist the legacy cost log.
3. `.opencode/plugins/review-matrix/index.ts` becomes
   `.opencode/plugins/invoke-agent/index.ts` and its scaffold mirror.
   Keep the plugin policy-free, but add manifest-backed agent
   validation, bounded input, parent-model resolution, prompt-level
   variant compatibility, timeout and cancellation propagation, usage
   provenance, and error redaction. Move plan and artifact policy into
   the separate `review-dispatch` plugin.
4. `.opencode/commands/uf.review-council.md` updates the canonical
   target command and its scaffold mirror. Preserve target review
   gates,
   severity, fix-loop, and posting semantics while adding the closed
   argument grammar, immutable PR or local input context, validated
   plans, planned invocations, deterministic consolidation, final
   artifacts, model self-reporting, and fail-closed no-success
   outcomes.
5. `.opencode/commands/uf.triage-issue.md` updates the canonical target
   command and its scaffold mirror. Replace the reference's generic
   change signal with deterministic normalized issue content and its
   versioned hash frame. Apply majority first within each persona and
   then across personas, preserve all mutation gates, and add planned
   invocations, provenance, artifacts, model self-reporting, and
   fail-closed no-success handling without inventing a code diff.
6. `.opencode/commands/uf.address-feedback.md` updates the canonical
   target command and its scaffold mirror. Preserve Tier 1 and all
   human, review, and push gates. Route only Tier 2 through validated
   plans, use
   the strictest successful recommendation, and add provenance,
   artifacts, model self-reporting, and fail-closed no-success
   handling.
7. `.uf/sibling-repos.yaml` becomes the live target hero declaration,
   while `internal/scaffold/assets/uf/sibling-repos.yaml` is an empty
   commented template. Replace fullsend-specific entries and unsafe
   absolute fallbacks with closed-schema declarations for relevant
   `unbound-force/*` heroes. Acquisition verifies origin, clean state,
   and commit; confines and hashes bounded evidence; treats content as
   untrusted; and replaces raw lesson directives with structured,
   source-grounded proposals.

The installed `@opencode-ai/plugin` client is v1 typed. Its prompt and
message types omit runtime `variant` fields. Two narrow local
compatibility types are therefore required at those exact boundaries;
changing the pinned SDK is not.

### Task 1.2: Verified OpenCode invocation ABI

Verification on 2026-10-01 used only installed declarations, generated
client source, and the local runtime. The exact evidence is:

- `.opencode/node_modules/@opencode-ai/plugin/package.json` is version
  `1.4.10` and depends on `@opencode-ai/sdk` version `1.4.10`.
- The installed SDK package is version `1.4.10`. No pin was changed.
- `PluginInput.client` is the return type of the root SDK
  `createOpencodeClient`. It is the v1 client, not the v2 client.
- `/opt/homebrew/bin/opencode --version` returned `1.18.32`.
- The running local service returned version `1.18.32` from
  `GET /global/health`. Its generated `GET /doc` contract was used to
  check the server request and response shapes below.

The installed v1 `SessionPromptData["body"]` supports `messageID`,
`model`, `agent`, `noReply`, `system`, `tools`, and `parts`. It omits
`variant`. The installed v2 declaration and the runtime-generated
OpenAPI contract both place `variant` beside `model` in the prompt
body. It is not nested inside `model`.

The implementation MUST keep the prompt compatibility type at that
boundary. It MUST derive the body from the plugin client's own prompt
parameter rather than import or instantiate a second client:

```ts
type PromptOptions = Parameters<
  PluginInput["client"]["session"]["prompt"]
>[0]

type PromptBodyWithVariant =
  NonNullable<PromptOptions["body"]> & {
    variant?: string
  }
```

The installed current-message response types expose the actual
`providerID` and `modelID` but omit the runtime `variant`. After
narrowing the response role to `assistant`, the implementation MAY add
only that missing field at the response boundary:

```ts
type CurrentAssistantWithVariant<T> = T & {
  variant?: string
}
```

The value MUST still be checked as a non-empty bounded variant before
use. No compatibility type may widen the session, client, or complete
message response.

The common v1 client and local-runtime child contract is:

1. For a host run, call `client.session.message` with
   `context.sessionID`, `context.messageID`, and the current directory.
2. Require current-message data with assistant role, non-empty
   `providerID`, non-empty `modelID`, and a valid active `variant`.
   Build the resolved parent pair from those response values.
3. Call `client.session.create` with `body.parentID` equal to
   `context.sessionID`, a diagnostic title, and the current directory
   in `query.directory`.
4. Read the child id from the fields-style response at `data.id`.
   Handle `error` or absent `data` before prompting.
5. Call `client.session.prompt` with `path.id` equal to that child id,
   the same `query.directory`, and one text part.
6. Put `agent`, the requested or resolved model, top-level `variant`,
   and `parts` in the body. Do not send `tools` or a permission
   override. A host run MUST replay the resolved parent pair and active
   variant explicitly.
7. Treat `data.info.error` as a failed run even when response data is
   present. The union includes `MessageAbortedError` and provider or
   API errors.

The prompt response is `{ info, parts }`. The installed and runtime
contracts define ordered `TextPart` entries and `StepFinishPart`
entries. Text extraction MUST retain ordered text parts whose
`ignored` flag is not true, join them with two line feeds, and trim
only the combined edges.

A response can contain more than one `step-finish` part because each
provider step emits one. Each part contains `cost` and token counts for
`input`, `output`, `reasoning`, `cache.read`, and `cache.write`.
The local runtime may also emit `tokens.total`, but the installed v1
type does not declare it. The implementation MUST NOT depend on that
extra field.

Usage MUST sum every `step-finish` part, not select the first one.
Zero is a valid reported value. Usage is null only when no finish part
is available. The authoritative actual model pair comes from
`data.info.providerID` and `data.info.modelID`. Requested model and
variant remain separate request provenance.

`ToolContext.abort` is the parent cancellation signal. SDK request
options extend `RequestInit`, so `signal` is valid for create, prompt,
and abort requests. Generated client source copies that signal into
the Fetch `Request`. The root SDK client also disables its
non-standard request timeout, so it supplies no per-run timeout.

The plugin MUST create its own timer and linked `AbortController`.
It MUST pass the linked signal to child creation and prompting. Once a
child id exists, parent cancellation or timeout MUST also call
`client.session.abort` for that child. The abort call MUST use an
independent cleanup signal because the linked signal is already
aborted. The timer and parent listener MUST be removed in `finally`.
An abort that races with child creation MUST abort a returned child
before any prompt starts. Runtime `session.abort` returns a boolean and
stops active AI processing or command execution.

#### Selected host-model resolution

The approved design assumed that `parentID` plus an omitted prompt
model made the child inherit the caller model. A provider-free runtime
probe disproved that assumption:

- A parent was created with probe provider, model, and variant values.
- A child created with only that `parentID` reported null agent and
  null model values.
- A `noReply` child prompt with no model resolved the configured
  default `opencode-go/deepseek-v4-pro` and variant `high`, not the
  parent probe values.
- Both probe sessions were deleted after verification.

The selected correction keeps one `invoke_agent` path. For `source`
`host`, the plugin resolves the current assistant through
`context.sessionID` and `context.messageID`. It reads the actual model
pair and active variant, then replays all three values explicitly to
the child prompt. It does not read project or agent defaults.

Plan and request provenance remain `source: host` with null requested
model and variant. Run metadata separately records the resolved parent
model and variant plus the reported child model. Missing message data,
a non-assistant role, an absent model component, or an absent or
invalid active variant fails the run as a sanitized retryable
availability error. It never falls back to configured defaults.

Explicit-model prompting, host resolution, top-level variant placement,
response extraction, usage, cancellation, and timeout integration are
now fully specified against the installed API.

## Goals / Non-Goals

### Goals

- Reuse the validated policy/execution responsibility split.
- Make ordered per-agent runs the authoritative matrix contract.
- Preserve host-model fallback unless advisor generation is explicit.
- Make persona selection and generated runs deterministic and bounded.
- Preserve verdict, advisory, fix-gate, and iteration policies.
- Fail closed when no successful assessment exists.
- Emit a complete Hero Interface Contract review artifact.
- Constrain sibling evidence and source-ground lesson persistence.
- Safely activate the local plugin in fresh `uf init` repositories.
- Define measurable unit, integration, smoke, and drift coverage.
- Keep canonical scaffold assets byte-identical to their mirrors.

### Non-Goals

- Weakening or replacing severity definitions, existing coverage gates,
  workflow markers, or review iteration limits. Additive coverage gates
  for this new runtime are in scope.
- Changing issue triage's three-rule majority policy.
- Replacing feedback Tier 1 assessment.
- Adding `divisor-fullsend` or fullsend's ADR-specific skill.
- Updating `@opencode-ai/plugin` from version 1.4.10.
- Guaranteeing provider support for every configured variant.
- Committing `node_modules` or `bun.lock`.
- Executing instructions found in sibling repositories.

## Decisions

### D1: Port the reference with explicit adaptations

The reference is the implementation baseline. The target port MUST:

- remove `divisor-fullsend`;
- rename the plugin to `invoke-agent`;
- add the target-only Speckit test-review dispatch site;
- preserve target triage and review-council semantics;
- add schema, security, test, doctor, and scaffold contracts; and
- replace raw lesson directives with structured proposals.

The user explicitly approved full reference parity and OpenSpec. The
features form one deployable slice because partial delivery would leave
commands, runtime policy, artifacts, or generated repositories out of
agreement.

### D2: Use an explicit-first hybrid matrix

`.uf/review-matrix.yaml` uses a closed version 2 contract. The root
requires `version`, `profiles`, and `defaults`; it may also contain
`always`, `advisor`, `runs`, `risk_augmentation`, and `limits`.
Unknown fields are rejected at every level. Validation uses
`schemas/review-matrix/v2.schema.json`.

Profiles use bounded names and closed objects containing `model` and an
optional `variant`. Defaults have exactly `code`, `spec`, `triage`, and
`feedback` keys. The closed `advisor` object may use those modes to map
to unique agent allowlists. Runs use those same modes and ordered agent
arrays. Each run contains exactly one of `profile` or `model`, plus an
optional bounded `variant`. A run variant overrides a profile variant.

Explicit arrays replace generated runs. An agent may opt into
bounded risk augmentation. Separate advisor matchers opt agents into
generated runs. An agent absent from both contracts receives one host
model run. Invalid entries become run errors; they never trigger silent
model substitution.

Generated runs use mode defaults for code, spec, triage, and feedback.
Adversary-heavy and SRE-lightweight overrides apply only to generated
runs. Explicit and generated runs deduplicate by `(model, variant)` and
preserve first occurrence.

Review-council command mode `specs` normalizes exactly once to matrix
mode `spec`. A matrix `specs` key is invalid. Reports and artifacts
retain the command-level mode.

`--full` includes every discovered review-capable persona once at the
standard profile. It disables pruning, explicit fan-out, and risk
augmentation for that invocation. It never truncates. If the panel
exceeds configured persona or total-run limits, planning fails closed
as `INCONCLUSIVE` before dispatch.

### D3: Use a closed reviewer manifest and executable planner

The advisor skill owns change profiling and invokes the registered
`review-dispatch` policy plugin. That plugin exposes
`plan_review_dispatch` and `finalize_review_dispatch`. The first tool
parses the matrix and reviewer manifest, applies relevance, tiers,
augmentation, and limits, then emits a validated version 1 plan. The
second performs
semantic result validation and atomic artifact persistence. Markdown
commands MUST NOT reimplement those deterministic operations.

Plan entries are ordered by agent name, then run sequence. Each entry
contains `agent`, `decision`, `reason_code`, `reason`, and `source`. It
also contains nullable `tier`, `model`, and `variant`, plus `sequence`,
`read_only` metadata, and validation errors. `source` is `explicit`,
`advisor`, or `host`.

`.uf/reviewer-capabilities.yaml` is the only reviewer eligibility
source. It validates against a closed version 1 schema and contains one
unique entry for each known or opted-in `divisor-*` agent. Agent
frontmatter is not modified and MUST NOT receive reviewer metadata as
provider options.

The manifest fixes these ordered scopes:

- adversary: review; security, dependencies, standard;
- architect: review; standard, cli-ux, ci-cd, documentation;
- curator: review; documentation;
- guard: review; standard, cli-ux, documentation;
- SRE: review; ci-cd, dependencies, security;
- testing: review; test-quality; and
- envoy, herald, and scribe: content with empty scopes.

Adversary and guard are always included. Other review agents require a
scope intersection. Curator is included only for documentation or
user-facing changes. Unknown agents require an explicit valid manifest
entry. Missing, duplicate, conflicting, or unknown manifest data fails
planning as `INCONCLUSIVE`. An OpenCode-load test proves no reviewer
metadata enters agent provider options.

Speckit test review requests code profiling restricted to
`divisor-testing`; it does not add a fifth mode default.

Issue triage never synthesizes a code diff. It profiles normalized
issue title, body, and comments. Text is UTF-8 NFC. CRLF and bare CR
become LF.
Comments are ordered by creation time, then numeric comment id.

Hash input uses the `uf-issue-content-v1` length-prefixed UTF-8 frame.
Lengths and counts are canonical ASCII decimal without leading zeros.
The frame is:

```text
uf-issue-content-v1\n
title:<bytes>\n<title-bytes>\n
body:null\n
```

For a present body, including an empty body, `body:null\n` is
replaced by `body:<bytes>\n<body-bytes>\n`. The remaining frame is:

```text
comments:<count>\n
comment-id:<bytes>\n<decimal-id>\n
created-at:<bytes>\n<UTC-RFC3339-seconds>\n
comment-body:<bytes>\n<comment-bytes>\n
```

The three comment fields repeat for every ordered comment. SHA256
covers the complete frame, including labels, lengths, and separator
LFs. `text_bytes` counts only normalized title, present body, and
comment body bytes. It excludes framing, ids, timestamps, and
separator bytes. A null body contributes zero text bytes but hashes
differently from an empty body.

ASCII case-folded tokens map through a checked-in closed keyword table
to the Protocol 3 categories. Unmatched content receives `standard`.
Matched rule ids are unique and sorted. Security matches set the
security flag. Issue tiering is heavy for security, more than 32768
text bytes, or more than 20 comments. It is standard for more than 4096
bytes, more than five comments, or more than one category, and
lightweight otherwise. The plan and artifact record byte count, comment
count, content hash, categories, matched rules, and tier.

### D4: Make tiering total and fan-out bounded

Tier precedence is:

1. Heavy when security-sensitive, lines exceed 300, or files exceed 10.
2. Standard when lines exceed 50, files exceed 3, or component span
   exceeds 1.
3. Lightweight otherwise.

Component span counts normalized parent directories. Root files use
the `<root>` bucket. Tests are grouped with their source directory when
relationship is identifiable.

Risk fan-out is opt-in. The adjacent direction is lightweight to
standard, standard to heavy, and heavy to standard. Security-sensitive
adversary review and architect review above 500 lines may add the one
adjacent run. Duplicate model and variant pairs are omitted.

Default limits are 16 personas, 3 runs per persona, 24 total runs,
4 parallel runs, 600 seconds per run, and USD 25 cumulative reported
cost. Cost is checked between batches. Run limits remain enforceable
when cost is unavailable. Parent cancellation cancels children. Skipped
runs are recorded and may cause `INCONCLUSIVE` when no assessment runs.

Configuration is bounded to 1-32 personas, 1-5 runs per persona,
1-64 total runs, 1-8 parallel runs, 30-1800 seconds per run, and a
USD 1-100 reported-cost ceiling. Out-of-range configuration is invalid.

### D5: Keep `invoke_agent` policy-free and defensive

The plugin validates a manifest-declared review-capable `divisor-*`
agent, an optional bounded model, an optional bounded variant, and a
bounded prompt. Direct model validation intentionally matches the
reference when a model is supplied: split at the first slash and
require non-empty provider and model-id remainders. It does not
inherit the matrix grammar. For a host run, the plugin resolves the
current assistant message by the tool context ids and explicitly
replays its actual model pair and active variant. Resolution failure is
a sanitized availability error, never default substitution. The child
uses the selected agent's existing OpenCode
permission contract. This change MUST NOT broaden, narrow, replace, or
reinterpret those permissions. The `read_only` value is retained as
invocation metadata for compatibility with the reference. Its value
has no allow or deny effect. The child receives cancellation and
timeout.

Maximum lengths are 128 characters for agent, 256 for model, 64 for
variant, and 128 KiB UTF-8 for prompt. Error output redacts credential
values, bearer tokens, URL query strings, and absolute home paths.

The plugin returns text and authoritative request provenance. A host
run records null requested model and variant, resolved parent model and
variant, and the reported child model.
Cost and token usage are nullable and appear only when OpenCode
supplies them.
The plugin does not read policy, consolidate findings, or write
artifacts.

Variant is passed at the prompt-body top level through one narrow local
compatibility type. One additional narrow current-message compatibility
type reads the runtime variant omitted by the v1 response type.
Requested variant provenance is retained even when the runtime rejects
it; rejection is a run error.

### D6: Preserve workflow semantics and fail closed

Commands consolidate model runs per persona, then apply workflow
policy.
Fan-out therefore does not create extra votes.

Review council returns `REQUEST CHANGES` when a successful run blocks.
Otherwise it returns `APPROVE WITH ADVISORIES` when a run has
advisories, then `APPROVE`. Existing human confirmation for LOW/MEDIUM
auto-fixes, no-auto-fix behavior for HIGH/CRITICAL, and three-iteration
limits remain unchanged.

Triage applies its existing three-rule majority first within a persona,
then across persona votes. Feedback uses its strictest successful
recommendation. Speckit test review preserves its blocking contract.

Failed runs do not vote or fabricate findings. Every workflow requires
at least one successful assessment. Zero successes yield `INCONCLUSIVE`
or `UNAVAILABLE`, block automated progression, and request retry or
human review. No workflow may vacuously approve, validate, or accept.

Cause precedence is deterministic. A policy, plan, budget, limit,
persistence, or calculation failure yields `INCONCLUSIVE`, even when an
availability failure also occurred. `UNAVAILABLE` applies only when
every blocking cause is provider, model, or runtime availability. Mixed
therefore yield `INCONCLUSIVE`.

The artifact preserves the native workflow result and projects it to a
generic verdict without losing information. Council and test-review
results map identically. Triage maps `VALID` to `APPROVE`, `INVALID` to
`REQUEST CHANGES`, and `NEEDS-CLARIFICATION` to
`APPROVE WITH ADVISORIES`. Feedback maps `ACCEPT` to `APPROVE` and
`AUTHOR-DECIDES` to `APPROVE WITH ADVISORIES`. Native `INCONCLUSIVE`
and `UNAVAILABLE` map identically.

`UNAVAILABLE` means no runnable or successful assessment exists
because of provider, model, or runtime availability. `INCONCLUSIVE`
means policy
or plan validation failed, execution limits prevented assessment,
artifact persistence failed, or another condition prevented a valid
assessment from being calculated.

Findings deduplicate by normalized file and root cause while retaining
all run provenance. Existing compound-severity logic runs afterward.

### D7: Emit a versioned Hero Interface Contract artifact

Add `schemas/review-dispatch/v1.0.0.schema.json`, valid and invalid
fixtures, registry validation, and compatibility tests. Its envelope:

- `hero`: `the-divisor`;
- `version`: producer semver;
- `artifact_type`: `review-dispatch`;
- `schema_version`: `1.0.0`;
- UTC RFC3339 `timestamp`;
- context with branch, commit, correlation id, and workflow id; and
- a typed review-dispatch payload.

`review-dispatch` is an additive observability artifact. Canonical
`review-verdict` moves to schema 2.0.0. The new major version adds
native `INCONCLUSIVE` and `UNAVAILABLE` council decisions. Version
1.0.0 remains readable historical data, but producers emit 2.0.0 and
Mx F,
Cobalt-Crush, and Muti-Mind migrate atomically.

Canonical projection is lossless: `APPROVE` becomes `APPROVED`,
`APPROVE WITH ADVISORIES` becomes `ESCALATED`, and `REQUEST CHANGES`
becomes `CHANGES_REQUESTED`. `INCONCLUSIVE` and `UNAVAILABLE` retain
their names. Consumers MUST block automation for both no-success
decisions.
Registry docs, schemas, samples, producer/consumer tables, and major
version compatibility fixtures MUST describe the migration and dual
emission.

Specs 008 and 009 MUST migrate with the implementation. Spec 008 MUST
block orchestration for `INCONCLUSIVE` and `UNAVAILABLE`. Spec 009 MUST
register the version 2 enum, retain v1 historical reads, and
document the major-version compatibility boundary.

The closed payload contains command, mode, full state, an input
context, change profile, plan version and entries, all run outcomes,
coverage,
consolidated findings, advisories, final verdict, native workflow
result, run counts, and correlation id.

JSON Schema validates structure, types, enums, and nullability. The
`finalize_review_dispatch` semantic validator enforces native-result
mappings, terminal-state invariants, and arithmetic. Run counts have
distinct non-negative fields for `total`, `success`,
`failed`, `skipped`, `budget_skipped`, `limit_skipped`, and
`cancelled`. Their sum equals `total` and each value equals the
corresponding terminal
run state. Final artifacts contain no `pending` or `running` runs.

Input context is discriminated by `kind`. A PR context contains its
number, immutable base and head refs, and their 40-character lowercase
commit SHAs. A local context contains null PR number, the selected base
and head refs, and their resolved SHAs. Refs are printable ASCII and
bounded to 255 characters. The input context therefore reproduces the
exact base-to-head review even after branch refs move.

An issue context contains issue number, canonical GitHub issue URL, and
the lowercase SHA256 of normalized title, body, and ordered comments.
It contains no base or head fields. Diff profiles contain line, file,
and component counts. Issue profiles instead contain text bytes,
comment count, matched rule ids, categories, security state, and tier.

Plan and run objects reject unknown fields. Runs use discriminated
states for pending, running, success, failed, skipped, budget skipped,
limit skipped, and cancelled. Success requires an eligible verdict and
null error. Other terminal states do not vote or emit findings. Failed
runs require sanitized errors. Usage remains nullable and is never
fabricated. Findings require severity, category, description, root
cause, nullable location, and contributing run provenance.
Model identity comes from the requested pair or, for host runs, the
resolved parent pair, plus recorded OpenCode response metadata. Host
requested fields remain null. The resolved parent model and variant are
separate run fields. Disagreement is an error, not a silent rewrite.

Files use a correlation-id name under `.uf/artifacts/dispatch/`. The
writer creates directories with mode 0750 and files with mode 0600. It
writes a same-directory temporary file, flushes, closes, chmods, and
publishes the final name through an exclusive same-filesystem hard
link. Link creation fails with `EEXIST` instead of overwriting an
existing target. The writer then removes the temporary name. A
collision generates a new correlation id and retries. Artifact failure
retains the calculated assessment for
human-only reporting, but changes the operation result to
`INCONCLUSIVE`, returns failing status, and blocks automated
progression. It does not fabricate a review finding.

### D8: Constrain sibling evidence before prompt use

Sibling configuration validates against a checked-in closed version 1
schema. The root requires only integer `version: 1` and `siblings`.
Each sibling requires name, URL, default branch, role, fetch mode, and
contracts. Local paths, clone path, cache path, and notes are optional;
null and empty strings are invalid. Names match
`^[a-z][a-z0-9-]{0,62}$` and are unique. Roles are
`downstream-consumer`, `upstream-source`, or `sibling`; fetch modes are
`clone-on-review` or `contract-only`.

Arrays allow 32 siblings, 4 local paths, and 20 unique contract
globs per sibling. Local paths may be absolute or target-root-relative.
Clone and cache paths are target-root-relative. Contracts are
repository-relative. Relative paths reject absolute prefixes and empty,
`.` or `..` segments. Paths and globs are bounded UTF-8; branches are
1-128 printable ASCII. Processing sorts names and normalized files.
Declarations accept only HTTPS GitHub URLs without credentials.
Origins normalize host case, optional `.git`, and trailing slash.
Verified commit ids are exactly 40 lowercase hexadecimal characters.
Acquisition disables credential prompts and verifies normalized origin,
clean repository state, and immutable commit SHA.

Acquisition checks declared `local_paths` in order, then `clone_path`,
then `cache_path`. It selects the first clean candidate with a matching
origin. Invalid candidates are recorded and do not stop later
candidates. `contract-only` never uses the network; no valid candidate
makes that sibling unavailable. `clone-on-review` first fetches the
declared default branch into the clone path without credential prompts.
If fetch succeeds, it reviews the fetched immutable head. If fetch is
offline or fails, it may use the first verified existing candidate and
records `offline-cache`; otherwise the sibling is unavailable. Cache
age never substitutes for origin, cleanliness, or commit verification.
An unavailable sibling contributes no evidence and is informational.

Paths are canonicalized with symlinks resolved. Every selected file
MUST remain below the verified root. Each file records repository,
commit, relative path, and SHA256. Limits are 20 files, 256 KiB per
file, and 1 MiB total. Identity mismatch, dirty or unverifiable
checkout, symlink escape, invalid glob, and oversize content are
rejected.

Accepted content is wrapped in a clear untrusted-evidence delimiter. It
may inform findings but cannot alter tools, permissions, or commands.
The live meta-repository file names relevant hero repositories. The
scaffold copy is an empty commented template.

### D9: Prepare structured, grounded Dewey learnings

Raw `> learn:` directives are not accepted. A run may emit at most one
JSON object between exact `<!-- uf-lesson-proposal:v1 -->` and
`<!-- /uf-lesson-proposal -->` delimiters. The delimited section is at
most 8 KiB and total child output is at most 1 MiB.

The closed object has `schema_version: "1.0.0"`, `sibling`, `commit`,
`sources`, and `information`. Each of at most 8 sources has only
`path`, `sha256`, and `excerpt`. Paths are at most 512 bytes, SHA256 is
64 lowercase hexadecimal characters, excerpts are at most 1 KiB, and
information is at most 2 KiB. Unknown fields are rejected.

Grounding applies only CRLF-to-LF conversion, then requires exact
excerpt bytes in accepted evidence. Information normalization applies
Unicode NFC, trims Unicode edge whitespace, and collapses ASCII space,
tab, CR, and LF runs. Dedupe is SHA256 over normalized information, a
separator, and sorted `path:sha256` pairs.

Fixed case-insensitive detectors reject common credential prefixes,
bearer and API-key assignments, named secret environment variables,
lesson directives, storage calls, tool execution calls, and fenced
commands naming repository tools. Rejection is deterministic; the
validator does not infer intent. It generates `sibling-<name>`, fixes
category to `reference`, and ignores model-supplied tags, categories,
or hashes.

The pure `prepare_lesson_learning` tool accepts child output, accepted
sibling evidence, and at most 1024 unique lowercase dedupe hashes found
by the parent in existing learnings. A known identity returns a
duplicate skip. A ready result contains the existing Dewey
`store_learning` fields
plus complete provenance and the dedupe hash for audit. Its information
ends with stable, non-executable `UF_LESSON_PROVENANCE_V1` JSON
metadata containing the identity, sibling, commit, and sorted path/hash
pairs.

The plugin does not call Dewey. Parent commands query existing
learnings, call the preparation tool, then call `dewey_store_learning`
exactly once
only for a ready result. This human-approved override uses normal Dewey
learning storage; it adds no trust tier or promotion path. Dewey
unavailability and all skips remain informational and do not change the
review verdict.

### D10: Define review-council argument grammar

Tokens are case-sensitive and may appear in any order. Modes are
exactly `code` or `specs`; `--full` may appear once. `PR_NUMBER` is
ASCII `[0-9]+`, has no sign or decimal point, and represents 1 through
999999. Leading zeros are accepted and normalized to decimal. Unicode
digits, zero, overflow, multiple numbers, repeated flags, conflicting
modes, unknown tokens, and `specs` with a PR are rejected before
discovery. A PR implies code mode; explicit valid mode otherwise wins,
then current auto-detection.

### D11: Wire only direct dispatch sites

Review council, issue triage, feedback escalation, and Speckit test
review use the advisor and plugin. Indirect callers are not rewired.
Every site emits the review-dispatch artifact and handles no-success
outcomes. Sibling evidence and lesson proposals apply only where the
command requests sibling context.

Curator receives at most one assessment run, even when other personas
fan out. Review prompts prohibit issue creation. After consolidation
and the existing human gate, the parent may perform one deduplicated
curation action. Child runs never create issues directly.

### D12: Scaffold through staged, reproducible activation

Canonical and mirrored assets include the advisor, both plugins,
matrix, reviewer manifest, sibling template, package manifest, lock,
and three scaffold-owned commands. Speckit test review remains
live-only. The
scaffold maps `uf/` to `.uf/` and updates exact asset inventories.

Track exact npm dependencies and `package-lock.json`. Direct runtime
dependencies are `@opencode-ai/plugin` and `zod`.
Continue ignoring `node_modules` and `bun.lock`; stop ignoring
`package.json`.
Target installation runs `npm ci --ignore-scripts --omit=dev` in the
target `.opencode/` directory. Repository tests use the full locked
development install.

The supported runtime is Node 20 through 24 and npm 10 through 11.
`uf init` resolves both executables and parses semantic versions before
installation. Missing, malformed, or unsupported versions prevent
activation and report the detected value plus a Node 22/npm 10
remediation. `uf doctor` performs the same checks without network
access.

To avoid activation of an unloadable plugin, `uf init` MUST:

1. deploy manifest, lock, review policy, and non-plugin assets;
2. run the deterministic install without lifecycle scripts;
3. stage both plugin sources outside `.opencode/plugins/`;
4. run provider-free import and tool-definition probes;
5. atomically move source into `.opencode/plugins/`; and
6. register both plugins in `opencode.json` (the project `plugin` array)
   rather than relying on plugin auto-discovery.

On failure, plugin sources are absent from `.opencode/plugins/`. Other assets
remain and later independent subtools continue. The plugin subtool has
`name: review-plugins`, `action: failed`, and `activation: inactive`.
`Result.Status` is `partial`, `FailedSubTools` increases by one, and
successful file counts remain populated. The CLI prints one failed and
inactive subtool plus remediation, returns nil, and therefore exits 0.
Rerun retries safely. Fatal core scaffold I/O still returns an error
and exits non-zero. `uf doctor` verifies manifest and lock, runtime
versions, dependencies, registration state, and plugin load.

Version output permits one terminal LF or CRLF, then trims ASCII space
and tab at both edges. The remaining value MUST match anchored ASCII
Node regex `^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`
or the same npm regex without `v`. Leading-zero components, Unicode
digits, signs, embedded whitespace, prerelease, build metadata, extra
lines, and components above the platform unsigned 32-bit limit are
invalid. Parsing checks length and overflow before integer conversion.

This local plugin does not restore the removed swarm plugin. It is a
separate review execution capability and does not replace MCP tools.

The approved commit-scope amendment permits scaffold assets and tests
required by and traced to this active change to ship atomically on its
feature branch. Unrelated scaffold refreshes remain separate, and every
file MUST still be staged explicitly.

### D13: Justify and constrain JavaScript dependencies

`@opencode-ai/plugin` is required for the supported plugin, tool, and
client ABI; hand-writing that protocol is not stable. Zod is required
by
the plugin tool schema API. Vitest supplies TypeScript ESM mocking and
threshold
enforcement unavailable from the Go toolchain.
`@vitest/coverage-v8` is its required coverage adapter.

Node's built-in runner plus custom coverage parsing was rejected
because it would create project-owned test infrastructure and weaker
branch thresholds. All four direct packages use MIT licenses and active
upstream projects. Exact versions and npm lock integrity hashes protect
resolution. Delivery rechecks licenses and integrity. The tracked
baseline lock has SHA256
`b628c764236019785a620deadc06373949236ce84fb38bc68a39f619411874e3`.
Two transitive dependencies are pinned via `overrides` (toml `4.2.0`
and uuid `13.0.1`) to hold known-compatible versions against upstream
floating ranges and reduce resolution drift; they are rechecked under
the same integrity and license discipline as the direct packages.
Rollback removes plugin activation and npm scripts, restores host-model
dispatch and that lock, removes the new manifest, and restores
its prior ignore rule.

### D14: Use measurable isolated test layers

TypeScript unit tests cover parsing, validation, request creation,
redaction, cancellation, result extraction, and policy fixtures. Vitest
coverage MUST reach 90 percent statements and 85 percent branches.
The `plugin-test` Make target MUST run locked installation, unit,
integration, smoke, and coverage checks. `make check` and Local CI MUST
invoke that target so threshold failures block delivery.

Integration tests use a fake OpenCode client and scratch config for
tool definition and child-session contracts. A provider-free smoke
test loads the staged plugin. No test uses provider, GitHub, Dewey, or
package-registry network access.

New Go helper scopes MUST reach 80 percent statement coverage.
Pure mapping, config, artifact, and schema scopes MUST reach
90 percent.
`internal/artifacts` and `internal/schemas` are classified helper
(80 percent), not pure artifact/schema (90 percent), because they
perform filesystem I/O (scanning, reading, and writing artifact and
schema files) rather than pure serialization or generation. The
artifact and schema scopes remain defined at 90 percent for future
pure-only packages.
A checked-in manifest maps each scope to its threshold. The gate
derives changed production Go functions from the PR base SHA supplied
by CI, or the local `main...HEAD` merge base. Every changed function
outside test, generated, and vendored files MUST match exactly one
manifest scope.
Omission or multiple matches fail. One additive `coverage-gate` target
runs race-enabled uncached tests and retains the existing global 80 and
backlog 90 checks. It fails absent, non-numeric, or below-threshold
entries. `make check` and Local CI invoke it.
Existing function CRAP and GazeCRAP MUST NOT increase; new function
CRAP MUST be at most 30.

Command contracts use table-driven tests. Schema positive and negative
fixtures, package-lock consistency, exact asset lists, and canonical
mirror drift tests are mandatory. The existing pipeline harness remains
green.

## Risks / Trade-offs

### R1: Fan-out raises latency and cost

Limits, batching, opt-in augmentation, deduplication, timeout, and cost
ceilings bound exposure. Unknown cost cannot bypass run-count limits.

### R2: Provider variant behavior differs

Variant provenance is always retained. Runtime rejection becomes a
sanitized run error and participates in no-success handling.

### R3: Independent models disagree

Workflow-specific consolidation, one vote per persona, dissent
provenance, and deterministic deduplication make disagreement visible.

### R4: Sibling content can carry prompt injection

Identity checks, commit pinning, path and size bounds, hashes, explicit
untrusted delimiters, and grounded lesson validation constrain impact.

### R5: Offline installation can delay activation

The scaffold remains repairable and idempotent. It never leaves an
unloadable plugin in the auto-discovery path.

### R6: The integrated scope creates a broad diff

Requirement IDs, traced tasks, isolated phases, asset drift tests, and
the approved full-parity boundary prevent unrelated expansion.
