---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:06:06Z
identity: documentation-pipeline-20261110T230606-jay-flowers
tier: draft
---

Docs-only follow-up PRs in unbound-force complete the documentation pipeline when three artifacts verify: (1) a CHANGELOG.md entry in the diff matching the established `- <change-name>: <summary> (<issue>)` format; (2) the in-repo docs tracking issue (label `docs`) open and matching the change; (3) the cross-repo unbound-force/website issue referenced by the tracking issue — review context marks cross-repo links "not validated", so validate it with `gh issue view <N> --repo unbound-force/website` before approving. Doc claims about tool behavior should be verified against three implementation sources: the slash-command file (ask gates and call order), the .opencode/lib module (tool mechanics, resume behavior, resolvable-decision sets), and the schema README (version compatibility claims). A docs-only PR that itself introduces no new capability needs no blog or tutorial issue.
