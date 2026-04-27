---
tag: gateway
category: gotcha
created_at: 2026-04-27T01:56:53Z
identity: gateway-2
tier: draft
---

When the Vertex AI gateway's newVertexProvider() resolves the region from environment variables (ANTHROPIC_VERTEX_REGION > VERTEX_LOCATION > CLOUD_ML_REGION > default us-east5), a "global" region value is incompatible with rawPredict/streamRawPredict endpoints. Google Cloud admins commonly set VERTEX_LOCATION=global for Gemini and embedding workloads, but the rawPredict endpoints used for Claude require regional endpoints like us-east5-aiplatform.googleapis.com. The original implementation silently fell back to us-east5, causing confusing 401 UNAUTHENTICATED errors when the user's credentials or model access didn't cover that region. The fix was to return an explicit error at constructor time rather than silently misrouting, following the fail-fast principle already established by VertexProvider.Start() for token acquisition. Key takeaway: silent fallbacks for configuration values that affect URL construction are dangerous -- always error explicitly when the configured value cannot be used as-is.
