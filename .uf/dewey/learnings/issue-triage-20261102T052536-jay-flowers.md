---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:25:36Z
identity: issue-triage-20261102T052536-jay-flowers
tier: draft
---

When triaging audit-filed issues that bundle multiple gaps, assess each gap's fix for data-compatibility prerequisites before accepting the bundle as one unit of work. Precedent: unbound-force issue #561 bundled two input-validation gaps — the ReadEnvelope schema-validation fix is prerequisite-blocked by #557 (the envelope schema requires `context` while the writer emits it omitempty, so enforcement would reject existing artifacts), while the gateway ProjectDir traversal fix is independent and immediately actionable. Different packages + different root causes + different blocking status → recommend splitting per gap so the unblocked fix is not delayed behind the blocked one. Additionally, verify audit-issue citations against HEAD at triage time: #561's claimed consumers (internal/metrics/collect_*.go) had already been removed from HEAD, and its cited reference implementations (vibe-check metrics/security.go) live in the zero-dot-force sibling repo, meaning "reuse the pattern" fixes must be re-implemented locally rather than imported.
