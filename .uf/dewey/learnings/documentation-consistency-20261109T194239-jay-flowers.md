---
tag: documentation-consistency
author: jay-flowers
category: gotcha
created_at: 2026-11-09T19:42:39Z
identity: documentation-consistency-20261109T194239-jay-flowers
tier: draft
---

A documentation PR that corrects a numeric claim (persona/agent counts, thresholds, limits) must sweep every occurrence of the stale number in the files it touches, not only the lines it set out to rewrite: grep each touched file for the old and new numerals in both digit and word form before merge. Observed on unbound-force PR #1000 (docs-catchup): the roster story was corrected to "9 delivered by uf init, 11 in this repository" across QUICKSTART, usage, and architecture, and a bullet was added to heroes.md's Current Capabilities list, but the same section's intro paragraph still said "five canonical roles ship by default" — leaving the heroes page self-contradicting on exactly the claim two of the PR's closing issues targeted. Section intros survive list edits because corrections land where the new content goes, not where the old claim lives; the cheap countermeasure is a pre-merge grep of every touched file for the superseded numeral.
