## ADDED Requirements

### Requirement: Explicit-First Review Matrix

**ID: RD-FR-001** This requirement MUST be satisfied.

The system MUST load the closed version 2 object from
`.uf/review-matrix.yaml` against
`schemas/review-matrix/v2.schema.json`. The root MUST require
`version`,
`profiles`, and `defaults`. It MAY contain only `always`, `runs`,
`advisor`, `risk_augmentation`, and `limits` beyond required fields.
Unknown fields MUST be rejected at every level.

`version` MUST equal integer 2. `profiles` MUST contain 1-32 unique
keys matching `^[a-z][a-z0-9-]{0,31}$`. Each value MUST be a closed
object requiring `model` and optionally containing `variant`. Models
MUST be 3-256 ASCII characters and contain a provider followed by one
or more non-empty slash-separated model segments. Each segment MUST
match `[A-Za-z0-9][A-Za-z0-9._-]*`. Whitespace, query, fragment,
trailing slash, and repeated slash MUST be rejected.
Variants MUST match `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`.

`defaults` MUST have exactly `code`, `spec`, `triage`, and `feedback`.
Each value MUST reference a declared profile. The shipped profiles
MUST be `lightweight`, `standard`, and `heavy`.

`always` MAY contain at most 32 unique agent keys. Each key MUST match
`^divisor-[a-z0-9-]{1,63}$` and reference one profile. It applies only
to generated runs. `advisor` MAY contain only the four matrix modes.
Each mode MUST contain 0-32 unique matching agent names. Membership
opts an agent into generated runs for that mode.

`runs` MAY contain only the four matrix modes. Each mode MAY contain
at most 32 matching agent keys. Each key MUST map to 1-5 ordered run
objects. A run MUST contain exactly one of `profile` or `model` and
MAY contain `variant`. Profile references MUST resolve. Direct models
and variants MUST satisfy the profile constraints. A run variant MUST
override the referenced profile variant.

`risk_augmentation` MUST be a closed object requiring `enabled` and
`agents`. `enabled` MUST be boolean. `agents` MUST contain 0-32 unique
matching agent names. An omitted object MUST mean disabled.

`limits` MUST be closed and MAY use only `max_personas`,
`max_runs_per_persona`, `max_total_runs`, `max_parallel_runs`,
`per_run_timeout_seconds`, and `max_reported_cost_usd`. Values and
defaults MUST follow RD-FR-004.

The command mode `specs` MUST normalize once to matrix mode `spec`.
The matrix key `specs` MUST be rejected. Human reports and artifacts
MUST retain the command mode rather than the normalized lookup key.

Explicit runs MUST be authoritative and MUST replace generated runs by
default. Risk augmentation MUST require an explicit opt-in. Advisor
generation MUST require a separate matching rule. An agent absent from
explicit and advisor configuration MUST receive one host-model run.

Generated runs MUST support adversary-heavy and SRE-lightweight
overrides. Overrides MUST NOT rewrite explicit runs.

Identical `(model, variant)` pairs MUST deduplicate while preserving
the first occurrence. Invalid entries MUST become run errors and MUST
NOT trigger silent substitution.

#### Scenario: Ordered explicit runs replace generation

- **GIVEN** adversary has two ordered explicit runs
- **AND** adversary also matches an advisor rule
- **WHEN** the matrix is resolved without augmentation opt-in
- **THEN** the two explicit runs appear in declared order
- **AND** no advisor-generated run is added

#### Scenario: Host-model fallback

- **GIVEN** a discovered agent has no explicit runs
- **AND** it does not match an advisor-generation rule
- **WHEN** its plan is created
- **THEN** it receives exactly one host-model run
- **AND** it is not dropped because a mode default exists

#### Scenario: Invalid explicit item

- **GIVEN** an explicit run names an unresolved profile
- **WHEN** the matrix is resolved
- **THEN** that item is recorded as a validation error
- **AND** another model is not silently substituted

#### Scenario: Specs mode normalization

- **GIVEN** review council receives command mode `specs`
- **WHEN** matrix defaults and explicit runs are resolved
- **THEN** the resolver reads only the canonical `spec` key
- **AND** a matrix containing `specs` fails closed validation

### Requirement: Deterministic Dispatch Plan

**ID: RD-FR-002** This requirement MUST be satisfied.

