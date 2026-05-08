---
tag: ollamaproxy
category: pattern
created_at: 2026-05-03T22:01:21Z
identity: ollamaproxy-2
tier: draft
---

When adding a new command that shares lifecycle infrastructure with an existing command (ollama-proxy reusing gateway patterns), extract shared utilities to dedicated packages BEFORE building the new command. The ollama-proxy change needed both internal/pidfile/ (from gateway/pid.go) and internal/auth/ (from gateway/provider.go+refresh.go). Extracting these first as independent task groups meant the new proxy package could be built cleanly without importing internal/gateway/ at all. The alternative — importing gateway directly — would have created a backwards dependency (proxy depends on gateway at the package level), which is architecturally incorrect since the proxy should be independent of the gateway.
