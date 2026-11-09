## Why

Section 1.2.1 of `/uf.address-feedback` ("Check Out the PR Head",
added by PR #955) is 108 lines of judgment-free git choreography as
prose: compare `HEAD` with the frozen head SHA, gate on a dirty tree,
run `gh pr checkout`, re-verify, read the push remote and merge ref,
reject read-only `refs/pull/` merge refs, cross-check the merge ref
against `refs/heads/<head_ref>`, and gate PR-derived refnames on a
charset before they reach shell templates. The prose lives in two
byte-identical copies (canonical command and scaffold mirror), so
each of PR #955's seven edge-case fixes was a two-file markdown edit,
and every model run re-interprets the branch conditions.

AGENTS.md's Determinism gate and determinism pack DE-001/DE-005
apply: this is newly added mechanical instruction, not legacy prose
under the opportunistic-migration carve-out. Issue #965 is the named
extraction follow-up; the charset-gate prose itself says "#965
tracks the argument-array tool".

`originating_issue: 965` is set in `.openspec.yaml`.

## What Changes

1. **New `checkout_pr_head` tool** in the `uf-workflow` plugin,
   beside `uf_unleash_branch_gate`, with injectable git and gh
   executors (argument arrays, no shell). Input: `pr_number`,
   `head_ref`, `head_sha`. Success data: `branch`, `head_remote`,
   `merge_ref`, `ahead_of_head_sha`, `checked_out`.
2. **Failure envelopes** for every former STOP condition:
   `DIRTY_TREE`, `CHECKOUT_FAILED` (specific git/gh error,
   `git stash -u` remediation for the untracked-file collision,
   `retryable` for fetch/network failures), `UNRELATED_BRANCH`
   (post-checkout HEAD unrelated, or merge ref is not
   `refs/heads/<head_ref>`), `READ_ONLY_MERGE_REF`,
   `NO_PUSH_TARGET` (detached HEAD or unset remote/merge), and
   `UNSAFE_PUSH_TARGET` (the charset gate; a sixth code beyond the
   issue's five, because the gate is a distinct STOP condition and
   folding it into another code would lose the remediation).
3. **Section 1.2.1 collapsed** to the tool call, one output example
   (DE-002), and a per-code handling list, in both command copies.
   Section 4.4 is unchanged: it still reads `HEAD_REMOTE` and
   `HEAD_MERGE_REF` "recorded in Section 1.2.1", now from the tool's
   `head_remote` and `merge_ref`.
4. **Plugin mirror** synced byte-identical to
   `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`.
5. **Contract tests** rewritten against the tool contract: the Go
   command-contract clauses pin the tool call and every code, a
   forbidden-fragment list keeps the old choreography from
   returning, and a TS test pins the command's code list to the
   exported `CHECKOUT_PR_HEAD_ERROR_CODES` in both copies.

## Capabilities

### New Capabilities
- `address-feedback-checkout`: deterministic PR-head checkout and
  push-target resolution for `/uf.address-feedback`.

### Modified Capabilities
- None beyond the command-prose collapse; observable STOP behavior is
  preserved.

### Removed Capabilities
- None.

## Impact

- **Files created**: `.opencode/test/checkout-pr-head.test.ts`
- **Files modified**: `.opencode/plugins/uf-workflow/index.ts` and its
  scaffold mirror; `.opencode/commands/uf.address-feedback.md` and its
  scaffold mirror (Section 1.2.1 only);
  `.opencode/test/{uf-workflow,scratch-smoke,golden-output-shape}.test.ts`;
  `internal/scaffold/command_contract_test.go`; `CHANGELOG.md`.
- **No scaffold manifest change**: both assets already ship.
- **Instruction delta (DE-004)**: Section 1.2.1 108 → 50 lines
  (733 → 336 words) per copy.
- **Documentation**: user-facing behavior is unchanged from PR #955;
  the existing docs issue #967 covers it (no new issue).

## Constitution Alignment

- **I Autonomous Collaboration** — PASS: structured JSON envelope,
  no hero coupling.
- **II Composability** — PASS: independently invocable tool, no new
  dependency.
- **III Observable Quality** — PASS: every failure carries a code and
  the specific git/gh error.
- **IV Testability** — PASS: injected executors; unit, golden, and
  registration tests; coverage floors unchanged.
- **V Security** — PASS: argument arrays remove shell interpolation;
  the charset gate still protects the Section 4.4 templates.
- **VI Deterministic Execution** — PASS: mechanical choreography moves
  from prose into a tested tool.
