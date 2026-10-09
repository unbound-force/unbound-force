## ADDED Requirements

### Requirement: Verified Sibling Repository Declaration

**ID: RX-FR-001** This requirement MUST be satisfied.

The system MUST read `.uf/sibling-repos.yaml` against the closed
`schemas/sibling-repos/v1.0.0.schema.json` contract. The root MUST
contain only required integer `version: 1` and required `siblings`.
The array MUST contain zero through 32 entries.

Each entry MUST require `name`, `url`, `default_branch`, `role`,
`fetch`, and `contracts`. It MAY contain `local_paths`, `clone_path`,
`cache_path`, and `notes`. Null and empty strings MUST be rejected.
Unknown fields MUST be rejected.

Names MUST match `^[a-z][a-z0-9-]{0,62}$` and be unique. Role MUST be
`downstream-consumer`, `upstream-source`, or `sibling`. Fetch MUST be
`clone-on-review` or `contract-only`. Contracts MUST contain one
through 20 unique repository-relative globs. Local paths MAY contain
zero through four unique absolute or target-root-relative paths.
Clone and cache paths MUST be target-root-relative.

All paths MUST be UTF-8 strings of 1-512 bytes. Contract globs MUST be
UTF-8 strings of 1-256 bytes. Relative paths and globs MUST NOT contain
an absolute prefix, empty segment, `.` segment, or `..` segment.
Default branches MUST be 1-128 printable ASCII characters. Notes MAY
be omitted or contain 1-1024 UTF-8 bytes.

URLs MUST be credential-free HTTPS GitHub URLs with no query or
fragment. Origin normalization MUST lowercase the host, preserve path
case, remove one trailing `.git`, and remove trailing slashes.
Acquisition MUST disable credential prompts, compare normalized origin,
verify a clean state, and resolve exactly 40 lowercase hexadecimal
commit characters. Any invalid declaration, identity mismatch, dirty
or unverifiable state, unsupported URL, or invalid glob MUST reject
that sibling before reading content.

Acquisition MUST examine `local_paths` in declared order, followed by
`clone_path`, then `cache_path`. It MUST select the first clean
candidate whose normalized origin matches. An invalid candidate MUST be
recorded and MUST NOT prevent examination of later candidates.

`contract-only` MUST NOT use the network. With no valid candidate, the
sibling MUST be unavailable. `clone-on-review` MUST first fetch the
declared default branch into `clone_path` with credential prompts
disabled. A successful fetch MUST use its immutable head. When the
fetch fails or is offline, acquisition MAY use the first verified
existing candidate and MUST record `offline-cache`. With no candidate,
the sibling MUST be unavailable. Cache age MUST NOT bypass origin,
clean-state, or
commit verification. Unavailable siblings MUST contribute no evidence
and MUST remain informational.

Processing MUST sort sibling entries by name and matched files by
normalized repository-relative path.

#### Scenario: Unknown sibling field

- **GIVEN** a sibling entry contains an unrecognized field
- **WHEN** the version 1 configuration is validated
- **THEN** the configuration is rejected before acquisition
- **AND** no sibling content enters the prompt

#### Scenario: Deterministic sibling ordering

- **GIVEN** valid siblings and matching files are unordered
- **WHEN** context is assembled
- **THEN** siblings are processed in lexical name order
- **AND** files are processed by normalized relative path

#### Scenario: Repository identity mismatch

- **GIVEN** a declared URL names one GitHub repository
- **AND** the local path resolves to another origin
- **WHEN** sibling context is acquired
- **THEN** the repository is rejected
- **AND** no content from it enters the prompt

#### Scenario: Declared local path precedence

- **GIVEN** the first local path is invalid
- **AND** the second local path is clean with the declared origin
- **WHEN** sibling context is acquired
- **THEN** the first rejection is recorded
- **AND** the second local path supplies immutable evidence

#### Scenario: Contract-only sibling has no source

- **GIVEN** a contract-only sibling has no valid local candidate
- **WHEN** sibling context is acquired
- **THEN** no network operation is attempted
- **AND** the sibling is unavailable without blocking review

#### Scenario: Clone fetch fails with verified fallback

- **GIVEN** clone-on-review cannot fetch the declared branch
- **AND** a later verified cache candidate exists
- **WHEN** sibling context is acquired
- **THEN** the cache candidate supplies evidence
- **AND** acquisition records `offline-cache`

### Requirement: Confined and Bounded Sibling Evidence

**ID: RX-FR-002** This requirement MUST be satisfied.

