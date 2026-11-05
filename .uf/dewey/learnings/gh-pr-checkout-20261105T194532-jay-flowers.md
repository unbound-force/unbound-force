---
tag: gh-pr-checkout
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:45:32Z
identity: gh-pr-checkout-20261105T194532-jay-flowers
tier: draft
---

gh pr checkout on a fork PR the current user cannot push to (outside contributor's fork with "Allow edits by maintainers" off, or no push access to the head repository) still writes branch tracking config: branch.<b>.remote points at the base repository and branch.<b>.merge is refs/pull/<N>/head. Both values are set, so "stop if unset" gates over these config keys pass — but refs/pull/* refs are read-only on GitHub and every push to them is rejected. Workflows that derive a push target from gh pr checkout's config (uf.address-feedback Section 1.2.1 / Section 4.4, and any future command that records a push target) must treat a HEAD_MERGE_REF starting with refs/pull/ as a fail-closed no-push-target condition and stop before assessment, triage, commits, or review dispatches run; otherwise the failure lands only at the final push, after all expensive work has run, with no retry path that can succeed.
