<!-- capability: scaffold
     change: heredoc-ban-escaped-and-ansic
     originating_issue: 1048 -->

## Why

From the uf-divisor review on PR #1024 (MEDIUM finding,
discussion_r4531561260): the heredoc detection regex in
`TestCommandContracts_RejectShellHeredocPayloads`
(`internal/scaffold/command_contract_test.go:858`) misses the
escaped-delimiter heredoc. Bash treats `<<\EOF` (and `<<\ 'EOF'`)
identically to a quoted heredoc -- no expansion, body still
terminating at any line equal to the delimiter -- so it carries
the exact #1011 breakout vector, but the pattern
`<<-?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?` requires a quote or
identifier immediately after `<<` and a backslash matches neither
class.

The ANSI-C-quoted form `<<$'EOF'` is a second miss: the `$`
prefix is valid shell for the same quoted-delimiter semantics and
must also be banned. Left unfixed, a command, agent, or skill
template could reintroduce a heredoc breakout through either
spelling and pass the contract test.

The existing test is a single regex match with no table-driven
pinning of positives and negatives, so future regressions in either
direction (missing a spelling or false-positive on non-heredoc
prose) are likely.

Fixes #1048 (lineage: #1011 breakout, #1047 scope to agents/skills).

## What Changes

Harden the heredoc-ban contract test:

- Extend the regex to cover the escaped-delimiter `<<\EOF`
  (`<<\` with optional whitespace handling) and the ANSI-C-quoted
  `<<$'EOF'` form. The suggested fix in the finding --
  ```go
  heredoc := regexp.MustCompile(`<<-?\s*\\?['"]?[A-Za-z_][A-Za-z0-9_]*['"]?`)
  ```
  -- is the base; this change also prefixes an optional `\$?`
  so that `<<$'EOF'`, `<<$'X'`, and their dashed `<<-$'EOF'`
  variants are matched, and inserts an optional `\s*` between
  the optional backslash and the optional quote so the spaced
  escaped-delimiter `<<\ 'EOF'` is matched, while `>>> MANDATORY
  GATE <<<` and `<<<<<<<` conflict-marker prose remain
  non-matches.
  Here-string `<<<` is covered via its second `<` aligning to
  the `<<` prefix (exercised in the table). The exact regex
  spelling is recorded in the spec delta (SC-FR-016) and
  validated by the table.

- Add a table-driven subtest that pins (a) every recognized
  spelling -- unquoted `<<EOF`, single-quoted `<<'EOF'`,
  double-quoted `<<"EOF"`, dashed `<<-EOF`, escaped `<<\EOF`
  and `<<\ 'EOF'`, ANSI-C `<<$'EOF'`, and here-string `<<<` --
  and (b) every non-match -- `>>> MANDATORY GATE <<<` markers
  (three `<` without an identifier) and `<<<<<<<` conflict
  markers. The table lives in `command_contract_test.go` next to
  the scan loop and is the source of truth for future spellings.

No command, agent, or skill template is changed; only the
contract test that guards them.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `scaffold`: heredoc-ban contract test (SC-FR-016) -- regex
  extended to escaped-delimiter and ANSI-C-quoted forms; adds
  table-driven positive/negative pinning. Existing scan over
  `.opencode/commands`, `internal/scaffold/assets/opencode/commands`,
  `.opencode/agents`, `internal/scaffold/assets/opencode/agents`,
  `.opencode/skills`, `internal/scaffold/assets/opencode/skills`
  is unchanged.

### Removed Capabilities
- None.

## Impact

- **Systems/files**: `internal/scaffold/command_contract_test.go`
  only. No template, scaffold asset, CLI, workflow, or schema change.
- **Behavior**: strictly tighter ban -- previously-passing
  escaped/ANSI-C heredocs now correctly fail the test. No
  relaxation; no new runtime dependency.
- **Compatibility**: change is a test hardening; it cannot break
  production code, only surface latent template violations earlier.
  The `>>> MANDATORY GATE <<<` and `<<<<<<<` non-matches ensure
  the existing human-confirmation markers and merge-conflict prose
  remain unaffected (see Risks).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0, six principles).

### I. Autonomous Collaboration

**Assessment**: PASS

Does not introduce runtime coupling or synchronous hero
interaction. Artifact envelope and self-describing outputs are
untouched; the change is confined to a local contract test.

### II. Composability First

**Assessment**: PASS

No new hard prerequisite or mandatory dependency. The scaffold
and command assets remain independently installable; the tighter
ban preserves standalone value and adds no integration requirement
or patching of another hero.

### III. Observable Quality

**Assessment**: PASS

Quality claim (heredoc ban coverage) is backed by an automated
reproducible test with table-driven evidence for positives and
negatives. The change improves machine-parseable evidence; the
test's JSON/plain failure output and provenance remain stable.

### IV. Testability

**Assessment**: PASS

Directly improves testability: verifies observable side effects
(the contract scan) rather than implementation details, pins
coverage strategy for current and future heredoc spellings, and
is testable in isolation without network or shared mutable state.
No missing coverage strategy remains.

### V. Security by Default

**Assessment**: PASS

Treats security as structural: heredoc breakout is a shell
injection vector via delimiter-equal lines in untrusted
title/body/comment payloads (#1011). Closing the escaped and
ANSI-C gaps enforces input-validation hygiene before
security-sensitive shell execution. No new dependency or
privilege expansion.

### VI. Deterministic Execution

**Assessment**: N/A

No deterministic procedure is extracted to or from a tool, and no
prescribed output template is introduced. The change hardens an
existing test's pattern without adding or removing tool-executed
mechanical procedures.
