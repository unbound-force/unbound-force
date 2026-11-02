---
tag: uf-workflow
author: jay-flowers
category: gotcha
created_at: 2026-10-31T19:52:43Z
identity: uf-workflow-20261031T195243-jay-flowers
tier: draft
---

A PR body or commit message composed from session memory can claim work the committed diff does not contain. Verified reviewing PR #890 (opsx/reviewer-capabilities-init-survival, 2026-10-31): both the PR body and the head commit message claimed "Correct the stale CLI asset-count test range ([71-75] -> [73-77]) for the 78-asset inventory", but cmd/unbound-force/main_test.go was untouched by the PR and at both base and head TestRunInit_FreshDir asserted [74-78] with a comment deriving 79 total embedded assets — neither claimed range existed anywhere in the committed tree, and the claimed "78-asset inventory" contradicted the committed comment's count of 79. The range edit had evidently informed the local CI-parity run from the working tree but was never committed. Rule: derive a PR's change list from the committed diff (git diff <base>...<head> --stat / numstat), not from recollection of edits made during the session; reconcile every claimed change against the diff before composing the PR body, and treat any working-tree-only fix that made a local green run pass as either committed or explicitly excluded from the change record.
