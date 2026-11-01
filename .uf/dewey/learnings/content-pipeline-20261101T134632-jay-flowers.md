---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-01T13:46:32Z
identity: content-pipeline-20261101T134632-jay-flowers
tier: draft
---

Content-coverage dispositions (blog/tutorial threshold decisions) recorded for a partial-slice change should name the completing slice as an explicit re-evaluation trigger. In the model-matrix-delegation change (PR #894), the proposal's Content coverage bullet correctly dispositioned blog below threshold and tutorial satisfied-by-reference-docs for the repo-first slice, with the reasoning recorded in proposal.md and unbound-force/website#358 — but slice 2 (#889, deploying the matrix and meta-prompt pack to hero repos via uf init) expands the audience from one repo's operators to every uf init consumer, which reopens the significance thresholds (new tool integration arriving in downstream repos). Without a named trigger, the evaluation burden silently transfers to the slice-2 change's own review — the same loss mode recorded in content-pipeline-20261020T125031 (tier routing shipped with documentation gates satisfied but no content evaluation anywhere). Pattern: scope every content-coverage disposition to the capability's full delivery plan and name the slice where the threshold question reopens.
