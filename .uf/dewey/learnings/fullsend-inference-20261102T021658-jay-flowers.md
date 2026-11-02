---
tag: fullsend-inference
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:16:58Z
identity: fullsend-inference-20261102T021658-jay-flowers
tier: draft
---

unbound-force org inference topology (per issue #730 and its correcting comment, corroborated by fullsend v0.46.0/v0.46.1 release notes in unbound-force PR #701): the Fullsend WIF settings (inference.project / inference.wif_provider in .fullsend/config.yaml, the vertex-wif auth route) are the Vertex inference path that Claude Code runs on in all four hero repos (unbound-force, replicator, gaze, dewey) — deleting those secrets or the GCP identity-pool provider breaks every agent run. The self-deployed agentgateway (Cloud Run; gateway route requires fullsend v0.47.0+, and repo pins diverge: v0.47.1/v0.43.0/v0.38.0/v0.38.0) centralizes the OpenAI API key only; only gateway/<model> entries route through it while all other models stay on Vertex with their own credentials. Triage gotcha: when a correcting comment supersedes destructive steps in an issue body (here the "delete WIF secrets/provider" checklist items), reconcile the body before the issue is scheduled — agents working checklists top-to-bottom will execute stale harmful steps.
