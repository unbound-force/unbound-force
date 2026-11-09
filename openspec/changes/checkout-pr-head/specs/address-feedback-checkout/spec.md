<!-- capability: address-feedback-checkout
     change: checkout-pr-head
     requirement ids: AFC-FR-001 to AFC-FR-004 allocated here -->

## ADDED Requirements

### Requirement: checkout_pr_head Tool

**ID: AFC-FR-001** This requirement MUST be satisfied.

The `uf-workflow` plugin MUST register a `checkout_pr_head` tool
taking `pr_number` (positive integer), `head_ref` (non-empty
string), and `head_sha` (40-character lowercase hex). It MUST
invoke git and gh through argument arrays only. On success it MUST
return `branch`, `head_remote`, `merge_ref`, `ahead_of_head_sha`,
and `checked_out`. It MUST NOT reset or discard local commits.

#### Scenario: Already at the PR head

- **GIVEN** `HEAD` equals `head_sha` on a branch whose merge ref is
  `refs/heads/<head_ref>`
- **WHEN** `checkout_pr_head` is called
- **THEN** it MUST return ok with `checked_out: false` and
  `ahead_of_head_sha: false`
- **AND** it MUST NOT run `gh pr checkout`

#### Scenario: Unpushed commits from a prior run

- **GIVEN** `HEAD` is a strict descendant of `head_sha`
- **WHEN** `checkout_pr_head` is called
- **THEN** it MUST return ok with `ahead_of_head_sha: true`
- **AND** it MUST NOT run `gh pr checkout`

#### Scenario: Fork default branch

- **GIVEN** `gh pr checkout` checks out `contrib/main` for a PR
  whose `head_ref` is `main`
- **WHEN** the push target is read
- **THEN** `branch` MUST be `contrib/main` and the config MUST be
  read for `branch.contrib/main.*`

### Requirement: Checkout Failure Envelopes

**ID: AFC-FR-002** This requirement MUST be satisfied.

When `HEAD` is neither `head_sha` nor its descendant, the tool MUST
return `DIRTY_TREE` if tracked files are modified, and otherwise run
`gh pr checkout <pr_number>`. A failed checkout MUST return
`CHECKOUT_FAILED` carrying the specific git or gh error; an
untracked-file collision MUST name the `git stash -u` remediation
and be non-retryable, and any other failure (including fetch and
network errors) MUST be retryable. If `HEAD` is still unrelated
after the checkout, the tool MUST return `UNRELATED_BRANCH`
reporting both SHAs.

#### Scenario: Untracked collision

- **GIVEN** `gh pr checkout` fails with "untracked working tree
  files would be overwritten"
- **WHEN** `checkout_pr_head` is called
- **THEN** it MUST return `CHECKOUT_FAILED` with `retryable: false`
  and a message naming `git stash -u`

### Requirement: Push Target Resolution

**ID: AFC-FR-003** This requirement MUST be satisfied.

The tool MUST read `branch.<branch>.remote` and
`branch.<branch>.merge` for the actual checked-out branch. It MUST
return `NO_PUSH_TARGET` for a detached `HEAD` or an unset value,
`READ_ONLY_MERGE_REF` for a merge ref under `refs/pull/`, and
`UNRELATED_BRANCH` for a merge ref other than
`refs/heads/<head_ref>`. It MUST return `UNSAFE_PUSH_TARGET` when
the branch or merge ref falls outside `A-Za-z0-9._/-`, or the
remote is neither a plain name (`A-Za-z0-9._-`, not `.` or `..`)
nor an `https://` URL without userinfo.

#### Scenario: Main absorbed the merged PR

- **GIVEN** the tree is on `main`, which descends from `head_sha`,
  with merge ref `refs/heads/main`, and `head_ref` is `feature/x`
- **WHEN** `checkout_pr_head` is called
- **THEN** it MUST return `UNRELATED_BRANCH`

### Requirement: Command Consumes the Tool

**ID: AFC-FR-004** This requirement MUST be satisfied.

Section 1.2.1 of `/uf.address-feedback` (canonical and scaffold
mirror, byte-identical) MUST call `checkout_pr_head` with the
frozen context, record `head_remote` and `merge_ref` as
`HEAD_REMOTE` and `HEAD_MERGE_REF`, and STOP on every error code
with a handling line per code. It MUST NOT carry the removed git
choreography as prose.

#### Scenario: Code list parity

- **GIVEN** the exported failure-code list of `checkout_pr_head`
- **WHEN** the Section 1.2.1 handling list is compared with it
- **THEN** the two MUST match exactly in both command copies
