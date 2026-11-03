# build-review-prompts

## ADDED Requirements

### Requirement: FR-001 build_review_prompts tool

The review-dispatch plugin MUST register a
`build_review_prompts` tool that deterministically assembles
one self-contained prompt file per requested Divisor agent.

### Requirement: FR-002 Input schema

The tool MUST accept the following input parameters:

- `agents` (string array, required): Agent names to build
  prompts for. Each name MUST match the pattern
  `^divisor-[a-z0-9-]+$`.
- `mode` (string, required): Review mode. MUST be one of
  `code`, `specs`, `triage`, `feedback`, or `test`.
- `command` (string, required): Calling command name. MUST
  be one of `review-council`, `triage-issue`,
  `address-feedback`, `speckit-testreview`.
- `diff_path` (string, required): Absolute path to the
  saved diff file on disk. The tool MUST read the diff
  content from this path. The path MUST be validated per
  FR-011 before reading.
- `changed_files` (string, required): Pre-formatted list
  of changed file paths with addition/deletion counts.
- `input_context` (string, required): Serialized
  base/head ref information (branch names, SHAs).
- `pre_flight_results` (string, optional): Formatted
  pre-flight check output.
- `sibling_evidence` (string, optional): Delimited
  sibling evidence block with provenance markers.
- `existing_reviews` (string, optional): Existing review
  state (comments, reviews) within a token budget.
- `walkthrough` (string, optional): PR walkthrough text.
- `review_context` (string, optional): Spec artifacts,
  linked issues, or other review context.

#### Scenario: Minimal required inputs

- **GIVEN** a valid `agents` array, `mode`, `command`,
  `diff_path`, `changed_files`, and `input_context`
- **WHEN** `build_review_prompts` is called with only
  required fields
- **THEN** the tool produces one prompt file per agent
  containing all 11 required sections, with optional
  sections omitted or empty-marked

#### Scenario: All optional inputs provided

- **GIVEN** all required and optional inputs are provided
- **WHEN** `build_review_prompts` is called
- **THEN** each prompt file includes pre-flight results,
  sibling evidence, existing reviews, walkthrough, and
  review context in their designated sections

### Requirement: FR-003 Output schema

The tool MUST return a JSON object with shape:
```
{
  prompts: Array<{
    agent: string,
    path: string,
    size_bytes: number
  }>
}
```

The `prompts` array MUST contain one entry per requested
agent, in the same order as the input `agents` array.
Each `path` MUST be an absolute path to a readable file
on disk. Each `size_bytes` MUST be the UTF-8 byte length
of the written file.

#### Scenario: Output matches input agent count

- **GIVEN** `agents` contains 5 agent names
- **WHEN** `build_review_prompts` succeeds
- **THEN** `prompts` contains exactly 5 entries, one per
  agent, in the same order

### Requirement: FR-004 Prompt template structure

Each generated prompt file MUST contain these 11 sections
in order:

1. **Step 0 instruction**: An instruction to read the
   agent's own definition file at
   `.opencode/agents/{agent}.md` before conducting the
   review.
2. **Persona role**: The agent's role description and
   mode-specific focus. The tool MUST read the agent
   definition file from `.opencode/agents/{agent}.md` and
   include its full content.
3. **Diff**: The complete immutable diff content, read
   from `diff_path`, enclosed in untrusted content
   delimiters per FR-012.
4. **Changed paths and input context**: The `changed_files`
   and `input_context` content verbatim.
5. **Project context**: The contents of `AGENTS.md` (read
   from disk), the constitution (read from
   `.specify/memory/constitution.md`), all convention pack
   files discovered from `.opencode/uf/packs/`, and the
   severity pack (read from
   `.opencode/uf/packs/severity.md`).
6. **Review evidence**: `pre_flight_results`,
   `review_context`, and `walkthrough` content when
   provided.
7. **Existing review state**: `existing_reviews` content
   when provided.
8. **Sibling evidence**: `sibling_evidence` content when
   provided, with provenance delimiters preserved.
9. **Confinement rule**: Static text restricting findings
   to changed lines and their downstream impact.
10. **Prohibitions**: Static text prohibiting issue
    creation, tool changes, permission changes, repository
    scope changes, and file scope changes.
11. **Response contract**: Static text specifying the
    required response format: findings with severity,
    category, description, root_cause, file, line; model
    self-report; verdict; and optional lesson proposal.

#### Scenario: Agent definition file included

- **GIVEN** `.opencode/agents/divisor-guard.md` exists
  with content "Guard persona definition..."
- **WHEN** `build_review_prompts` is called with
  `agents: ["divisor-guard"]`
- **THEN** the prompt file for `divisor-guard` contains
  the full content of `divisor-guard.md`

#### Scenario: Convention packs discovered from disk

