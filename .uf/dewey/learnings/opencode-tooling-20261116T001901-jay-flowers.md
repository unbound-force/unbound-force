---
tag: opencode-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-16T00:19:01Z
identity: opencode-tooling-20261116T001901-jay-flowers
tier: draft
---

The opencode Read tool's hashline rendering can display a hyphenated string as its slashed variant (e.g. showed recordPath(root, "opsx/change-name") where the file on disk has recordPath(root, "opsx-change-name") — verified by grep). When transcribing exact string literals from a hashline view into an edit (path keys, identifiers), verify the bytes with grep first. In human-decisions-store.test.ts this bit twice: the test-side recordPath helper expects the PRE-sanitized key ("opsx-change-name") while the library calls readHumanDecisionRecord/appendHumanDecisions sanitize the RAW branch key ("opsx/change-name") internally via resolveHumanDecisionsKey.
