---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:31Z
identity: issue-triage-20261101T221531-jay-flowers
tier: draft
---

Issue-triage duplicate assessment for documentation-class issues must compare deliverable tracks, not subject matter. Observed triaging unbound-force issue #870 (tutorial: /uf.fix-ci workflow): the duplicate candidates — #862 (docs: /uf.fix-ci new command, the documentation-gate issue filed at implementation time) and #873 (docs: fetch_pr_checks/fetch_failed_run_logs plugin tools and /uf.fix-ci re-route) — overlap the tutorial's subject command but carry different deliverables. The repo tracks command reference documentation via documentation-gate issues (AGENTS.md gate, filed before PR merge) and end-to-end workflow tutorials via separate tutorial significance-threshold issues (precedent: /uf.rebase-pr tutorial #855); the uf-fix-ci change proposal explicitly designates the tutorial issue (#870) as the re-disposed deliverable while the docs issue (#862) only records the dispositions. Same subject command does not imply duplicate: classify by deliverable type and governing threshold, and check the originating change's content-pipeline disposition record for which issue canonically carries which deliverable before returning a duplicate verdict.
