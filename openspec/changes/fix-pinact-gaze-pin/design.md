## Context

`.github/workflows/ci_local.yml:101` pins
`unbound-force/gaze/.github/workflows/reusable_gaze_ai_report.yml`
at the untagged merge commit `8bdfb59` (Gaze PR #280, merged
2026-10-27, sitting between v2.2.0 and v2.3.0) with no version
comment. It is the only `uses:` reference in the file lacking
one. pinact fails it with code 005, turning the
`Standardized CI / Verify action versions` check red on PR
#864; `release.yml` gates release preflight on that check.

A comment-only repair is impossible: pinact code 001 requires
the comment's named tag to resolve to exactly the pinned
commit, and no Gaze tag points at `8bdfb59`. The pin must move
to a tagged commit. Full constitution alignment is assessed in
`proposal.md` (PASS on Principles I-V, N/A on VI); this design
operationalizes Principle V (Security by Default) and
Principle III (Observable Quality).

## Goals / Non-Goals

### Goals

- Make the Gaze pin compliant with the
  `Standardized CI / Verify action versions` (pinact
  `--verify`) contract so the check passes once it runs
  (it executes only where org-infra reusable CI v0.9.0 is
  pinned, currently the dependabot bump PR #864).
- Keep the pin a full 40-character immutable commit SHA --
  never a mutable tag ref (Constitution Principle V).
- Guarantee zero behavior change: the referenced
  `reusable_gaze_ai_report.yml` is byte-identical between the
  old and new SHAs, so inputs, secrets, and permissions are
  untouched.
- Verify the tag-to-SHA mapping and byte-identity at
  authoring time with reproducible `gh api` commands (CI-002:
  never source SHAs from issue text alone).
- Record the previous SHA in the fix PR as the documented
  rollback path (Gatekeeping condition).

### Non-Goals

- No changes to workflow logic, job conditions, inputs,
  secrets, or permissions in `ci_local.yml`.
- No Gaze-side changes; this repo consumes a published Gaze
  release tag, not a fork or patch.
- No pinact configuration: no `.pinact.yaml`, no ignore
  rules, no `min_age` -- suppressing the finding would weaken
  the supply-chain gate.
- No dependabot configuration changes.
- No backfilling version comments on other `uses:`
  references; every other pin in the file already carries one.

## Decisions

1. **Move the pin to the v2.3.0 tag commit, not annotate the
   untagged SHA.** pinact codes 001 and 005 together leave no
   other compliant state: the comment must exist (005) and
   must name a tag resolving to exactly the pinned SHA (001).
   Alternatives rejected: keeping `8bdfb59` with a `# v2.3.0`
   comment fails `--verify` (SHA != tag commit); ignoring the
   action via `.pinact.yaml` weakens the supply-chain gate.

2. **SHA-to-SHA repin with an accurate version comment**
   (`8031d41943b06407cb92322791b177b41b7b2775  # v2.3.0`),
   satisfying CI-001 (immutable SHA), CI-002 (authoring-time
   upstream verification), and CI-003 (accurate tag comment),
   and restoring the file's uniform pin format.

3. **Authoring-time verification before the edit.** Resolve
   the v2.3.0 tag commit via the GitHub API, confirm it equals
   the target SHA, and compare the workflow file's content
   (line count and SHA-256) between `8bdfb59` and the target
   before writing the pin. If any check fails, stop and report
   per the Gatekeeping rule rather than proceeding.

4. **Rollback via the PR record.** The previous SHA
   `8bdfb59f8b10de854ba289e5e545d1eb9fee0f56` remains in the
   diff and is named in the PR body as the documented rollback
   path; reverting is a one-line change.

5. **Verification surface is authoring-time API evidence,
   not the fixing PR's checks.** The `gaze-ai-report` job
   runs only on push to `main` (`if: github.event_name ==
   'push' && github.ref == 'refs/heads/main'`), so the fix
   PR does not exercise the repinned job. The pinact check
   does not run on this PR either: `ci_checks.yml` pins
   org-infra reusable CI at v0.8.0 (`bd3718a`), whose job
   set has no pinact job -- the failing observation came
   from the v0.9.0 dependabot bump branch (PR #864). The
   PR proves correctness through the authoring-time API
   verification (decision 3) and the written line's
   compliance with the pinact `--verify` contract; the
   check first runs where v0.9.0 is pinned, including the
   bump PR itself. This keeps the cross-repo consumption
   opt-in and self-describing -- the version comment names
   the consumed Gaze release (Principle I, Autonomous
   Collaboration), and the pin stays a published-artifact
   reference with no coupling (Principle II, Composability
   First).

## Risks / Trade-offs

- **Upstream tag mutation**: the v2.3.0 tag could be retagged
  between authoring-time verification and merge. Mitigated by
  re-running the verification commands at authoring time;
  once written, the 40-character SHA is immutable. pinact
  `--verify` re-checks the tag-to-SHA equality wherever
  org-infra reusable CI v0.9.0 runs (starting with the bump
  PR #864), not on this PR.
- **Byte-identity claim wrong**: if `reusable_gaze_ai_report.yml`
  differs between the two SHAs, the "zero behavior change"
  premise fails. Mitigated by the decision-3 content check
  before the edit; on mismatch, stop and reassess rather than
  merge an unverified behavior delta.
- **Sixteen commits of drift accepted**: the pin moves forward
  past tooling bumps, a persona rename, secret redaction in
  adapter stderr, `gaze init --ci`, and docs. Accepted because
  the consumed workflow file is byte-identical, so the
  referenced surface is unchanged, and no tag points at
  `8bdfb59` -- waiting is not an option.
- **Min-age exposure**: v2.3.0 was tagged 2026-11-01; the job
  runs `pinact --verify-min-age`, but no `min_age` is
  configured (no `--min-age`, no `PINACT_MIN_AGE`, no
  `.pinact.yaml`), so the young tag does not trip it. Noted
  and accepted.
- **No new tests**: the change is one configuration line; its
  correctness is verified by the authoring-time API checks
  (tasks 1.1-1.2), reproducible by any contributor. CI-only
  changes are exempt from the documentation-issue gate.
- **Sibling delta divergence**: the unarchived `gaze-ai-report`
  delta hardcodes the same `uses:` reference at the untagged
  SHA `c133b1180f26bf0b2ae433d22f6c7adf9fbff992` with no
  version comment. Archiving both deltas unchanged would leave
  `specs/ci` and `specs/ci-local` asserting contradictory
  MUSTs, and a spec-directed correction back to `c133b118`
  would re-break the pinact gate. Accepted as documented
  governance debt: this change records the divergence instead
  of editing another change's artifacts (commit scope); the
  pre-archive follow-up repins `gaze-ai-report` to a tagged
  commit or marks it superseded.