The advisor MUST emit plan version 1. Agents MUST be ordered by name.
Runs MUST use stable sequence. Every entry MUST include agent,
decision, reason code, reason, source, sequence, and read-only intent.
It MUST include nullable tier, model, and variant plus validation
errors. Source MUST be `explicit`, `advisor`, or `host`.

The planner MUST load `.uf/reviewer-capabilities.yaml` against the
closed `schemas/reviewer-capabilities/v1.0.0.schema.json` contract.
The root MUST contain only integer `version: 1` and a non-empty
`reviewers` array.
Entries MUST contain only unique `agent`, `capability`, and `scopes`.
Agent names MUST match `^divisor-[a-z0-9-]{1,63}$`. Capability MUST be
`review` or `content`. Scopes MUST be ordered unique Protocol 3
category values. Review scopes MUST be non-empty; content scopes MUST
be empty.
Unknown fields and YAML aliases MUST be rejected.

The shipped table MUST contain these exact entries and ordered scopes:

- `divisor-adversary`: review; security, dependencies, standard;
- `divisor-architect`: review; standard, cli-ux, ci-cd, documentation;
- `divisor-curator`: review; documentation;
- `divisor-guard`: review; standard, cli-ux, documentation;
- `divisor-sre`: review; ci-cd, dependencies, security;
- `divisor-testing`: review; test-quality; and
- envoy, herald, and scribe: content with empty scopes.

Adversary and guard MUST always be included. Other review agents MUST
intersect classified categories. Curator MUST be included only for
documentation or user-facing changes. Content agents MUST NOT run. An
unknown agent MUST have an explicit valid manifest entry. Invalid data
MUST fail planning as `INCONCLUSIVE`. Agent frontmatter MUST NOT carry
this metadata. An OpenCode-load fixture MUST prove provider options
stay unchanged.

The `review-dispatch` policy plugin, registered in `opencode.json`, MUST expose
`plan_review_dispatch`. It MUST own matrix and manifest parsing,
relevance, limits, and byte-stable plan validation. A malformed plan
MUST stop dispatch as `INCONCLUSIVE` before any child session starts.

For issue triage, the planner MUST NOT require or synthesize a diff.
It MUST normalize issue title, nullable body, and comments as UTF-8
NFC. It MUST convert CRLF and bare CR to LF. Comments MUST be
ordered by creation time, then numeric comment id.

The hash input MUST use the exact `uf-issue-content-v1` UTF-8 frame.
Lengths and counts MUST use canonical ASCII decimal without leading
zeros. The frame MUST start with these bytes:

```text
uf-issue-content-v1\n
title:<bytes>\n<title-bytes>\n
body:null\n
```

For a present body, including empty, `body:null\n` MUST be replaced by
`body:<bytes>\n<body-bytes>\n`. The frame MUST then contain:

```text
comments:<count>\n
comment-id:<bytes>\n<decimal-id>\n
created-at:<bytes>\n<UTC-RFC3339-seconds>\n
comment-body:<bytes>\n<comment-bytes>\n
```

The three comment fields MUST repeat for every ordered comment. SHA256
MUST cover every framing and content byte. `text_bytes` MUST count only
normalized title, present body, and comment body bytes. It MUST exclude
labels, lengths, ids, timestamps, and separator LFs. A null body MUST
contribute zero text bytes. It MUST hash differently from an empty
body.

The planner MUST ASCII-case-fold and tokenize maximal `[a-z0-9]+`
sequences. A closed checked-in rule table MUST map tokens to Protocol 3
categories as follows:

- security: security, vulnerability, cve, secret, credential, token,
  auth, permission, injection;
- cli-ux: cli, command, flag, argument, output, prompt, terminal, ux;
- test-quality: test, testing, coverage, assertion, flaky, regression,
  fixture;
- documentation: documentation, docs, readme, guide, tutorial,
  changelog;
- ci-cd: ci, workflow, pipeline, action, deploy, release, build; and
- dependencies: dependency, dependencies, package, module, version,
  upgrade, supply-chain.

The two-token sequence `supply chain` MUST match `supply-chain`.
Unmatched content MUST receive only `standard`. Rule ids MUST use
`<category>:<keyword>`, be unique, and sort lexically. A security match
MUST set `security_sensitive`.

#### Scenario: Stable plan ordering

- **GIVEN** the same agents, matrix, and change profile
- **WHEN** the advisor creates the plan twice
- **THEN** both serialized plans are byte-equivalent
- **AND** agents and sequences have stable ordering

