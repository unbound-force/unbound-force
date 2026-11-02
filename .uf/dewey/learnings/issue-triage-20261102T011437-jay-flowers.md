---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:37Z
identity: issue-triage-20261102T011437-jay-flowers
tier: draft
---

When triaging a documentation-gate record for a convention-pack change, the pack file itself is the primary verification artifact: its frontmatter version corroborates the claimed version bump, its header-comment classification (determinism.md declares itself a "repo-local governance pack (not a scaffold asset)") is the in-repo evidence justifying a website-exemption disposition, and its rule text (severity mapping, named test layers) cross-checks verbatim against the issue's CHANGELOG-disposition claims. Verified triaging unbound-force/unbound-force#781 (seam-verification gate record): the DE-007 [MUST]/CRITICAL mapping and three seam-verification layers (registration surface, documented output shape, round-trips) match determinism.md exactly, the AGENTS.md no-change disposition matches the pack's existing listing in AGENTS.md's Convention Packs section, and the website exemption cites precedent #762. Successive governance-pack bumps (DE-007 at v1.2.0, DE-008 later) generate structurally similar docs-gate issues that keyword-match each other in duplicate searches; the referenced OpenSpec change or issue number is the discriminator.
