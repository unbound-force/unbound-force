---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T07:12:25Z
identity: instruction-quality-20261116T071225-jay-flowers
tier: draft
---

The instruction linter's IL-050 duplication rule and byte-exact pin tests interact: a Go contract test that pins a prose block verbatim across live and scaffold copies (e.g. TestCommandContracts_SyntheticPromptNotConsent) turns synchronized instruction copies into a tested contract, but the linter still anchors IL-050 at each changed copy on every run. When a change deliberately ships synchronized instruction copies for per-role visibility, pair the pin test with an instruction-lint-disable-next-line IL-050 comment placed outside the pinned block (on the preceding line, synced to the scaffold mirror) carrying the reason — or plan consolidation into a single loadable source — so the acceptance is recorded instead of re-litigated by every lint run.
