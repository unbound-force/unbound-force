---
tag: feedback-triage
author: jay-flowers
category: decision
created_at: 2026-11-07T17:06:18Z
identity: feedback-triage-20261107T170618-jay-flowers
tier: draft
---

When maintainer feedback bundles two remediations for one pin seam, split the disposition by axis. Mutual consistency of duplicated pin sites is convention-backed and executable: accept it — the repo's mechanism of record is a parity drift test (parse both sites, assert version equality, TestEmbeddedAssets_MatchSource precedent) or single-sourcing the CI step through the Makefile variable that owns the pin. Recording dist.integrity for a pin resolved inside the package manager's own registry verification is an accepted-risk documentation improvement: author-decides — a content-anchored hash beside the pin goes silently stale on bumps because nothing checks it, so if adopted it must carry the resolution command and the re-resolution-on-bump obligation, or the author declines citing the accepted-risk record. PR 957 item 4 precedent (Makefile PLUGIN_TEST_NPM twin of ci_security.yml osv_plugin_transitive): ACCEPT scoped to the parity guard, AUTHOR-DECIDES on the integrity recording.
