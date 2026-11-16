<!-- capability: scaffold
     change: heredoc-ban-escaped-and-ansic
     originating_issue: 1048
     Archive ordering: SC-FR-016 is a new scaffold requirement
     introduced here. SC-FR-001..015 remain as allocated:
     SC-FR-001..006 review-council-multi-model-fanout and
     successors, SC-FR-007 model-matrix-delegation,
     SC-FR-008..010 model-matrix-scaffold-deployment,
     SC-FR-011..015 init-override-stubs-gitignore-sync.
     SC-FR-017..018 stay reserved. -->

Requirement numbering: SC-FR-016 is allocated by this change.
SC-FR-017..018 remain reserved for future scaffold changes.
SC-FR-001..015 are allocated across prior scaffold deltas as
noted in the header. This delta does not modify existing
requirement text; it adds one new requirement.

## ADDED Requirements

### Requirement: Heredoc-Ban Contract Coverage

**ID: SC-FR-016** This requirement MUST be satisfied.

The scaffold's instruction-template contract suite MUST reject
shell heredoc payloads in every instruction directory and its
scaffold mirror -- `.opencode/commands`,
`internal/scaffold/assets/opencode/commands`, `.opencode/agents`,
`internal/scaffold/assets/opencode/agents`, `.opencode/skills`,
`internal/scaffold/assets/opencode/skills` -- using a single
detection pattern that covers all bash quoted-heredoc spellings
that carry the #1011 delimiter-breakout vector.

The pattern MUST match each of the following positive spellings.
The list is exhaustive and measurable: every spelling below MUST
appear as a positive case in the table-driven subtest, and no
other spelling is required:

- Unquoted: `<<EOF`
- Single-quoted: `<<'EOF'`
- Double-quoted: `<<"EOF"`
- Dashed (strip-tab): `<<-EOF`, `<<-'EOF'`, `<<-"EOF"`
- Escaped-delimiter: `<<\EOF`, `<<\ 'EOF'`, `<<\'EOF'`, and the
  dashed `<<-\EOF`, `<<-\ 'EOF'` (backslash escaping the
  delimiter word, disabling expansion identically to a quoted
  heredoc)
- ANSI-C-quoted: `<<$'EOF'` and `<<-$'EOF'` (`$'...'` prefix
  after `<<`)
- Here-string: `<<<` with a word payload (e.g., `<<< "foo"`,
  `<<< 'bar'`), detected via its `<<` prefix

The contract test `TestCommandContracts_RejectShellHeredocPayloads`
in `internal/scaffold/command_contract_test.go` MUST enforce this
via a compiled regular expression whose canonical spelling is

```
<<-?\s*\$?\\?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?
```

and whose behavior MUST be pinned by a table-driven subtest
(adjoining `TestCommandContracts_HeredocPattern` or equivalent)
whose table is exhaustive and quantified: it MUST assert all 15
positive cases enumerated above and all 4 negative cases
enumerated below (19 total), so silent case loss fails
visibly. The subtest asserts:

#### Scenario: Every recognized heredoc spelling is banned

- **GIVEN** any string equal to one of the positive spellings
  listed above (including the escaped and ANSI-C forms and
  at least one here-string example)
- **WHEN** the contract's heredoc pattern is applied
- **THEN** the pattern MUST produce a match

#### Scenario: Non-heredoc prose is not banned

- **GIVEN** a string equal to `>>> MANDATORY GATE <<<` or
  `>>> MANDATORY GATE: HUMAN CONFIRMATION REQUIRED <<<`
  (human-confirmation marker prose containing `<<<` without an
  identifier), or equal to `<<<<<<<` or `>>>>` (bare
  conflict-marker runs without an identifier)
- **WHEN** the contract's heredoc pattern is applied
- **THEN** the pattern MUST produce no match

#### Scenario: Scan enforces the pattern over instruction directories

- **GIVEN** any markdown file under the six instruction
  directories listed above
- **WHEN** `TestCommandContracts_RejectShellHeredocPayloads` walks
  each file line-by-line
- **THEN** any line that contains a substring matching the pattern
  MUST cause the test to fail with an actionable message naming
  the file, line, and matched text, directing the author to write
  the payload with the host's file-write tool instead

#### Scenario: Walk failures fail closed

- **GIVEN** a walk error during the scan of any instruction
  directory -- a missing directory, an unreadable file, a
  symlink loop, or a permission failure
- **WHEN** `TestCommandContracts_RejectShellHeredocPayloads`
  walks the directory
- **THEN** the scan MUST fail the test rather than silently
  reducing coverage

#### Scenario: Synthetic fixture pins the walk-to-message path

- **GIVEN** a `t.TempDir()` fixture directory containing a
  markdown file with a line that matches the pattern (e.g.,
  `<<\EOF` planted on a known line)
- **WHEN** the contract scan walks the fixture directory
- **THEN** the scan MUST report exactly that violation with an
  actionable message naming the file, the line number, and the
  matched text, directing the author to write the payload with
  the host's file-write tool instead

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
