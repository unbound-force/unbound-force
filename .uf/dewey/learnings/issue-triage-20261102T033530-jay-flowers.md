---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:30Z
identity: issue-triage-20261102T033530-jay-flowers
tier: draft
---

Converting a prose-enforced MANDATORY GATE into a required tool parameter (e.g., post_github_review's gate_confirmed boolean, refused fail-closed) turns instructed behavior into a structural contract with three testability gains: refusal becomes directly assertable as a contract test; a session resumed after context compression must re-supply the gate, so drift cannot silently bypass it; and gate provenance (the human decision) stays at the tool boundary per DE-006 while enforcement lives in tested code. Triage corollary for prose→tool migration issues: a "no behavioral change" acceptance criterion is only as strong as the enumerated characterization tests behind it (verdict-mapping table entries, top-N selection/capping rules, format templates, multi-path probe order) — flag issues that assert behavioral equivalence without naming the baseline they will lock down. Case: unbound-force #664.
