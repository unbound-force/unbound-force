# Design: Refine Bash Permissions

## Context

OpenCode evaluates `permission.bash` rules using
last-match-wins semantics: the final rule whose glob
pattern matches the command string determines the
action (`allow` or `ask`).

The original PR #477/8900a34 permission block
contained 15 rules (14 ask rules + 1 global allow
default) gating: `gh issue create/edit/close/comment`,
`gh pr create/merge/close/comment/edit/review`,
`gh api`, `git push`, `git commit`, and `rm`.

The upstream `granular-bash-permissions` change
restored this block. This change refines the `gh api`
rules to distinguish read-only from mutating calls.

## Goals

- Allow read-only `gh api` calls without prompting
- Gate mutating `gh api` calls behind approval
- Preserve all non-api mutation guards

## Non-Goals

- Adding new mutation categories beyond `gh api`
- Modifying CI pipelines or Go source code

## Design Decisions

### D1: Option 2 — Allow-First for `gh api`

Default `gh api*` to `allow`, then explicitly `ask`
for mutating patterns. This correctly handles the
common case where agents call `gh api /repos/...`
without an explicit `-X GET` flag — under Option 1
(ask-first with GET override), such calls would
incorrectly prompt because `gh api /repos/...` does
not contain `-X GET`.

### D2: Rule Ordering (Last-Match-Wins)

1. `"*": "allow"` — global default
2. `gh issue/pr/git/rm` mutation guards — ask
3. `"gh api*": "allow"` — allow all gh api by default
4. `gh api` mutation patterns — ask (override step 3)

### D3: Wildcard Mutation Patterns

Eleven patterns gate mutating `gh api` calls:

**Method flags (short form)**:
- `"gh api * -X POST*": "ask"`
- `"gh api * -X PATCH*": "ask"`
- `"gh api * -X PUT*": "ask"`
- `"gh api * -X DELETE*": "ask"`

**Method flags (long form)** — actively used by this
project's own commands (uf.triage-issue, uf.review-pr,
uf.review-council, uf.address-feedback):
- `"gh api * --method POST*": "ask"`
- `"gh api * --method PATCH*": "ask"`
- `"gh api * --method PUT*": "ask"`
- `"gh api * --method DELETE*": "ask"`

**Data-sending flags**:
- `"gh api * -f *": "ask"`
- `"gh api * -F *": "ask"`
- `"gh api * --input*": "ask"`

### D4: Preserve Non-API Rules

All 14 non-api rules from the upstream permission
block (13 ask rules + 1 global allow default) are
restored unchanged:

- `gh issue create/edit/close/comment` (4 rules)
- `gh pr create/merge/close/comment/edit/review`
  (6 rules)
- `git push`, `git commit` (2 rules)
- `rm` (1 rule)
- `*: allow` global default (1 rule)

Total: 14 non-api + 1 `gh api*: allow` + 11 gh api
mutation patterns = **26 rules**.

## Risks

### R1: Pattern Gaps (Accepted)

Glob patterns cannot cover every possible flag
combination. The covered patterns match all documented
`gh` CLI mutation flags in both `-X` and `--method`
forms. The `<protect>` tags from PR #499 provide the
primary guard; permission rules are defense-in-depth.

### R2: Flag-Ordering Sensitivity (Accepted)

The pattern `"gh api * -X POST*"` expects content
between `gh api` and `-X` (the endpoint path). If a
command places the flag before the endpoint (e.g.,
`gh api -X POST /repos/...`), the `*` between `gh api`
and `-X` would need to match zero characters. Whether
this matches depends on OpenCode's glob implementation.

This is an accepted limitation:
- Agent usage and gh CLI convention typically place
  the endpoint before flags
- Data-sending flag patterns (`-f`, `-F`, `--input`)
  provide partial backstop coverage for mutations
  that include data
- The `<protect>` tags from PR #499 remain the primary
  guard for this edge case

The same ordering sensitivity applies to `--method`
flag patterns.

### R3: Rollback Procedure

Revert the commit to remove the refined `gh api`
rules. The upstream blanket `"gh api*": "ask"` rule
would be restored, or if the entire permission block
is removed, OpenCode defaults all commands to `allow`.
