---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:41Z
identity: issue-triage-20261101T222241-jay-flowers
tier: draft
---

Issue-triage pattern (Guard lens): follow-up issues that complete a deferred acceptance criterion sometimes carry the deferred work's leftovers as inline "Optional" scope. When an optional scope item touches a protected gatekeeping value — coverage ratchets, CI enforcement, severity definitions, thresholds — recommend extracting it into its own tracked issue even when the author annotates it "needs its own spec". The annotation documents intent, but the issue's scope list is what implementers execute; an inline optional gate item invites bundling a gate-value change into unrelated wiring work. Verified instance: unbound-force issue #869 (uf-gate FullSend wiring) lists the internal/gate CI coverage ratchet (documented 90.7%, not enforced) as optional scope; the author's own annotation marks the correct boundary, and extraction makes that boundary structural rather than prose-only.