#### Scenario: Malformed plan

- **GIVEN** a plan entry lacks a required reason code
- **WHEN** the parent validates the plan
- **THEN** no child session starts
- **AND** the workflow result is `INCONCLUSIVE`

#### Scenario: Fixed issue-content hash

- **GIVEN** title is `Coverage regression`
- **AND** body is null and there are no comments
- **WHEN** the version 1 frame is hashed
- **THEN** the SHA256 is
  `60a7be4394be0186439a588e93c57dc9095c1b96f42f27464dbfeb2ce5cfdbf5`
- **AND** `text_bytes` is 19 rather than the 70 framed bytes

### Requirement: Total Tier Classification

**ID: RD-FR-003** This requirement MUST be satisfied.

The advisor MUST assign heavy first for a security-sensitive change,
has more than 300 changed lines, or has more than 10 changed files. It
MUST otherwise assign standard when lines exceed 50, files exceed 3, or
component span exceeds 1. It MUST assign lightweight otherwise.

Component span MUST count normalized parent directories. Root files
use `<root>`. Tests SHOULD be grouped with their source directory when
that relationship is identifiable.

Issue triage MUST use content metrics instead of diff metrics. It MUST
assign heavy for security, more than 32768 normalized UTF-8 bytes, or
more than 20 comments. It MUST otherwise assign standard for more than
4096 bytes, more than five comments, or more than one category. It MUST
assign lightweight otherwise. The profile MUST record text bytes,
comment count, input hash, categories, matched rule ids, security
state, and tier.

Boundary fixtures with null body and no comments MUST use an ASCII `x`
title of the named length. Their lowercase SHA256 and tier MUST be:

- 4096 bytes:
  `4f30e0423cec84abfc13ae44c9fe73dde044a0852c5492884e52ba020f7b8275`,
  lightweight;
- 4097 bytes:
  `19c10c78070cae1fb718628a7e276ed1b9d05d7d3f017bfb32d09bc309aa7207`,
  standard;
- 32768 bytes:
  `5e06dbe70b16a7d00fb864d511e8542bfdbc1473275d33076f33d5a917be0168`,
  standard; and
- 32769 bytes:
  `15b20fab5e843a6526a81d7809f0c847f1aa237408cd6351b8c2c997d71ae3fe`,
  heavy.

#### Scenario: Small multi-component change

- **GIVEN** a non-security change has 40 lines in 3 files
- **AND** the files span two normalized components
- **WHEN** the advisor computes the tier
- **THEN** the tier is standard

#### Scenario: Security overrides size

- **GIVEN** a security-sensitive change has 10 lines in one file
- **WHEN** the advisor computes the tier
- **THEN** the tier is heavy

#### Scenario: Issue content selects reviewers and tier

- **GIVEN** an issue title contains `coverage regression`
- **AND** its normalized content is at most 4096 bytes with no comments
- **WHEN** triage planning classifies the issue
- **THEN** category is `test-quality` and tier is lightweight
- **AND** the testing reviewer is eligible

#### Scenario: Large issue discussion

- **GIVEN** normalized issue content has six comments
- **WHEN** triage planning classifies the issue
- **THEN** tier is standard even when one category matches

### Requirement: Bounded Fan-Out and Execution

**ID: RD-FR-004** This requirement MUST be satisfied.

Risk fan-out MUST be opt-in. Adjacent direction MUST be lightweight to
standard, standard to heavy, and heavy to standard. Security-sensitive
adversary review and architect review above 500 lines MAY add one
adjacent run. Duplicate model and variant pairs MUST be omitted.

Curator MUST have at most one assessment run. Explicit arrays or
augmentation that would create more MUST fail planning. Curator
prompts MUST forbid issue creation. Only the parent MAY perform one
deduplicated curation action after consolidation and the existing
human gate.

Default limits MUST be 16 personas, 3 runs per persona, 24 total runs,
4 parallel runs, 600 seconds per run, and USD 25 in reported cumulative
cost. Cost MUST be checked between batches. Run-count limits MUST apply
when cost is unknown. Parent cancellation MUST cancel active children.
Limit and budget skips MUST be recorded.

Configuration MUST remain within 1-32 personas, 1-5 runs per persona,
1-64 total runs, 1-8 parallel runs, 30-1800 seconds per run, and a
USD 1-100 cost ceiling. Out-of-range configuration MUST fail validation
before dispatch.

