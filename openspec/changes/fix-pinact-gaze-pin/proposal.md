## Why

The `Standardized CI / Verify action versions` check fails on
PR #864 ([run 40817531850][run]) with pinact code 005:
`.github/workflows/ci_local.yml:101` pins the Gaze reusable
workflow to the untagged merge commit `8bdfb59` with no version
comment. The check is an enforced gate -- `release.yml` gates
release preflight on Standardized CI -- so the red check blocks
the pipeline.

The pin cannot be repaired by adding a comment. pinact code 001
requires the version comment to name a tag whose commit exactly
equals the pinned SHA, and no Gaze tag points at `8bdfb59` (it
sits between v2.2.0 and v2.3.0, the merge commit of Gaze PR
#280). The pin must move to a tagged commit. Originating issue:
#961.

[run]: https://github.com/unbound-force/unbound-force/actions/runs/40817531850

## What Changes

Re-pin `unbound-force/gaze/.github/workflows/reusable_gaze_ai_report.yml`
in `.github/workflows/ci_local.yml` from the untagged commit
`8bdfb59f8b10de854ba289e5e545d1eb9fee0f56` to the v2.3.0 tag
commit `8031d41943b06407cb92322791b177b41b7b2775`, adding the
`# v2.3.0` version comment beside the SHA:

```diff
-    uses: unbound-force/gaze/.github/workflows/reusable_gaze_ai_report.yml@8bdfb59f8b10de854ba289e5e545d1eb9fee0f56
+    uses: unbound-force/gaze/.github/workflows/reusable_gaze_ai_report.yml@8031d41943b06407cb92322791b177b41b7b2775  # v2.3.0
```

One line, one file. The tag-to-SHA mapping and the byte-identity
of the referenced workflow file are re-verified at authoring
time via the GitHub API; the previous SHA is recorded in the fix
PR as the documented rollback path.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `ci`: the Reusable Workflow Pinning requirement gains
  version-comment verifiability. Every SHA-pinned `uses:`
  reference MUST carry a version comment naming a tag whose
  commit exactly equals the pinned SHA (the pinact `--verify`
  contract), and the Gaze reusable workflow pin in
  `ci_local.yml` MUST point at the v2.3.0 tagged commit.

### Removed Capabilities

- none

## Impact

- **Files**: `.github/workflows/ci_local.yml:101` only. The SHA
  `8bdfb59` appears exactly once in the repository.
- **Behavior**: none. `reusable_gaze_ai_report.yml` is
  byte-for-byte identical between the two SHAs (170 lines,
  matching SHA-256, verified via the GitHub API), so inputs,
  secrets, permissions, and job behavior are unchanged. The
  `gaze-ai-report` job runs only on push to `main`, so the fix
  PR does not exercise it post-repin.
- **Gates**: the `Standardized CI / Verify action versions`
  check changes from red to green. The gate is strengthened,
  not bypassed -- suppressing the finding via `.pinact.yaml`
  ignore rules is rejected as supply-chain gate weakening.
- **Gatekeeping conditions** (pinned-dependency update policy):
  documented CI reliability fix (the failing check); target
  immutable 40-character SHA verified against the upstream
  v2.3.0 tag; existing gates remain enforced; rollback to
  `8bdfb59f8b10de854ba289e5e545d1eb9fee0f56` documented;
  reproducible validation evidence (the `gh api` commands);
  normal review and CI approval to follow.

## Constitution Alignment

Assessed against the Unbound Force org constitution
(v1.4.0, `.specify/memory/constitution.md`).

### I. Autonomous Collaboration

**Assessment**: PASS

The change improves the self-describing quality of the
cross-repo artifact reference: the version comment names the
consumed Gaze release beside the immutable SHA. No hero
interaction, artifact envelope, or exchange format changes.

### II. Composability First

**Assessment**: PASS

Gaze and this repository remain independently installable and
usable. The reusable workflow reference stays an opt-in, pinned
consumption of a published Gaze artifact; no mandatory coupling
is introduced or removed.

### III. Observable Quality

**Assessment**: PASS

The fix restores a machine-verified quality signal: the pinact
`--verify` check passes. The tag-to-SHA mapping and the
byte-identity claims are backed by reproducible `gh api`
commands any contributor can re-run.

### IV. Testability

**Assessment**: PASS

The change's observable effect -- the `Standardized CI / Verify
action versions` check passing on the fix PR -- is verifiable
in isolation without external services beyond GitHub Actions.
No new code components are introduced.

### V. Security by Default

**Assessment**: PASS

This is the governing principle. The pin remains a full
40-character immutable commit SHA, never a mutable tag ref, and
gains a verifiable version comment (Constitution Principle V:
"CI pipelines MUST pin actions and reusable workflows by commit
SHA, not mutable tags"). The supply-chain gate is strengthened,
not weakened, and every Gatekeeping condition for a
pinned-dependency update is satisfied.

### VI. Deterministic Execution

**Assessment**: N/A

No command, agent, or skill instruction text changes. The
verification procedure is two deterministic `gh api` commands
executed once at implementation time, not re-reasoned prose.
