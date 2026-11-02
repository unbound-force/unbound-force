---
tag: convention-packs
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:38:08Z
identity: convention-packs-20261102T023808-jay-flowers
tier: draft
---

Convention-pack listings in unbound-force/unbound-force drift because each pack deployment updates only some of the surfaces that enumerate the pack set. Triage of issue #719 (verified 2026-11-01): the ci-convention-pack change originally left AGENTS.md, docs/usage.md, and docs/architecture.md stale; later changes fixed the ci and python entries, but both docs pages had re-drifted to 15 of 18 packs (determinism.md, meta-prompt.md, meta-prompt-custom.md missing). AGENTS.md stays exhaustive only because the determinism-codification change added an explicit "AGENTS.md pack listing is exhaustive" requirement naming the previously unlisted ci.md and ci-custom.md. Hand-maintained pack enumerations in docs/ carry no equivalent guard — the durable fix generates the listings from .opencode/uf/packs/ directory contents or extends the exhaustiveness requirement to docs/usage.md and docs/architecture.md.
