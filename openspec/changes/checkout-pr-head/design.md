## Context

PR #955 added Section 1.2.1 to `/uf.address-feedback` so the working
tree is on the PR head before any branch-keyed progress read. The
section is fixed procedure with the same outcome for the same repo
state. Issue #965 asks for the `uf_unleash_branch_gate` treatment:
an injectable-exec plugin tool that owns the sequence.

## Goals / Non-Goals

**Goals:**

- One tested implementation of every Section 1.2.1 decision.
- No PR-derived refname reaches a shell inside the tool.
- Command prose shrinks to a call plus per-code handling.

**Non-Goals:**

- Moving the Section 4.4 fetch, divergence check, or push into the
  tool (separate judgment gate: rebase vs abort).
- Editing Sections 4.5/4.6 (concurrent work, #966).

## Decisions

### D1: Tool lives in uf-workflow index.ts

The factory `createCheckoutPrHeadTool(exec, gh, cwd)` sits beside
`createUnleashResolveIssueTool` and reuses `GitExec`, `GhRunner`,
`defaultGitExec`, and `defaultGhRunner`. No new `lib/` module: that
would add a scaffold asset and manifest entries for ~150 lines with
one consumer.

### D2: Decision sequence

1. `git rev-parse HEAD`; equal to `head_sha` → identity holds;
   else `git merge-base --is-ancestor <head_sha> HEAD` exit 0 →
   descendant (`ahead_of_head_sha: true`); any other exit →
   unrelated.
2. Unrelated: `git status --porcelain --untracked-files=no`
   non-empty → `DIRTY_TREE`; else `gh pr checkout <N>`. A throw is
   `CHECKOUT_FAILED` with stderr; the untracked-collision message
   gets the `git stash -u` remediation and `retryable: false`;
   anything else (fetch, network, auth) is `retryable: true`.
   Re-run step 1; still unrelated → `UNRELATED_BRANCH` with both
   SHAs.
3. `git branch --show-current`; empty (detached) →
   `NO_PUSH_TARGET`. Branch outside `A-Za-z0-9._/-` →
   `UNSAFE_PUSH_TARGET` before any config read.
4. `git config --get branch.<b>.remote` / `.merge` (unset reads as
   empty) → `NO_PUSH_TARGET` if either is empty.
5. Merge ref under `refs/pull/` → `READ_ONLY_MERGE_REF`.
6. Merge ref other than `refs/heads/<head_ref>` →
   `UNRELATED_BRANCH` (ancestry cannot reject a `main` that absorbed
   the merged PR).
7. Merge ref outside the refname charset, or remote neither a plain
   name (`A-Za-z0-9._-`, excluding `.`/`..`) nor an `https://` URL
   without userinfo → `UNSAFE_PUSH_TARGET`.

Any unexpected throw (for example not a git repository) returns
`CHECKOUT_PR_HEAD_ERROR`.

### D3: Keep the charset gate as a sixth code

Argument arrays make the tool's own git calls safe, but Section 4.4
still renders `HEAD_REMOTE` and `HEAD_MERGE_REF` into shell
templates. The gate therefore stays, as `UNSAFE_PUSH_TARGET`. `.`
and `..` are rejected as remotes (they resolve as local repository
paths), matching `RemoteNameSchema`.

### D4: `checked_out` in the success data

Reports whether `gh pr checkout` ran. It is additive to the issue's
four fields and makes the round trip observable in tests and logs.

### D5: Contract tests move to the tool contract

Go clauses pin the call, the input shape, the output-field
recording, and each code's handling line; a forbidden list keeps the
removed git templates out. A TS test extracts the `- \`CODE\`:`
bullets from Section 1.2.1 of both copies and requires equality with
the exported `CHECKOUT_PR_HEAD_ERROR_CODES`, so adding a code without
command handling (or the reverse) fails. The Section-1.2.1-before-
first-CHECKPOINT ordering pin is unchanged.

## Risks / Trade-offs

- A `gh pr checkout` that succeeds but leaves a branch whose config
  differs from what gh writes is caught by steps 4–6, as in prose.
- `git config --get` failures other than "unset" also read as unset,
  yielding `NO_PUSH_TARGET`: fail-closed.