#### Scenario: Adjacent duplicate

- **GIVEN** heavy and standard profiles resolve to the same pair
- **WHEN** heavy risk augmentation selects standard
- **THEN** the duplicate adjacent run is omitted
- **AND** the omission is recorded in plan provenance

#### Scenario: Cost ceiling between batches

- **GIVEN** completed batches report USD 25 in cumulative cost
- **WHEN** another batch is pending
- **THEN** the pending runs are not started
- **AND** each budget skip is recorded

### Requirement: Full Panel Escape Hatch

**ID: RD-FR-005** This requirement MUST be satisfied.

Advisor commands accepting `--full` MUST include every discovered
review-capable persona once at the standard profile. They MUST disable
pruning, explicit fan-out, and risk augmentation for that invocation.
Content-capability personas MUST remain excluded. The plan and artifact
MUST record full-panel mode.

Full-panel mode MUST NOT truncate personas. If the discovered panel
exceeds the configured persona or total-run limit, the plan MUST fail
before dispatch, return `INCONCLUSIVE`, and report the required and
configured counts. Absolute configuration bounds from RD-FR-004 MUST
still apply.

#### Scenario: Full panel requested

- **GIVEN** six valid review-capable Divisor agents are discovered
- **WHEN** the command runs with `--full`
- **THEN** all six agents run exactly once at standard
- **AND** no second model is added

### Requirement: Defensive Explicit-Model Invocation

**ID: RD-FR-006** This requirement MUST be satisfied.

The `invoke-agent` plugin MUST expose `invoke_agent`. It MUST accept
an allowlisted manifest-declared `divisor-*` review agent, prompt,
optional model, optional variant, and optional boolean `read_only`
metadata.

Agent length MUST be at most 128 characters, model length 256 when
present, variant length 64, and prompt size 128 KiB UTF-8. A present
model MUST split at the first slash. The provider prefix and complete
model-id remainder MUST both be non-empty. No other RD-FR-001 model
grammar MUST apply at this boundary. The plugin MUST reject unknown
agents, edge-position slashes, other malformed present values, or
oversize input before creating a child session.

An omitted requested model MUST identify host resolution. For a
validated `source: host` entry, the plugin MUST call
`client.session.message` with `context.sessionID` and
`context.messageID`. It MUST require the current message to have
assistant role, non-empty actual `providerID` and `modelID`, and a
valid active variant. It MUST replay that resolved pair and variant
explicitly in the child prompt. Explicit and advisor entries MUST
provide their requested model normally.

The installed v1 response type omits the runtime assistant variant.
The plugin MAY use one narrow compatibility type that adds only the
optional variant at the current-message response boundary. It MUST use
the separate narrow prompt-body compatibility type to send variant at
the prompt-body top level. Neither type MAY widen other client or
message fields.

Current-message lookup failure, a non-assistant current message, an
absent provider or model id, or an absent or invalid active variant
MUST fail the run as a sanitized retryable availability error. The
plugin MUST NOT substitute project, agent, provider, or other defaults.

The child MUST be parented to the caller and receive timeout and
cancellation. It MUST use the selected agent's existing OpenCode
permission contract. The plugin MUST NOT broaden, narrow, replace, or
reinterpret those permissions. The `read_only` value MUST be retained
as invocation metadata for compatibility with the validated reference.
Its value MUST NOT allow, deny, or otherwise change invocation.

Variant MUST be sent at the prompt-body top level and retained as
provenance. One `invoke_agent` path MUST handle explicit, advisor, and
host runs with the same timeout, cancellation, response extraction,
usage, redaction, and provenance behavior. Usage MUST be nullable.

Errors MUST redact credential values, bearer tokens, URL query strings,
and absolute home paths. The plugin MUST NOT select policy, consolidate
findings, or write artifacts.

#### Scenario: Valid explicit invocation

- **GIVEN** an allowlisted agent and bounded valid request
- **WHEN** `invoke_agent` runs
- **THEN** the child uses the requested model and variant
- **AND** output retains authoritative request provenance

#### Scenario: Host-model invocation

- **GIVEN** a validated host-source plan has no requested model
- **AND** the current assistant has an actual model and active variant
- **WHEN** `invoke_agent` resolves and creates the child prompt
- **THEN** the prompt explicitly replays that model pair and variant
- **AND** requested model and variant provenance remain null
- **AND** resolved parent and reported child provenance are recorded

