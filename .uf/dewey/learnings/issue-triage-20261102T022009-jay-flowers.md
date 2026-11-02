---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:20:09Z
identity: issue-triage-20261102T022009-jay-flowers
tier: draft
---

Docs-issue triage must re-verify tool identifiers named in an issue against the current registered surface before the issue is executed, because spec-time names go stale when a later change absorbs the mechanism. Verified on unbound-force/unbound-force issue #736 (docs: persisted unleash summaries): it names `persist_unleash_summary` per its openspec/changes/persist-unleash-summary spec (FR-001..FR-008), but the unleash-deterministic-tools change (PR #745) absorbed that tool into the 8-parameter `write_unleash_summary` ("the mechanism of record"), and the later unleash-pipeline-hardening change removed the redundant alias registration — the uf-workflow tool map (.opencode/plugins/uf-workflow/index.ts) registers only write_unleash_summary, and internal/scaffold/scaffold_test.go's step-10 test errors if command prose references the superseded name. Executing the issue as titled would document an unregistered tool and violate FA-001/VB-004 (content reflects current state). Verification recipe for docs issues naming tools: repo-grep the identifier; the plugin tool-map registration and scaffold drift tests are the authoritative current-name sources; CHANGELOG supersession entries record the history. Disposition: update the issue (title, summary bullet, tool-reference checklist item, schema list) rather than close it — the docs gap itself remained real (docs/ had zero unleash-summary mentions while code, connecting tests, and scaffold mirrors all shipped). Where the absorbing change carries its own docs issue (#739, whose artifact-contract scope fix was deferred to #743), reconcile ownership so the durable artifact contract (.uf/artifacts/unleash-summary/<slug>.json: producer write_unleash_summary, consumer uf_finale_pr_context) is documented exactly once.
