---
tag: fullsend-inference
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:16:55Z
identity: fullsend-inference-20261102T021655-jay-flowers
tier: draft
---

In Fullsend-managed repos, the WIF settings in .fullsend/config.yaml (inference.project, inference.wif_provider) and the FULLSEND_GCP_WIF_PROVIDER / FULLSEND_GCP_PROJECT_ID secrets route Vertex inference for Claude Code — they are not the OpenAI path. Deleting those secrets or the GCP identity-pool provider mapping breaks every agent run in the affected repos. The self-deployed agentgateway route (requires fullsend v0.47.0+) is OpenAI-only: keep Vertex on per-repo WIF, mount only the OpenAI key in Secret Manager, enforce strict GitHub-OIDC jwtAuth with a repo allowlist, remove x-api-key on every model, and bound spend with localRateLimit plus --max-instances 1. GPT models through the gateway run on pi and Codex; Claude Code is not a supported path for GPT models (Messages-to-OpenAI translation untested). Tracked in unbound-force/unbound-force#730.