#### Scenario: Host-model resolution is unavailable

- **GIVEN** a host run cannot resolve its current assistant variant
- **WHEN** `invoke_agent` prepares the child request
- **THEN** the run fails with a sanitized availability error
- **AND** no configured default model or variant is substituted

#### Scenario: Direct validation is intentionally minimal

- **GIVEN** `provider/model?candidate=1` has non-empty slash sides
- **WHEN** direct invocation validation runs
- **THEN** validation accepts the bounded model string
- **AND** matrix validation still rejects it under RD-FR-001

#### Scenario: Read-only metadata is false

- **GIVEN** a valid review invocation records `read_only: false`
- **WHEN** the plugin validates the request
- **THEN** the value is retained in invocation provenance
- **AND** it does not allow or deny session creation

#### Scenario: Existing permissions are preserved

- **GIVEN** an allowlisted persona has an existing permission contract
- **WHEN** the plugin creates its child invocation
- **THEN** no permission override is added by the plugin
- **AND** OpenCode continues to enforce the selected agent contract

#### Scenario: Sensitive provider error

- **GIVEN** a provider error contains a token and home path
- **WHEN** the plugin returns the run error
- **THEN** both values are redacted
- **AND** model and variant provenance remain available

### Requirement: Fail-Closed Run Consolidation

**ID: RD-FR-007** This requirement MUST be satisfied.

Failed runs MUST be recorded and MUST NOT vote, create findings, or
increase severity. Successful sibling runs MUST continue. Each workflow
MUST require at least one successful assessment before an approving,
validating, or accepting result.

When no successful assessment remains, the workflow MUST apply this
cause precedence and block automation. Any policy, plan, budget,
limit, persistence, or calculation cause MUST produce `INCONCLUSIVE`.
It MUST win over availability causes in a mixed failure. `UNAVAILABLE`
MUST apply only when every blocking cause is provider, model, or
runtime availability.
Both results MUST request retry or human review.

#### Scenario: Partial failure

- **GIVEN** one model fails and one model succeeds
- **WHEN** results are consolidated
- **THEN** the failure is informational
- **AND** the successful assessment determines the eligible result

#### Scenario: Every provider is unavailable

- **GIVEN** all planned runs fail only from provider unavailability
- **WHEN** results are consolidated
- **THEN** the result is `UNAVAILABLE`
- **AND** the workflow cannot report approval

#### Scenario: Mixed no-success causes

- **GIVEN** one run has a provider availability failure
- **AND** another run is precluded by an execution limit
- **WHEN** results are consolidated
- **THEN** the result is `INCONCLUSIVE`
- **AND** the availability failure remains in provenance

### Requirement: Review Dispatch Artifact

**ID: RD-FR-008** This requirement MUST be satisfied.

Every dispatch MUST write a version 1.0.0 `review-dispatch` payload.
It MUST use a Hero Interface Contract envelope that identifies
`the-divisor`, producer version, UTC timestamp, artifact and schema
versions, branch, commit, correlation id, and workflow id.

`review-dispatch` MUST be additive execution provenance. Canonical
review decisions MUST use `review-verdict` schema 2.0.0 under
RV-FR-001. Registry documentation, samples, producer and consumer
tables, and compatibility tests MUST describe dual emission and the
major-version consumer migration.

The closed payload MUST require `command`, `mode`, `full`,
`input_context`, `change_profile`, `plan_version`, `plan`, `runs`,
`coverage`,
`findings`, `advisories`, `verdict`, `verdict_reason`, `run_counts`,
`workflow_result`, and `correlation_id`. Unknown properties MUST be
rejected at every object level.

`command` MUST be `review-council`, `triage-issue`,
`address-feedback`, or `speckit-testreview`. `mode` MUST be `code`,
`specs`, `triage`, `feedback`, or `test`.

`input_context` MUST be a closed discriminated object. PR and local
kinds MUST require `kind`, `pr_number`, `base_ref`, `base_sha`,
`head_ref`, and `head_sha`. Refs MUST be 1-255 printable ASCII
characters. SHAs MUST be exactly 40 lowercase hexadecimal characters.
PR kind MUST have an integer PR number from 1 through 999999. Local
kind MUST have null PR number. The resolved SHAs MUST identify the
reviewed diff.

