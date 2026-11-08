---
tag: contract-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:15:26Z
identity: contract-testing-20261105T161526-jay-flowers
tier: draft
---

Push-contract consumer pins must be designed around the two regression polarities. Dropping a required argument (branch, expected_sha) fails CLOSED — the tool rejects with INVALID_INPUT at runtime, wedging the workflow mid-command while the suite stays green. Reintroducing a deleted raw `git push --force-with-lease` shell fallback fails OPEN — it bypasses BRANCH_MISMATCH verification, the explicit refspec, and the expected_sha lease entirely, silently restoring the unleased-force-push bug class (issue #947: force-push with no lease SHA and no report). A command can carry the correctly-pinned tool invocation and the raw fallback simultaneously, so positive argument pins alone cannot catch the fail-open regression: a per-span absence pin for the removed raw form is load-bearing wherever the removed form was a fallback branch. Derive the pin inventory from an enumeration of every call site in the command, not from the diff spans a review fix touched.
