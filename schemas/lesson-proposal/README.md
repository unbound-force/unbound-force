# Lesson Proposal Schema

Version 1.0.0 defines the closed JSON object that a child review may emit
between `<!-- uf-lesson-proposal:v1 -->` and
`<!-- /uf-lesson-proposal -->`. The runtime boundary permits one section,
checks UTF-8 byte limits, and grounds every source against accepted sibling
evidence before preparing an existing-Dewey learning payload.

## Parent Storage Contract

The policy normalizes only CRLF line endings while grounding excerpts. It
normalizes lesson information with Unicode NFC and the specified whitespace
rules for storage and deduplication. Fixed secret and tool detectors reject
unsafe information without recording matched values.

`prepare_lesson_learning` accepts child output, accepted sibling evidence, and
at most 1024 unique lowercase SHA256 dedupe identities found by the parent in
existing learnings. A ready result contains the exact `store_learning` fields:
normalized `information`, generated `sibling-<name>` `tag`, and fixed
`reference` category. It also returns complete source provenance and the
parent-computed dedupe SHA256 for audit.

The information sent to Dewey ends with one stable
`UF_LESSON_PROVENANCE_V1` JSON metadata line containing the dedupe identity,
sibling, commit, and sorted source path/hash pairs. The plugin never calls
Dewey. Parent commands MUST query existing learnings for dedupe identities,
call `prepare_lesson_learning`, and call `dewey_store_learning` exactly once
only for a ready result. A known dedupe identity returns `duplicate`. Dewey
unavailability is an informational parent-recorded skip and does not change
the review verdict.

## Fixtures

`samples/sample-lesson-proposal.json` is valid. Files beginning with
`invalid-` exercise closed-object and structural bounds. JSON Schema length
keywords count characters; the executable policy separately enforces every
specified UTF-8 byte limit.