- **GIVEN** `.opencode/uf/packs/` contains `default.md`,
  `go.md`, and `severity.md`
- **WHEN** `build_review_prompts` is called
- **THEN** each prompt file contains the contents of all
  three pack files

#### Scenario: Missing agent definition file

- **GIVEN** `.opencode/agents/divisor-unknown.md` does
  not exist
- **WHEN** `build_review_prompts` is called with
  `agents: ["divisor-unknown"]`
- **THEN** the tool SHOULD produce the prompt with an
  empty persona section and a warning in the output,
  rather than failing entirely

### Requirement: FR-005 File I/O dependencies

The tool MUST read the following files from disk without
agent involvement:

- `.opencode/agents/{agent}.md` for each requested agent
- `AGENTS.md` from the project root
- `.specify/memory/constitution.md`
- All `*.md` files in `.opencode/uf/packs/`
- The diff file at the provided `diff_path`

File reads MUST use the existing `readText` dependency
pattern from the review-dispatch plugin. The tool MUST
NOT require the agent to read and pass these files.

#### Scenario: Project without convention packs

- **GIVEN** `.opencode/uf/packs/` does not exist
- **WHEN** `build_review_prompts` is called
- **THEN** the tool produces prompts without convention
  pack content and does not fail

### Requirement: FR-006 Prompt file output location

When a dispatch session exists (correlation_id resolved
for the current session), the tool MUST write prompt
files to `<dispatch-session-dir>/prompts/{agent}.md`.

When no dispatch session exists, the tool MUST write
prompt files to a new temporary directory under
`$TMPDIR/opencode/prompts-<uuid>/`.

The tool MUST create the output directory if it does not
exist. Created directories MUST have permissions `0o700`
and created files MUST have permissions `0o600` to
prevent information disclosure on shared systems.

Cleanup of prompt files is the caller's responsibility.
The tool does not provide an explicit cleanup mechanism.
Callers MUST remove the temp directory after dispatch
completion (on both success and failure paths). The
`dispatch_agent_run` tool's session lifecycle handles
cleanup for dispatch-session-scoped output. For ad-hoc
`$TMPDIR/opencode/prompts-<uuid>/` paths, callers MUST
remove the directory in a finally/cleanup block to
prevent sensitive content persistence.

#### Scenario: Dispatch session exists

- **GIVEN** a prior `dispatch_agent_run` call established
  a correlation_id for this session
- **WHEN** `build_review_prompts` is called
- **THEN** prompt files are written under the dispatch
  session directory at `prompts/{agent}.md`

#### Scenario: No dispatch session

- **GIVEN** no prior `dispatch_agent_run` call in this
  session
- **WHEN** `build_review_prompts` is called
- **THEN** prompt files are written to a new temp
  directory

### Requirement: FR-007 Dual-copy sync

The tool implementation MUST be maintained in both
`.opencode/plugins/review-dispatch/index.ts` and
`internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`.
Both copies MUST be byte-identical.

### Requirement: FR-008 Tool registration

The tool MUST be registered in
`ReviewDispatchPlugin.server()` alongside existing tools.
The `build_review_prompts` tool MUST appear in the
registered tools list after initialization.

#### Scenario: Plugin registration

- **GIVEN** the review-dispatch plugin is initialized
- **WHEN** the plugin's tool list is inspected
- **THEN** `build_review_prompts` appears in the
  registered tools

### Requirement: FR-009 Error handling

The tool MUST return a structured error result when:

- `diff_path` does not exist or is not readable
- `diff_path` fails path validation (FR-011)
- No agents are provided (empty array)
- An agent name does not match the required pattern
- Output directory cannot be created (permissions, disk)
- A prompt file write fails (partial write, disk full)

The tool MUST distinguish between "file not found"
(graceful degradation) and "I/O error" (permission
denied, disk error) for optional project files. File-
not-found for optional files (AGENTS.md, constitution,
packs) SHOULD produce prompts with those sections empty
and include warnings in the output. I/O errors on
optional files SHOULD surface as warnings but not fail
the tool. I/O errors on required files (diff_path) MUST
fail with a structured error.

#### Scenario: Diff file not readable

- **GIVEN** `diff_path` names an existing file inside the
  allowed scope that cannot be read
- **WHEN** `build_review_prompts` is called
- **THEN** the tool returns a failed result with error
  code `diff_read_failed` and a sanitized error message

A `diff_path` that does not exist is classified by FR-011
before any read: inside the allowed scope it fails with
`diff_not_found`, outside it with `path_validation_failed`.

### Requirement: FR-010 Response contract content

The static response contract section MUST instruct child
agents to:

- Use `submit_review_findings` to submit structured
  findings with severity, category, description,
  root_cause, file, and line
