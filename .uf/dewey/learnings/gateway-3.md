---
tag: gateway
category: pattern
created_at: 2026-04-28T00:44:45Z
identity: gateway-3
tier: draft
---

When the Vertex AI gateway's background token refresh fails (e.g., gcloud ADC credentials expire), the old token must be cleared immediately rather than preserved. The original implementation logged the error but kept forwarding the stale token, producing cryptic ACCESS_TOKEN_TYPE_UNSUPPORTED (401) errors from Vertex. The fix clears both the token string and tokenExpiry timestamp atomically under tokenMu write lock when refresh fails, causing PrepareRequest to return a clear "Re-authenticate: gcloud auth application-default login" error instead. This is the same "fail explicitly rather than silently misrouting" principle that was applied to the global-region fallback in the earlier gateway-global-region-error change. The asymmetric failure behavior is intentional: background refresh clears the token (the token is likely permanently invalid), while proactive refresh in PrepareRequest preserves the token (it may still be valid until actual expiry). The proactive refresh uses sync.Mutex.TryLock() for deduplication with a 5-second timeout via goroutine+channel to prevent hung gcloud subprocesses from blocking HTTP requests.
