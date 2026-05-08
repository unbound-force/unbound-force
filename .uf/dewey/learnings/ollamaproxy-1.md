---
tag: ollamaproxy
category: pattern
created_at: 2026-05-03T22:01:16Z
identity: ollamaproxy-1
tier: draft
---

The ollama-proxy spec review surfaced a pattern for API proxy security hardening that should be applied to all future proxy commands: (1) SSRF prevention via loopback-only URL validation for upstream endpoints, (2) token redaction in error responses using a dedicated redactToken function, (3) model name validation against a safe character set before URL interpolation to prevent path traversal, (4) max request body size limits (10MB), and (5) log file permissions at 0o600. The Adversary reviewer caught all five of these from first principles — they form a reusable security checklist for any command that translates between API formats and forwards requests to cloud providers.