Issue kind MUST require only `kind`, `issue_number`, `issue_url`, and
`content_sha256`. Issue number MUST be from 1 through 999999. URL MUST
be a canonical credential-free HTTPS GitHub issue URL. Content hash
MUST be 64 lowercase hexadecimal characters. Git refs and SHAs MUST be
absent.

`change_profile` MUST be closed and discriminated by `kind`. Both kinds
MUST require boolean `security_sensitive`, unique sorted `categories`,
and tier. Categories MUST use the review-context Protocol 3 enum. Tier
MUST be `lightweight`, `standard`, or `heavy`.

Diff kind MUST require non-negative integer `lines`, `files`, and
`components`. Issue kind MUST instead require non-negative integer
`text_bytes` and `comment_count`, 64-character lowercase
`content_sha256`, and sorted unique `matched_rules`. Fields from the
other profile kind MUST be absent.

`plan_version` MUST equal 1. Each closed plan entry MUST use the
RD-FR-002 fields. Decision MUST be `include` or `skip`. Source MUST be
`explicit`, `advisor`, or `host`. Sequence MUST be a positive integer.
Tier, model, and variant MUST be nullable. Validation errors MUST be
unique bounded strings.

Each planned run MUST appear once in `runs`. A closed run MUST require
`run_id`, `agent`, `source`, `requested_model`, `provider`, `model_id`,
`variant`, `resolved_parent_model`, `resolved_parent_variant`,
`sequence`, `status`, `started_at`, `finished_at`, `usage`, `error`,
`reported_model`, `model_mismatch`, `workflow_verdict`, and `findings`.
Run ids MUST be unique UUIDs. Sequence MUST be positive.

For host source, requested model, provider, model id, and variant MUST
be null. Resolved parent model and variant MUST be null until lookup
succeeds. After successful lookup, both MUST be non-null and MUST form
the exact pair and active variant replayed to the child. For explicit
or advisor source, requested model, provider, and model id MUST be
non-null and form the same model identity; both resolved-parent fields
MUST be null. A successful host run MUST report its child model. A
reported model mismatch against the requested or resolved pair MUST
make the run failed.

`resolved_parent_model` MUST be null or a 3-256 character model string
formed from the resolved provider and model id with one separating
slash. `resolved_parent_variant` MUST be null or satisfy the RD-FR-001
variant constraint. A host availability failure MAY leave either field
null but MUST include its sanitized error.

Status MUST be `pending`, `running`, `success`, `failed`, `skipped`,
`budget_skipped`, `limit_skipped`, or `cancelled`. Final artifacts MUST
contain only terminal states. Start and finish times MUST be nullable
UTC RFC3339 values, with finish not earlier than start.

Usage MUST be null or a closed object with nullable non-negative
`cost_usd` and nullable `tokens`. Tokens MUST be null or a closed
object with non-negative integer `input` and `output`. Missing usage
MUST remain null. Error MUST be null or a closed object requiring a
bounded `code`, sanitized bounded `message`, and boolean `retryable`.

A successful run MUST have null error and a non-null native verdict.
Council and test verdicts MUST use `APPROVE`,
`APPROVE WITH ADVISORIES`, or `REQUEST CHANGES`. Triage verdicts MUST
use `VALID`, `INVALID`, or `NEEDS-CLARIFICATION`. Feedback verdicts
MUST use `ACCEPT` or `AUTHOR-DECIDES`. Non-success states MUST have a
null workflow verdict and no findings. Failed runs MUST have an error.

Each run finding MUST be closed and require severity, category,
description, root cause, nullable file, and nullable positive line.
Severity MUST be `CRITICAL`, `HIGH`, `MEDIUM`, or `LOW`.
Consolidated findings MUST add a non-empty unique `run_ids` array.
Advisories MUST require severity, description, and contributing ids.

`coverage` MUST be closed and require `preflight_verdict`,
`checks_total`, and `checks_passed`. The verdict MUST be `PASS`,
`FAIL`, `SOFT_GATE`, or `NOT_RUN`. Counts MUST be non-negative,
and passed MUST NOT exceed total.

`run_counts` MUST require non-negative `total`, `success`, `failed`,
`skipped`, `budget_skipped`, `limit_skipped`, and `cancelled` values.
The JSON Schema MUST enforce fields, types, and bounds. A deterministic
semantic validator MUST require each value to equal the number of runs
in that terminal state and their sum to equal `total`. Final artifacts
MUST NOT contain `pending` or `running` runs. Final verdict MUST be
`APPROVE`,
`APPROVE WITH ADVISORIES`, `REQUEST CHANGES`, `INCONCLUSIVE`, or
`UNAVAILABLE`.