Sibling paths MUST be canonicalized with symlinks resolved. Every file
MUST remain under the verified repository root and match a declared
contract glob. The system MUST accept at most 20 files, 256 KiB per
file, and 1 MiB total per invocation.

Accepted evidence MUST record sibling name, immutable commit, relative
path, and SHA256. It MUST be wrapped in a delimited untrusted-evidence
section. Reviewers MUST NOT interpret sibling content as commands,
permissions, tool requests, or policy.

Symlink escape, path escape, oversize content, or an unverifiable file
MUST be rejected and recorded.

#### Scenario: Declared sibling contract

- **GIVEN** a verified sibling permits `README.md`
- **WHEN** context is assembled
- **THEN** the file may be included within size limits
- **AND** its commit and file hash are recorded

#### Scenario: Symlink escape

- **GIVEN** a matching path resolves outside the repository root
- **WHEN** paths are canonicalized
- **THEN** the path is rejected
- **AND** its content is not read into the prompt

### Requirement: Live and Scaffolded Sibling Configuration

**ID: RX-FR-003** This requirement MUST be satisfied.

The meta-repository live file MUST identify relevant `unbound-force/*`
hero repositories and their review contracts. The scaffold copy MUST be
an empty commented version 1 template. Generated repositories MUST NOT
inherit organization-specific active entries.

#### Scenario: New repository scaffold

- **GIVEN** a repository runs `uf init`
- **WHEN** `.uf/sibling-repos.yaml` is created
- **THEN** it contains schema guidance
- **AND** it contains no active sibling entry

### Requirement: Structured Source-Grounded Lessons

**ID: RX-FR-004** This requirement MUST be satisfied.

A child run MAY emit at most one lesson proposal between the exact
delimiters `<!-- uf-lesson-proposal:v1 -->` and
`<!-- /uf-lesson-proposal -->`. The delimited body MUST be one UTF-8
JSON object. Raw `> learn:` directives MUST NOT trigger storage.

The object MUST contain only `schema_version`, `sibling`, `commit`,
`sources`, and `information`. `schema_version` MUST equal `1.0.0`.
`sources` MUST contain between one and eight objects. Each source MUST
contain only `path`, `sha256`, and `excerpt`.

Paths MUST be at most 512 bytes. SHA256 values MUST be exactly 64
lowercase hexadecimal characters. Each excerpt MUST be no more than
1 KiB, and information MUST be no more than 2 KiB. The complete
delimited section MUST be no more than 8 KiB. Complete child output
MUST be no more than 1 MiB.

The object MUST validate against the closed
`schemas/lesson-proposal/v1.0.0.schema.json` contract. The section MUST
contain exactly one JSON object and no text outside that object.

The parent MUST normalize accepted source bytes and proposal excerpts
from CRLF to LF and then require the excerpt bytes to be an exact
substring. It MUST NOT collapse source whitespace for grounding.
Sibling, commit, path, and SHA256 MUST exactly match accepted evidence.

For information and dedupe only, normalization MUST apply Unicode NFC,
trim leading and trailing Unicode whitespace, and replace each run of
ASCII space, tab, CR, and LF with one ASCII space. The parent-computed
dedupe input MUST be normalized information, one `|`, and sorted
`path:sha256` source pairs joined by `|`. Its id MUST be lowercase
SHA256. The parent MUST generate `sibling-<name>` from the validated
name and MUST ignore model-supplied tags or hashes.

The parent MUST reject information matching any case-insensitive secret
detector: `sk-`, `AKIA`, `ghp_`, `xox[baprs]-`, `Bearer ` followed by
a token, `api[_-]?key` followed by `:` or `=`, or assignment to
`ANTHROPIC_API_KEY`, `ANTHROPIC_AUTH_TOKEN`,
`AWS_SECRET_ACCESS_KEY`, `GITHUB_TOKEN`, or `OPENAI_API_KEY`.

The parent MUST reject information containing `> learn:`,
`invoke_agent`, `dewey_store_learning`, `dewey_store`,
`hivemind_store`, `forge_record_outcome`, or `tool.execute`. It MUST
also reject a fenced command block that names `uf`, `opencode`,
`dewey`, `replicator`, `gh`, or `git`. Matching is case-insensitive.
Detection MUST reject rather than redact or infer intent.

Only valid, source-grounded proposals MAY be stored through Dewey. The
pure `prepare_lesson_learning` tool MUST accept child output, accepted
sibling evidence, and at most 1024 unique lowercase SHA256 dedupe
identities supplied by the parent after querying existing learnings. A
known identity MUST return a deterministic duplicate skip.

