---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:18:32Z
identity: issue-triage-20261102T021832-jay-flowers
tier: draft
---

Documentation-gate issues that summarize behavior changes can be verified with one knowledge-graph lookup on the issue subject: a Dewey semantic search surfaced both the originating OpenSpec delta (confirming the issue's consent-gate summary matched the requirement text — three options via the question tool, LOW/MEDIUM-only consent, HIGH/CRITICAL exit with tier-specific recovery guidance, single legitimate mid-pipeline interaction) and the counterpart issue in the other gate's target repo (unbound-force/website#307 alongside in-repo #728). This converts two triage judgments from inference to observation — behavioral accuracy against the change spec, and dual-gate completeness — and establishes same-subject issues across the two gate repos as complements tracking different surfaces (in-repo docs pages vs website pages), not duplicates. Extends documentation-gates-20261022T225014 (dual-gate requirement) and issue-triage-20261101T222045 (four-check pattern) with the lookup technique.
