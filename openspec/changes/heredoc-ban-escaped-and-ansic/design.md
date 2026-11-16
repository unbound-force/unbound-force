<!-- capability: scaffold
     change: heredoc-ban-escaped-and-ansic
     originating_issue: 1048 -->

## Context

`TestCommandContracts_RejectShellHeredocPayloads`
(`internal/scaffold/command_contract_test.go:845-883`) bans shell
heredocs in instruction templates because a heredoc ends at any line
equal to its delimiter, quoted or not (#1011). Composed text
(comment bodies, child issue payloads) written through a heredoc
could break out and execute as shell. The ban was introduced in
commit 79107ad1 (Closes #1011) for canonical and scaffolded command
templates and extended in 505dedf8 (Fixes #1047) to agents and skills
-- the current scan covers six directories:
`.opencode/commands`, `internal/scaffold/assets/opencode/commands`,
`.opencode/agents`, `internal/scaffold/assets/opencode/agents`,
`.opencode/skills`, `internal/scaffold/assets/opencode/skills`.

Current regex (L858):

```go
heredoc := regexp.MustCompile(`<<-?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?`)
```

It matches `<<EOF`, `<<'EOF'`, `<<"EOF"`, `<<-EOF`, and here-string
`<<<` via its trailing `<<`, but misses two spellings bash treats as
quoted (no expansion, still delimiter-terminated):

- Escaped delimiter: `<<\EOF` and `<<\ 'EOF'` -- the `\` escapes the
  delimiter word, disabling expansion exactly like a quoted delimiter.
- ANSI-C-quoted: `<<$'EOF'` (and dashed `<<-$'EOF'`) -- the `$'...'`
  form is the shell's ANSI-C quoting; after `<<` bash parses the
  `$'` prefix as part of the quoted delimiter.

The uf-divisor review on PR #1024 flagged the first gap (MEDIUM)
and suggested `<<-?\s*\\?['"]?[A-Za-z_][A-Za-z0-9_]*['"]?`; the issue
also asks to cover `$'EOF'` and to add a table pinning both
recognized spellings and non-matches (`>>> MANDATORY GATE <<<`,
`<<<<<<<` conflict markers).

## Goals / Non-Goals

### Goals

- Match every bash quoted-heredoc spelling that carries the #1011
  breakout vector: unquoted, single-quoted, double-quoted, dashed,
  escaped-delimiter, ANSI-C-quoted, and here-string.
- Pin the mapping in a table-driven subtest that asserts both
  positives (must match) and negatives (must not match), so future
  spillings and false positives are caught at test time.
- Preserve the existing non-matches: `>>> MANDATORY GATE <<<` and
  `<<<<<<<` prose must remain allowed.

### Non-Goals

- Rewriting or changing any instruction template, scaffold asset,
  or `uf` command behavior -- only the guarding test changes.
- Banning here-docs in product Go code or in documentation prose
  outside the six instruction directories.
- Introducing a new capability, schema, or CI workflow.

## Decisions

### D1: Regex extension is `\$?\\?` not `\\?` alone

The review's `\\?` covers the escaped form but not `$'EOF'`. The
ANSI-C form is `$'` immediately after optional whitespace -- bash
parses `<<$'EOF'` as `<<` + `$'` prefix + `EOF`. Placing an
optional `\$?` before the optional `\\?` -- itself followed by
another optional `\s*` -- gives the canonical SC-FR-016 spelling
`<<-?\s*\$?\\?\s*['"]?[A-Za-z_][A-Za-z0-9_]*['"]?`, which
matches in order: optional `-` (strip-tab), optional whitespace,
optional `$`, optional `\`, optional whitespace, optional quote,
identifier `[A-Za-z_][A-Za-z0-9_]*`, optional closing quote.
Ordering `$` before `\` reflects bash's
lexing (`$` then `\` then quote) and avoids needing alternations.

Alternative considered: `(\$|\\|['"])?` alternation. Rejected --
over-general, harder to reason about the required ordering, and
not needed for the finite set of spellings; the sequential optional
prefixes are minimal and match the shell's actual token order.

Trade-off: the resulting pattern is slightly more permissive than
strict bash (e.g., `<<\ "EOF"` would match, though bash would not
write it that way). Permissiveness is acceptable for a ban -- a
false positive on a non-shell string in an instruction template
would be immediately visible and the template can be reworded; a
false negative is the security gap we are closing.

### D2: Table-driven subtest lives alongside the scan loop

The scan loop remains the enforcement mechanism (fails the test on
any heredoc in the six directories). The table is a peer subtest
(`TestCommandContracts_HeredocPattern`) that asserts the regex
against a static list of strings -- positives must produce a match,
negatives must produce none. Co-location keeps the pattern and its
spec in one file, so a regex edit must update the table or the
table fails.

Alternative considered: split into a separate helper package.
Rejected -- the regex is test-local by construction (not exported,
not shared with runtime), and extracting it would create an
indirection with no consumer.

### D3: Non-matches explicitly list `>>> MANDATORY GATE <<<` and `<<<<<<<`

The `>>> MANDATORY GATE <<<` human-confirmation markers in
`.opencode/commands/uf.address-feedback.md` and `uf.fix-ci.md`
contain `<<<` but no identifier (`<<<` followed by end-of-line
or non-identifier) and therefore must not match; the table asserts
this. Similarly `<<<<<<<` conflict-marker prose (bare `<` run
without an identifier) must not match. Both are cited in the
originating issue as required non-matches.

### D4: Here-string covered via the `<<` prefix of `<<<`

Bash here-string `<<< "foo"` tokenizes as `<<` + `<` + word; the
contract regex matches the first `<<` + `<`-as-whitespace? No --
it matches `<<` then `\s*` then `['"]?` + identifier. For `<<< "foo"`
the string is `<<< "foo"`: positions `<<` + `< "foo"` -- after
`<<` the next char is `<`, then space, then `"foo"`. The pattern
`<<-?\s*` would require whitespace after `<<`, so `<<< "foo"` would
not match if read literally. In practice the prior regex already
matched `<<< "foo"` via `<<` + `<`? Re-examined: `<<< "foo"` is
`<<` + `<` -- but our whitespace class does not include `<`. The
existing test's prior pass for this case relied on the `<` being
consumed as part of `\s*`? Actually `\s` does not include `<`. So
the original expectation that `<<<` is caught "via the second `<`"
means the pattern should be read as `<<` matching the last two
`<` of `<<<`, and the identifier comes after. The table pins the
exact expectation and will drive the final pattern wording if an
adjustment (e.g., allowing `<` in the whitespace gap) is needed;
the spec delta records that the here-string spelling MUST be
matched, without mandating the exact whitespace semantics beyond
the existing `<<` prefix.

### D5: Requirement ID SC-FR-016

SC-FR-001..006 are allocated by `review-council-multi-model-fanout`
and successors; SC-FR-007 by `model-matrix-delegation`;
SC-FR-008..010 by `model-matrix-scaffold-deployment`; SC-FR-011..015
by `init-override-stubs-gitignore-sync`. SC-FR-016 was reserved
for future scaffold changes. This change claims it. SC-FR-017..018
remain reserved.

## Risks / Trade-offs

- **False positives on prose containing `$` or `\`**: instruction
  prose rarely contains a literal `<<$'X'` or `<<\X` outside a shell
  block. If it does, the table's negative list can be extended or
  the template reworded; the ban is intentionally conservative.
- **Regex drift from bash semantics**: the pattern is a
  purposefully over-approximate ban, not a bash parser. Over-
  matching is safe (fail closed); under-matching is the gap being
  fixed. The table mitigates drift by pinning both directions.
- **No template change**: if a template already contains an
  escaped/ANSI-C heredoc, CI will newly fail after merge. No such
  payload exists today (verified by scanning the six directories
  pre-merge); if it did, the fix is to migrate that payload to
  the host's file-write tool per #1011.