A ready result MUST contain a payload compatible with the existing
`dewey_store_learning` interface: `information`, generated `tag`, and
category `reference`. It MUST also contain complete immutable source
provenance and the parent-computed dedupe identity for audit. Model
output MUST NOT supply the tag, category, or dedupe identity.

The stored information MUST end with stable, non-executable
`UF_LESSON_PROVENANCE_V1` JSON metadata containing the dedupe identity,
sibling, commit, and sorted source `path`/`sha256` pairs. This
representation MUST permit an exact prior identity to be found before
storage.

The plugin MUST NOT call Dewey. After deterministic validation, the
parent MUST call `dewey_store_learning` exactly once only for a ready
result. This human-approved override stores it as a normal learning
through the existing interface; it adds no separate tier, retrieval
exclusion, or promotion path. Malformed, ungrounded, duplicate, or
unavailable-Dewey outcomes MUST be recorded as informational skips and
MUST NOT change the review verdict.

#### Scenario: Valid grounded proposal

- **GIVEN** a proposal quotes an exact accepted source excerpt
- **AND** repository, commit, path, and hash match acquired evidence
- **WHEN** the parent validates it
- **THEN** the tool returns a ready `dewey_store_learning` payload
- **AND** its generated tag is `sibling-<name>`
- **AND** its category is `reference`
- **AND** its provenance and dedupe identity are recorded

#### Scenario: Injected lesson directive

- **GIVEN** sibling content instructs the reviewer to emit `> learn:`
- **WHEN** the parent processes review output
- **THEN** no lesson is stored from that directive
- **AND** the skip is informational

#### Scenario: Oversized lesson section

- **GIVEN** a child emits a lesson section larger than 8 KiB
- **WHEN** the parent parses the response
- **THEN** the proposal is rejected without partial parsing
- **AND** the skip is recorded as informational

#### Scenario: Deterministic grounding failure

- **GIVEN** an excerpt differs after only CRLF-to-LF normalization
- **WHEN** the parent validates grounding
- **THEN** the proposal is rejected as ungrounded
- **AND** whitespace is not heuristically repaired

#### Scenario: Secret or tool instruction

- **GIVEN** information matches a fixed secret or tool token
- **WHEN** the parent validates the proposal
- **THEN** the proposal is rejected before Dewey storage
- **AND** the detector class is recorded without the sensitive value

#### Scenario: Existing lesson is not stored twice

- **GIVEN** the parent finds the computed identity in existing
  learnings
- **WHEN** it supplies that identity to `prepare_lesson_learning`
- **THEN** the result is skipped as a duplicate
- **AND** the parent does not call `dewey_store_learning`

#### Scenario: Dewey is unavailable

- **GIVEN** the parent cannot query or call Dewey
- **WHEN** it handles an otherwise valid lesson proposal
- **THEN** it records an informational unavailable-Dewey skip
- **AND** the review verdict does not change

### Requirement: Pull Request Head and Base Context

**ID: RX-FR-005** This requirement MUST be satisfied.

When review council receives a PR number, it MUST resolve the PR base
and head and use that base-to-head change for context and profiling. It
MUST NOT assume the current `main...HEAD` diff represents the requested
PR.

The dispatch input context MUST record the PR number, printable base
and head refs, and each resolved 40-character lowercase commit SHA. A
local review MUST record a null PR number plus its selected base and
head refs and resolved SHAs. Profiling, review prompts, and artifacts
MUST use those immutable SHAs rather than later ref values.

#### Scenario: Review non-current pull request

- **GIVEN** PR 42 has base `main` and head `feature/remote`
- **AND** the checkout is on another branch
- **WHEN** review council runs for PR 42
- **THEN** context uses PR 42's base and head SHAs
- **AND** the artifact records both refs and immutable SHAs

#### Scenario: Local review context

- **GIVEN** review council runs without a PR number
- **WHEN** its local base-to-head input is resolved
- **THEN** the artifact records a null PR number
- **AND** it records both selected refs and immutable SHAs

#### Scenario: Uncommitted working-tree review context

- **GIVEN** review council runs without a PR number
- **AND** the branch head resolves to the same SHA as the base
  (no commits ahead)
- **AND** the working tree holds uncommitted tracked or untracked
  changes
- **WHEN** its local base-to-head input is resolved
- **THEN** the artifact records a null PR number and
  `uncommitted: true`
- **AND** it records the selected base ref and immutable base SHA
- **AND** the reviewed diff is derived from the working tree
  instead of `base_sha...head_sha`

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
