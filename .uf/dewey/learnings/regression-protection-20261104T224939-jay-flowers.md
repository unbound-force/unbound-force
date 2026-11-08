---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:49:39Z
identity: regression-protection-20261104T224939-jay-flowers
tier: draft
---

When a change stops shipping an artifact AND adds a gitignore line for that artifact, any one-time manual \"no artifact left behind\" verification must become a self-verifying guard inside the recipe that produces the artifact (e.g. `test ! -f <artifact>` chained after the producing command). The gitignore line masks exactly the regression mode the manual check was meant to catch: if the producing command starts writing the artifact again, git status stays clean and no CI check fails. Observed in unbound-forc

> Note: payload truncated at store time (2026-11-04T22:49); the
> full text is unrecoverable — the Dewey graph page, git history
> (single commit 24b20b55), and dispatch artifacts all carry the
> same cut. Treat the surviving fragment as the record.
