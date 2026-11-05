<!-- capability: uf-rebase-pr
     change: retroactive-push-skill-contracts
     Archive ordering: this change MUST archive after
     port-uf-rebase-pr (whose text it restates).
     Supersession strategy: this MODIFIED delta restates the
     "Bounded force-push" requirement from port-uf-rebase-pr's
     text with the push strategy corrected to the shipped
     always-plugin-tool shape (PR 948, #947): the command
     calls git_push_force_with_lease with structured arguments
     and never invokes a raw shell force-push. The
     port-uf-rebase-pr delta carries a light inline annotation
     at the requirement prose. Design rationale: design.md D4
     and D5 in this change. -->

Requirement numbering: this delta allocates no new requirement
ids. The uf-rebase-pr capability's unarchived delta
(port-uf-rebase-pr/specs/uf-rebase-pr/spec.md) uses
requirement-name addressing with no FR ids; this delta keeps
that addressing unchanged. The "Bounded force-push"
requirement is restated below with the corrected push
strategy.

## MODIFIED Requirements

### Requirement: Bounded force-push

The rebase rewrites history, so the head branch MUST be
force-pushed with lease protection through the
`git_push_force_with_lease` plugin tool, called with
`{branch: <HEAD_BRANCH>, remote: <HEAD_REMOTE>, expected_sha:
<HEAD_SHA>}` — the tool pushes `HEAD:refs/heads/<HEAD_BRANCH>`
to `<HEAD_REMOTE>` with an explicit refspec, leasing on
`<HEAD_SHA>` (the head the PR reported before the rebase),
independent of any configured upstream. The command MUST NOT
invoke `git push` with `--force`, `--force-with-lease`, or any
other raw shell force-push form: every push goes through the
plugin tool, whose lease and branch guards provide the
bounding. The command MUST push only to `HEAD_REMOTE` (the
remote resolved for `HEAD_REPO`) and only the PR's head
branch. The command MUST NOT push to `BASE_REMOTE`, `main`,
`master`, or any protected branch. A `BRANCH_MISMATCH`
refusal means the checkout is not on a local branch named
`<HEAD_BRANCH>` — the command MUST NOT switch branches or
retry; an `EXPECTED_SHA_UNKNOWN` or lease rejection means the
remote head moved after the PR was fetched — the command MUST
NOT retry with a refreshed lease or `--force`.

#### Scenario: Plugin tool push after fallback checkout

- **GIVEN** the checkout fell back to a detached local ref
  (no local branch named `<HEAD_BRANCH>`) and a local branch
  `<HEAD_BRANCH>` was created at it
- **WHEN** the command force-pushes
- **THEN** it calls `git_push_force_with_lease` with
  `{branch: <HEAD_BRANCH>, remote: <HEAD_REMOTE>, expected_sha:
  <HEAD_SHA>}` — the tool's explicit refspec pushes the HEAD
  commit to `refs/heads/<HEAD_BRANCH>` on `<HEAD_REMOTE>`,
  independent of the branch's configured upstream, and no raw
  shell force-push runs

#### Scenario: Never pushes the base branch

- **GIVEN** the rebase completed on `<HEAD_BRANCH>`
- **WHEN** the command pushes
- **THEN** only `refs/heads/<HEAD_BRANCH>` on `<HEAD_REMOTE>`
  receives the push — never `<BASE_REMOTE>`, `main`, `master`,
  or any protected branch

Previously: the requirement (as added by port-uf-rebase-pr)
mandated the raw shell command `git push --force-with-lease
<HEAD_REMOTE> <HEAD_BRANCH>`, directing the command to use the
`git_push_force_with_lease` plugin tool only when the
checked-out branch's configured upstream verifiably matched
`<HEAD_REMOTE>/<HEAD_BRANCH>` and to fall back to the raw
refspec command otherwise. The shipped command (#947) always
routes the force-push through the plugin tool with structured
arguments — including after the fallback checkout — and raw
shell force-push forms are forbidden.

## REMOVED Requirements

None.
