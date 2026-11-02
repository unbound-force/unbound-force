---
tag: supply-chain-pinning
author: jay-flowers
category: context
created_at: 2026-11-02T05:22:27Z
identity: supply-chain-pinning-20261102T052227-jay-flowers
tier: draft
---

Unbound Force supply-chain pinning rule scope (surfaced by issue #560 triage): Constitution Principle V covers two distinct pinning surfaces with different rule mappings. Workflow `uses:` references are governed directly by the MUST "CI pipelines MUST pin actions and reusable workflows by commit SHA, not mutable tags" and by convention-pack CI-001 [MUST] (full 40-char SHA). Containerfile `FROM` base images are governed by the constitution's broader clause "Dependencies MUST be verified by content hash (SHA256 or equivalent) when downloaded outside a package manager's built-in verification" — CI-001's `docker://` exclusion is workflow-container-scoped and does not extend to Containerfiles, so mutable FROM tags (`:latest`, `fedora:41`) are constitutional violations even though CI-001's text is silent on Containerfiles. When both gap types appear in one build pipeline (unpinned actions in the workflow that builds images from mutable FROM bases), the severity pack's compound escalation rule applies: assess them as a single supply-chain attack surface on the published images, not as independent findings. Triage corollary: an issue filed in the meta repo that names only sibling-repo files cannot be closed by fixes in this repo — split into per-repo issues and keep the meta issue as an org-level umbrella.
