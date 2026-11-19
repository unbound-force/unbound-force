---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T08:42:36Z
identity: determinism-20261116T084236-jay-flowers
tier: draft
---

When a prose→tool migration replaces mechanical prose with a parameter mapping, audit each new parameter against DE-006's boundary: is it a judgment product (triage records, composed text, human gate decisions), routing data the tool should own, or a mechanically derivable environment value? The smell detector is a new prose guard that polices the agent's sourcing of the parameter — e.g. /uf.address-feedback Section 4.7's "Never reuse the frozen Section 1.2 head_sha after fixes have added commits" exists only because `commit` (current HEAD at call time) arrives as a caller-supplied parameter instead of being derived inside write_feedback_triage_artifact, whose factory already shells out to git for the branch-keyed state path. A parameter whose selection rule collapses to a single environment lookup ("both are simply the current HEAD") is a candidate for tool-side derivation: the guard, the parameter, and the failure mode it patches all disappear together. Precedent: write_unleash_summary derives `branch` from git internally and takes only judgment products.