- Use `submit_lesson_proposal` to submit lesson proposals
  with information, tag, and optional category
- Include `**Model**: <family>` self-report in their
  text response
- Include one of the native verdict values in their
  text response

#### Scenario: Response contract mentions tool calls

- **GIVEN** a prompt file is generated
- **WHEN** the response contract section is inspected
- **THEN** it references `submit_review_findings` and
  `submit_lesson_proposal` by name

### Requirement: FR-011 Path validation

The tool MUST validate that `diff_path` resolves (after
symlink and `..` canonicalization) to a path within the
project root directory or the system temp directory
(`$TMPDIR`). Paths that resolve outside these scopes
MUST be rejected with a structured error result
containing error code `path_validation_failed`.

A `diff_path` that does not exist but whose location
(the nearest existing ancestor canonicalized, with the
missing suffix joined back on) lies inside one of these
scopes MUST be rejected with error code `diff_not_found`
instead. Its message MUST quote the supplied path and
tell the caller to copy the path exactly from the prompt
or manifest rather than retype it. A missing path whose
location lies outside the scopes MUST still fail with
`path_validation_failed`. Neither failure reads any file.

#### Scenario: Path traversal rejected

- **GIVEN** `diff_path` is `../../etc/shadow`
- **WHEN** `build_review_prompts` is called
- **THEN** the tool returns a failed result with error
  code `path_validation_failed`

#### Scenario: Missing path inside the allowed scope

- **GIVEN** `diff_path` names a file that does not exist
  in a directory inside the project root or the system
  temp directory (for example a mistyped temp file name)
- **WHEN** `build_review_prompts` is called
- **THEN** the tool returns a failed result with error
  code `diff_not_found`
- **AND** the message quotes the supplied path and says to
  copy the path exactly rather than retype it
- **AND** no file is read

#### Scenario: Missing path outside the allowed scope

- **GIVEN** `diff_path` names a non-existent file whose
  location resolves outside the project root and the
  system temp directory
- **WHEN** `build_review_prompts` is called
- **THEN** the tool returns a failed result with error
  code `path_validation_failed`

### Requirement: FR-012 Untrusted content delimiters

All caller-provided string content embedded in prompt
files (diff content from `diff_path`, `changed_files`,
`input_context`, `pre_flight_results`, `sibling_evidence`,
`existing_reviews`, `walkthrough`, `review_context`) MUST
be enclosed in clearly delimited untrusted content
markers. The tool MUST use a defensive preamble before
untrusted content sections instructing the model to treat
the enclosed content as data, not instructions.

The delimiter format MUST use a unique boundary string
that is unlikely to appear in legitimate content (e.g.,
`<!-- BEGIN UNTRUSTED: {section_name} -->` and
`<!-- END UNTRUSTED: {section_name} -->`).

The tool MUST strip or escape any occurrence of the
delimiter pattern within untrusted content before
enclosure. If the untrusted content contains the exact
delimiter string, the tool MUST replace it with a
neutralized form (e.g., inserting a zero-width space
or HTML-encoding the angle brackets) to prevent
delimiter breakout.

#### Scenario: Diff content delimited

- **GIVEN** a diff file containing arbitrary content
- **WHEN** `build_review_prompts` generates a prompt
- **THEN** the diff content is enclosed in untrusted
  content delimiters with a defensive preamble

### Requirement: FR-013 Input size limits

The tool MUST enforce a maximum total prompt file size
of 4 MiB (4,194,304 bytes) per agent. If assembling a
prompt for an agent would exceed this limit, the tool
MUST return a structured error result with error code
`prompt_size_exceeded` for that agent, including the
calculated size and the limit.

Individual string inputs SHOULD be validated against
reasonable size bounds. The `diff_path` file size MUST
NOT exceed 2 MiB. Other string inputs MUST NOT exceed
1 MiB each.

#### Scenario: Oversized diff rejected

- **GIVEN** `diff_path` points to a file larger than 2 MiB
- **WHEN** `build_review_prompts` is called
- **THEN** the tool returns a failed result with error
  code `prompt_size_exceeded`

## MODIFIED Requirements

### Requirement: Command file prompt construction

Previously: Each command file (`uf.review-council`,
`uf.review-pr`, `uf.triage-issue`, `uf.address-feedback`)
contained ~100-150 lines describing how to construct child
prompt files, including section ordering, file reads, and
content assembly.

Modified: Command files SHOULD reference
`build_review_prompts` for prompt assembly. The command
file retains responsibility for gathering review context
(diff, CI checks, pre-flight, sibling evidence) and
passing it to the tool. The 11-section specification in
command files becomes a reference to the tool's template
rather than inline prose.

Note: Command file updates are a separate change. This
spec defines the tool; command file migration is tracked
independently.

## REMOVED Requirements

None.
