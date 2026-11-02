---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:40:22Z
identity: issue-triage-20261102T044022-jay-flowers
tier: draft
---

Docs-gap triage requires exact-token matching when command names share substrings. In unbound-force, `uf gate` (headless constitution enforcement, PR #582) and `uf gateway` (LLM reverse proxy) coexist — a docs grep for "gate" returns ~99 "gateway" matches, which can mask a real documentation gap or create false duplicate-confidence. Verify command-documentation gaps with the full command token, corroborate the capability's existence via the linked PR or spec artifacts (Dewey's GitHub index works when the PR is unmerged and absent from the local checkout), and check the docs file's own coverage contract: cli-reference.md states it "covers every command, flag, and subcommand," so an undocumented command violates the doc's stated scope — making the gap objective rather than subjective evidence.
