---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:02:02Z
identity: documentation-pipeline-20261107T160202-jay-flowers
tier: draft
---

A docs-gate tracking issue filed without its repo's tracking label passes content verification but is invisible to label-filtered discovery, presenting as a missing gate artifact. Verified on PR 948 (unbound-force/unbound-force): the website half of the docs gate, unbound-force/website#379, is OPEN with full page coverage (names website#310 and #317 as the pages to update) but carries zero labels, while every other open docs issue in unbound-force/website carries `docs`. `gh issue list --repo unbound-force/website --label docs --state open` did not return it, so the audit initially read the website gate half as unfiled. Gate audits must view each claimed issue number directly (`gh issue view N --repo <repo> --json state,labels,body`) rather than trusting label searches alone, and classify a label-absent-but-covering issue as a one-command labeling fix (`gh issue edit N --repo <repo> --add-label docs`) rather than a missing gate. This refines the per-repo/per-label verification rule: label search can produce a false "missing issue" signal, symmetric to the false "satisfied gate" signal a wrong-labeled issue produces.
