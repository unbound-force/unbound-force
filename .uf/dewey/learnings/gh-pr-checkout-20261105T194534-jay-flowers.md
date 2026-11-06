---
tag: gh-pr-checkout
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:45:34Z
identity: gh-pr-checkout-20261105T194534-jay-flowers
tier: draft
---

gh pr checkout on a fork PR the runner cannot push to ("Allow edits by maintainers" disabled) sets branch.<b>.remote to the base repository and branch.<b>.merge to refs/pull/<N>/head. Both config values are present, so any "is the push target configured?" presence check passes — but refs/pull/<N>/head is read-only on GitHub and every push to it is rejected. Workflow commands that record a push target from gh pr checkout must treat a merge ref starting with refs/pull/ as a guaranteed-failure condition and stop before running phases that produce commits. General trap: validating git config presence instead of push-target usability.
