---
tag: fullsend-inference
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:17:03Z
identity: fullsend-inference-20261102T021703-jay-flowers
tier: draft
---

Fullsend inference-auth gotcha (surfaced triaging unbound-force issue #730): the `.fullsend/config.yaml` WIF settings — `inference.project` and `inference.wif_provider`, backed by the `FULLSEND_GCP_PROJECT_ID` and `FULLSEND_GCP_WIF_PROVIDER` repo secrets — carry the Vertex inference route that Claude Code agent runs in all four org repos (unbound-force, replicator, gaze, dewey) depend on; they are not OpenAI credentials even though the repos share one OpenAI account. Corroboration: fullsend v0.46.0 renamed the setup flags `--inference-wif-provider`/`--inference-project` to `--vertex-wif-provider`/`--vertex-project`. Deleting those secrets or the GCP identity-pool provider mapping (as issue #730's original body checklist proposed) would break every agent run. A self-deployed agentgateway replaces only the long-lived OpenAI API key: keep Vertex on WIF, allowlist the repos via strict GitHub-OIDC jwtAuth, remove per-model x-api-key, and route only `gateway/<model>` entries through it. The gateway route requires fullsend v0.47.0+, and org CLI pins diverge (v0.47.1, v0.43.0, v0.38.0, v0.38.0), so per-repo version upgrades are the first prerequisite of any gateway rollout.
