---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:25:29Z
identity: issue-triage-20261102T052529-jay-flowers
tier: draft
---

Security-issue triage: verify headline code claims against the current tree, and treat auxiliary citations as potentially stale. In unbound-force issue #561, the cited reference implementation (metrics/security.go ValidateProjectPath) and cross-agent consumers (internal/metrics/collect_*.go) no longer exist in the repo, yet both headline gaps were confirmed present (artifacts.ReadEnvelope at internal/artifacts/artifacts.go:85-95 unmarshals without calling the existing schemas.ValidateBytes; gateway opts.ProjectDir reaches filepath.Join-based file writes at internal/gateway/gateway.go:561-571 with no traversal validation). Stale auxiliary citations refine the fix approach (no reusable function, smaller demonstrated blast radius) but do not invalidate core claims. Companion pattern: validation infrastructure shipping in a repo does not imply runtime enforcement — a Validate helper exercised only by tests, beside a registered schema (schemas/envelope/) that the runtime reader never loads, creates a false sense of validated inputs. Trace the runtime consumption call path, not just the presence of validation helpers.
