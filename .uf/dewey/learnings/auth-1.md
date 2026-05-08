---
tag: auth
category: gotcha
created_at: 2026-05-03T22:01:11Z
identity: auth-1
tier: draft
---

When extracting shared infrastructure from an existing package (like token refresh from gateway), the spec review will catch incomplete extraction scope. In the ollama-proxy change, the initial design proposed extracting only RefreshVertexToken and RefreshLoop, but all 5 reviewers independently flagged that this was insufficient: RefreshLoop was also used by BedrockProvider, and the proactive refresh and atomic invalidation patterns would need to be reimplemented. The resolution was to extract a full TokenManager struct to internal/auth/ that encapsulates the entire lifecycle. The key lesson: when extracting shared code, trace ALL callers across the entire codebase, not just the caller that motivated the extraction. The Bedrock dependency was invisible from the Vertex-focused perspective.
