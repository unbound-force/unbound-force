---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:17:05Z
identity: issue-triage-20261102T021705-jay-flowers
tier: draft
---

When triaging an issue whose plan includes an org-wide tool upgrade as a prerequisite, verify the actual per-repo version pins and scan the release notes of every major version the upgrade crosses. In unbound-force issue #730 (fullsend inference-gateway rollout), the "upgrade all four repos to v0.47.0+" step looked like a routine bump, but the oldest pins (v0.38.0) cross fullsend v0.46.0 — a breaking release that removes per-org installation, requires an explicit inference.auth in repos.yaml (repos without it fail install/status/converge; the fix is `fullsend repos set-default defaults.inference.auth vertex-wif`), and renames the Vertex install flags to --vertex-* with no aliases. v0.46.0 also ships an openai-wif repos-install auth option (GitHub only) that the issue's gateway-vs-long-lived-key framing never considered. Upgrade steps that cross breaking releases should be extracted into their own migration-tracked issue with rollback notes, not left as a checkbox inside an experimental rollout plan.
