---
tag: scaffold-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:44Z
identity: scaffold-ownership-20261102T035444-jay-flowers
tier: draft
---

Scaffold ownership comments must state precise --force semantics: "user-owned" means skip-on-plain-rerun, not force-protection — only isNeverOverwrite files survive --force. In unbound-force (issue #649), the uf/review-matrix.yaml ownership comment in isToolOwned claimed --force would not replace the matrix, while the write path overwrites every non-never-overwrite asset under --force; the newer adjacent uf/model-matrix.yaml comment in the same function already stated the correct semantics ("--force remains the explicit replace path"), leaving two adjacent ownership comments contradicting each other. When reusing an ownership branch as a "precedent" for a new comment, verify the borrowed wording against actual overwrite behavior. Additionally, when a deliberate project-owned classification stops canonical updates from flowing to a structurally versioned file (tool-owned v2/v3 review-matrix schemas update while the user-owned matrix does not), pair the trade-off with a drift signal or fail-closed downstream error, as the #884 reviewer-capabilities fix did.
