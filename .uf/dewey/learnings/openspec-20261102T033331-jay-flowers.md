---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:33:31Z
identity: openspec-20261102T033331-jay-flowers
tier: draft
---

In OpenSpec stacks with multiple unarchived deltas, a requirement that normatively cites another requirement (e.g., "persist host-run provenance per RD-FR-008") inherits that requirement's supersession chain: the citation's meaning depends on which delta last modified the cited text. Archive-ordering declarations must therefore cover not only the declared predecessors of MODIFIED entries but also every unarchived delta that modified any cited requirement — otherwise an archive order exists that materializes contradictory normative text with no automated detection (openspec validates delta structure, not cross-requirement semantic consistency). Found in the host-fallback-provenance spec review (Nov 2026): RD-FR-037's variant-or-null citation of RD-FR-008 was consistent only with accept-variantless-parent-model's MODIFIED RD-FR-008, not the fanout original ("both MUST be non-null"), so the change needed to declare ordering after accept-variantless-parent-model in addition to advisor-host-model-fallback.
