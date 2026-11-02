---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:32Z
identity: issue-triage-20261102T003332-jay-flowers
tier: draft
---

Documentation-gate and content-pipeline obligations in an umbrella-migration issue family bind to the phase that lands user-facing changes, not to verification-only phase trackers. Issue #802 (Phase 3 — dewey: OpenCode v2 verification pass, umbrella #797) is the precedent: its scope mandates running the repo under the v2 binary with the existing v1-shaped config unchanged, defers native-shape conversion to the Phase 5 cutover (umbrella decision D1), and routes any discovered breakage upstream (filed against opencode, not fixed in-repo). A phase whose done-when criteria require the same committed tree cannot change CLI surface, agent capabilities, installation steps, or workflows, so CHANGELOG/AGENTS/README updates, docs issues, and blog/tutorial opportunities attach only at the conversion phase (and blog coverage of the migration itself belongs to the umbrella, not a per-repo phase tracker). When triaging migration phase trackers, check the done-when criteria for a tree-change exclusion before applying documentation checklist items — the internal-only exemption extends to verification work that leaves the committed tree untouched.
