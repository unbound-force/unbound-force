---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:19:38Z
identity: issue-triage-20261101T221938-jay-flowers
tier: draft
---

Documentation-impact triage in unbound-force/unbound-force must evaluate two independent gates separately. The AGENTS.md documentation gate (file a repo-local docs issue before PR merge) fires on user-facing capability surface — including new plugin tools shipped to uf-init consumers through scaffold-mirrored assets such as .opencode/plugins/review-dispatch/index.ts — even when the change is an internal extraction that preserves all observable command behavior. The content-pipeline significance thresholds (divisor-curator blog/tutorial) fire only when a new user-facing capability debuts (new agent, new CLI command or subcommand, architectural migration, new workflow pattern). A change can trip the documentation gate while both content thresholds stay cold; the correct disposition is a docs-impact tracking issue with a recorded no-blog/no-tutorial rationale — the gate records the assessment, the thresholds decide the content.
