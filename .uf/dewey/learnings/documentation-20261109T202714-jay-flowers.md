---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-09T20:27:14Z
identity: documentation-20261109T202714-jay-flowers
tier: draft
---

Behavior documentation that cites a tracking issue in present tense ("Issue N tracks moving X into a tool") goes silently stale when a later PR closes N — even though every behavioral claim in the docs stays accurate. Only the provenance sentence's tense is the casualty, so ordinary accuracy checks pass. When reviewing a tool-extraction PR in unbound-force, grep docs/ for references to every issue the PR closes (originating_issue in .openspec.yaml, "Closes #N" in the PR body) and flag the anchor sentences for a tense/status refresh. Observed on PR #1002: docs/usage.md:187-189 anchored the /uf.address-feedback checkout section to #965 ("tracks moving the checkout into a tested tool"), which that same PR closes.
