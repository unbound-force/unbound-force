<!-- capability: artifact-provenance
     change: ref-lookup
     Provenance records pin the commit a ref resolved to.
     Design rationale: design.md (D5). -->

Requirement numbering: AP-FR-001 through AP-FR-003 are allocated
for this change. AP-FR-004 through AP-FR-020 are reserved for
future artifact-provenance changes. This is the capability's first
delta; no other range is allocated. Ref resolution, which Stamp
calls, is numbered RR-FR-NNN in specs/ref-resolution/spec.md.

## ADDED Requirements

### Requirement: Commit-Pinned Provenance

AP-FR-001: `provenance.Stamp(dir, ref)` MUST return a Record whose
`Ref` is the normalized ref and whose `Commit` is the commit
object name `refs.Lookup(dir, ref)` resolves.

AP-FR-002: Stamp MUST return an error, and a zero Record, when the
ref cannot be resolved. It MUST NOT return a Record with an empty
`Commit`.

AP-FR-003: Record MUST serialize to JSON as
`{"ref": <string>, "commit": <string>}`.

#### Scenario: Stamp pins the commit
- **GIVEN** a repository whose `main` points at commit C
- **WHEN** Stamp is called with ` refs/heads/main\n`
- **THEN** the Record is `{Ref: "main", Commit: C}`

#### Scenario: Unresolvable ref
- **GIVEN** no ref named `missing`
- **WHEN** Stamp is called with it
- **THEN** it returns a zero Record and an error naming the ref

#### Scenario: JSON shape
- **GIVEN** `Record{Ref: "main", Commit: "abc"}`
- **WHEN** it is marshaled to JSON
- **THEN** the output is `{"ref":"main","commit":"abc"}`
