---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-02T19:45:31Z
identity: openspec-20261102T194531-jay-flowers
tier: draft
---

When drafting an OpenSpec delta, survey the capability's UNARCHIVED sibling changes (openspec/changes/*/specs/<capability>*/) before assigning requirement numbers or declaring New vs Modified Capabilities — archived openspec/specs/ alone is not the full numbering state. Concretely, in unbound-force/unbound-force (observed 2026-11-02 reviewing fix-uf-finale-ci-poll-gh-schema): fetch-pr-review-context had FR-001..FR-010 in its origin delta plus FR-013..FR-016 from extend-fetch-pr-review-context, so a fresh delta numbering itself FR-001 collided; and a capability name claimed by any prior unarchived change counts as existing (Modified), while a never-claimed name with an ADDED-only delta is New. Additionally, tool-level requirements for a plugin's tools belong to that plugin's surface capability (e.g. plugin-tools in finale-token-reduction governs every uf-finale tool); opening a per-tool capability duplicates governance — reconcile via the supersedes annotation in the earlier delta (fix-tui-git-output-leak precedent) or extend the existing capability instead.
