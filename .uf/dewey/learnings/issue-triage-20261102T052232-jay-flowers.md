---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:22:32Z
identity: issue-triage-20261102T052232-jay-flowers
tier: draft
---

Unbound Force supply-chain pinning triage: classify by rule surface. Workflow `uses:` references fall under convention pack CI-001 [MUST] (full 40-char commit SHA, verified at authoring time per CI-002; the severity pack lists "unpinned CI action on mutable tag" as a HIGH Adversary example). Containerfile FROM statements fall under Constitution Principle V's content-hash MUST for dependencies downloaded outside package-manager verification — CI-001's docker:// carve-out (digest pinning as SHOULD) is scoped to workflow container references only and does not extend to Containerfiles, so a mutable-tag FROM (including :latest and version tags like :41, which registries can re-point) remains a MUST violation that cannot be downgraded via the carve-out. Issues reporting either class are objective compliance bugs, not enhancement requests. Precedent: the replicator ci-release-preflight OpenSpec change migrated @v4/@v5/@v6 tags to full SHAs, confirming this is an established remediation pattern. When an org-wide audit issue is filed against the meta repo but the offending files live in sibling repos (e.g., containerfile/, vibe-check), recommend splitting into per-repo issues so fixes are actionable where they land; note that portions citing a repo without file/line evidence (vibe-check in issue #560) need their own enumeration step.
