---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:45Z
identity: opencode-v2-migration-20261102T023545-jay-flowers
tier: draft
---

Dual-form plugin port test design (validated triaging unbound-force#714, Phase 1 of the opencode v2 migration umbrella #797): because the v1 and v2 runtimes cannot coexist on one machine (both claim the `opencode` command), the automated dual-path suite must be runtime-agnostic — v1-path tests exercise `server()` with fake clients while v2-path tests use a fake `ctx` with a domain-shaped client (session.prompt/wait/context), both under `make plugin-test` — so the 94% statements coverage floor stays enforceable in CI with no real runtime present. Real-binary E2E smoke (one v1 pass, one v2 pass) is correctly scoped as one-time manual validation; continuous dual-compat verification is carried by a separate CI v2-leg issue (#800). Triage corollary: when a maintainer comment supersedes a tracking issue's decisions (on #714, Decision 4 was redirected to Speckit spec #799 and the stale 1.4.10 pin corrected to the verified 1.19.0), the Constitution-IV coverage-strategy requirement is owned by the spec artifact of record; the issue's test workstream remains the technical reference and does not itself need measurable coverage criteria to be a valid tracker. Assessing the body in isolation would both use stale pin data and misattribute a missing-coverage-strategy finding to an issue whose spec now carries that responsibility.
