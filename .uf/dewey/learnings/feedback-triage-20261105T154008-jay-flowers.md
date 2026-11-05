---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-05T15:40:08Z
identity: feedback-triage-20261105T154008-jay-flowers
tier: draft
---

Before ACCEPTing an address-feedback item, re-verify the feedback's factual predicates against the frozen head diff — a 're-confirmed' review comment can describe an earlier commit because the author may have pushed a fix between review time and triage time. Cheap disproofs that need no git archaeology: (1) numstat arithmetic — a file inventory of +31/-131 cannot also contain an additional ~33-line test deletion; (2) manifest membership — a claimed-deleted file that is absent from the changed-files manifest was not deleted by the PR; (3) read the cited doc line at head and check it still says what the feedback quotes. Observed in unbound-force PR 957 item 8: the feedback claimed design.md:58 named TestPluginPackageInputs_RollbackPrerequisites as a deleted rollback safety net, but at frozen head 60ca3d0 the test exists untouched (scaffold_test.go:1825 at head / 1923 at base), its testdata is absent from the 18-file manifest, and design.md D5 reframes both as a disclaimed historical record. Accepting the stale feedback as-written would have directed the author to fix a problem that does not exist at head.
