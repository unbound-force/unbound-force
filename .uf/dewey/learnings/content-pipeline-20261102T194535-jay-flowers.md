---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-02T19:45:35Z
identity: content-pipeline-20261102T194535-jay-flowers
tier: draft
---

Documentation-gate exemptions in repair-change specs must compare against the broken runtime state, not the frozen contract. A bugfix that restores a user-visible workflow step triggers the documentation gate even when the tool's output contract is byte-frozen: fix-uf-finale-ci-poll-gh-schema tasks.md 5.3 claimed "no user-facing behavior change (output contract frozen per design D4)" while the same proposal's Principle I alignment called the breakage "an autonomy regression this change eliminates" — the spec review (2026-11-02) flagged the contradiction as HIGH. Precedent learning/content-pipeline-20261016T172845-jay-flowers (PR #744, fix-tool-args-schema-mismatch) already established the split: the documentation gate compares workflow behavior against the broken state (CHANGELOG entry plus both docs issues — in-repo per the AGENTS.md documentation gate, unbound-force/website per the constitution's Cross-Repo Documentation rule), while only the blog/tutorial significance thresholds compare against the pre-break baseline (restored capability is not new capability). Rule: a frozen output contract justifies "no command-doc edits required", never "no user-facing behavior change"; and docs-gate disposition wording must not contradict the proposal's own constitution-alignment framing.