`workflow_result` MUST be a closed discriminated object with `kind`
and `value`. Kind MUST be `council`, `triage`, `feedback`, or
`test-review`. Its value MUST use the corresponding native verdict
union above, plus `INCONCLUSIVE` or `UNAVAILABLE`. Partial and failed
runs MUST remain represented.

The native workflow result MUST remain authoritative. Council and
test-review values MUST map identically to the generic final verdict.
Triage MUST map `VALID` to `APPROVE`, `INVALID` to
`REQUEST CHANGES`, and `NEEDS-CLARIFICATION` to
`APPROVE WITH ADVISORIES`. Feedback MUST map `ACCEPT` to `APPROVE`
and `AUTHOR-DECIDES` to `APPROVE WITH ADVISORIES`. Native
`INCONCLUSIVE` and `UNAVAILABLE` MUST map identically.

`UNAVAILABLE` MUST mean zero runnable or successful assessments
because of provider, model, or runtime availability. `INCONCLUSIVE`
MUST mean policy or plan validation failed, execution limits prevented
assessment,
artifact persistence failed, or another condition prevented calculation
of a valid assessment.

The schema MUST encode structural discriminators, types, required
fields, and nullability. The `review-dispatch` policy
plugin MUST expose `finalize_review_dispatch` to enforce cross-field
mappings, terminal-state rules, count arithmetic, and atomic
persistence. It MUST include valid, missing-field, extra-property,
bad-enum, invalid-state, and same-major compatibility fixtures. Every
fixture MUST run through schema and semantic validation. Positive and
negative fixtures MUST cover every native mapping and count invariant.
Compatibility MUST follow the Hero Interface Contract.

#### Scenario: Malformed matrix model path

- **GIVEN** a model has a trailing slash or repeated slash
- **WHEN** matrix validation runs
- **THEN** matrix validation rejects the model before dispatch
- **AND** no provider request is made

Files MUST use a unique correlation-id name under
`.uf/artifacts/dispatch/`. Directories MUST use mode 0750 and files
0600. Writes MUST use a same-directory temporary file. The writer MUST
flush, close, and chmod it before exclusive atomic publication through
a same-filesystem hard link. Link creation MUST fail with `EEXIST` when
the final target exists and MUST NOT overwrite it. After publication,
the writer MUST remove the temporary name. `EEXIST` MUST generate a new
correlation id and retry. Write failure MUST retain the calculated
assessment for human-only reporting. It MUST change the operation
result to `INCONCLUSIVE`, return a failing status, block automated
progression, and MUST NOT fabricate a review finding.

#### Scenario: Mixed run artifact

- **GIVEN** two runs succeed and one fails
- **WHEN** dispatch finalizes its artifact
- **THEN** all three outcomes and final verdict are present
- **AND** unavailable usage is null rather than fabricated

#### Scenario: Native result maps without loss

- **GIVEN** triage returns `NEEDS-CLARIFICATION`
- **WHEN** dispatch finalizes its artifact
- **THEN** the native result remains `NEEDS-CLARIFICATION`
- **AND** the generic verdict is `APPROVE WITH ADVISORIES`

#### Scenario: Terminal run counts reconcile

- **GIVEN** final runs include every supported terminal state
- **WHEN** dispatch calculates run counts
- **THEN** each distinct count matches its run state
- **AND** their sum equals total
- **AND** no pending or running run is present

#### Scenario: Pull request input is reproducible

- **GIVEN** a review targets a pull request not checked out locally
- **WHEN** dispatch writes its final artifact
- **THEN** input context records PR number and base and head refs
- **AND** both immutable resolved commit SHAs are present

#### Scenario: Exclusive publication collision

- **GIVEN** the chosen target exists or appears during publication
- **WHEN** the writer creates the exclusive hard link
- **THEN** it chooses a new correlation id
- **AND** the existing artifact remains unchanged

#### Scenario: Artifact persistence fails

- **GIVEN** a calculated assessment exists
- **WHEN** its required artifact cannot be written
- **THEN** the operation returns failing `INCONCLUSIVE` status
- **AND** the assessment is non-authoritative human context only

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
