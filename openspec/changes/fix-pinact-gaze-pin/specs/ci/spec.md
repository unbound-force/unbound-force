<!-- capability: ci
     change: fix-pinact-gaze-pin
     originating_issue: 961
     Version-comment verifiability for SHA-pinned uses:
     references, mirroring the pinact --verify contract
     (codes 001 and 005). -->

Requirement numbering: this change allocates no new FR ids.
The modified `Reusable Workflow Pinning` requirement is
unnumbered in the live spec (`openspec/specs/ci/spec.md:46`)
and stays unnumbered here as a documented
unnumbered-predecessor exception; MP-002's no-renumbering
rule leaves no existing id to reuse. The CI-FR-001..CI-FR-020
range is reserved for future ci-capability changes that
introduce numbering. No sibling ci delta allocates numbered
ids -- `ts-plugin-coverage-gate` (2026-10-22) predates the
numbering convention -- and sibling capabilities' numbered
requirements (for example review-dispatch's RD-FR ids) live
in their own capability deltas. Requirement ids, when
introduced, MUST stay inside the allocated range.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Reusable Workflow Pinning

The consumer workflow MUST reference every reusable workflow
using a pinned commit SHA, not a branch name or mutable tag.
Each SHA-pinned `uses:` reference MUST carry a version comment
naming a tag whose commit exactly equals the pinned SHA, so
that pinact `--verify` can confirm the pin (codes 001 and
005). When a pinned SHA is untagged, the pin MUST move to a
tagged commit whose referenced workflow file is unchanged, and
the previous SHA MUST be recorded in the fixing pull request
as the documented rollback path. The `ci_crapload.yml` SHA
MUST remain managed by dependabot for automated updates.

#### Scenario: Dependabot updates reusable workflow SHA

- **GIVEN** a new commit is pushed to the default branch of
  `complytime/org-infra`
- **WHEN** dependabot runs its daily check
- **THEN** dependabot SHOULD open a PR updating the pinned
  SHA in `ci_crapload.yml`

#### Scenario: SHA pin lacks a version comment

- **GIVEN** a workflow pins a reusable workflow by commit SHA
  with no version comment beside the SHA
- **WHEN** the `Standardized CI / Verify action versions`
  check runs pinact `--verify`
- **THEN** the check MUST fail with pinact code 005 (SHA-pinned
  action requires a version comment for verifiability)

#### Scenario: Version comment names a different commit

- **GIVEN** a SHA-pinned `uses:` reference carries a version
  comment whose named tag resolves to a commit other than the
  pinned SHA
- **WHEN** the `Standardized CI / Verify action versions`
  check runs pinact `--verify`
- **THEN** the check MUST fail with pinact code 001
  (`action_version must be equal to
  commit_hash_of_version_annotation`)

#### Scenario: Untagged pin moved to the tagged commit

- **GIVEN** `ci_local.yml` pins the Gaze reusable workflow at
  an untagged commit between two Gaze releases
- **WHEN** the pin moves to the tagged release commit whose
  `reusable_gaze_ai_report.yml` is byte-identical to the
  untagged commit's copy, with an accurate `# vX.Y.Z` version
  comment
- **THEN** pinact `--verify` MUST pass, the referenced
  workflow's inputs, secrets, and permissions MUST be
  unchanged, and the fixing pull request MUST record the
  previous SHA as the rollback path

Previously: The consumer workflow MUST reference the reusable
workflow using a pinned commit SHA, not a branch name. The SHA
MUST be managed by dependabot for automated updates.

## REMOVED Requirements

None.
