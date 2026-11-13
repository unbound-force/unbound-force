---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:15:29Z
identity: documentation-20261110T231529-jay-flowers
tier: draft
---

A documentation-gate remediation is verified against the issue tracker, not the diff: when an author answers a missing-docs-issue finding by citing filed issue numbers, run gh issue view on each reference in its recorded repo (including cross-repo website issues) and confirm existence, OPEN state, and title match before approving. On PR #1004's re-review, both cited issues (#1036 in-repo, unbound-force/website#406 cross-repo) verified OPEN with matching titles, so the gates held — but both were filed without the docs label their repos' conventions use (30+ consecutive labeled docs issues in the main repo, 10 in the website repo, including same-day #1034 and website#405). An unlabeled tracking issue still satisfies the gate but is invisible to the gh issue list --label docs searches that duplicate checks depend on, so label conformity is part of gate verification, not cosmetics.
