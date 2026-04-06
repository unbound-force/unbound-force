## Why

The Unbound Force agent ecosystem has strong coverage for
code-facing workflows (development, review, testing,
management, product ownership) but no agents for
content-facing workflows. Technical documentation, blog
posts, and public communications are created ad hoc
without the structured quality gates, persona
consistency, or institutional memory that the Divisor
review agents bring to code review.

Adding three content-focused agents — Scribe (technical
docs), Herald (blog/announcements), and Envoy
(PR/comms) — extends the Divisor pattern to content
creation, giving the team consistent voice, quality
standards, and Dewey-integrated knowledge retrieval for
written output.

## What Changes

### New Capabilities

- `divisor-scribe`: Technical documentation agent.
  Creates and reviews READMEs, AGENTS.md, spec
  descriptions, CLI help text, and API documentation.
  Optimizes for accuracy, completeness, and developer
  audience clarity.
- `divisor-herald`: Blog and announcement agent.
  Creates release notes, blog posts, feature
  announcements, and changelog entries. Optimizes for
  narrative flow, audience engagement, and technical
  accuracy without jargon.
- `divisor-envoy`: Public relations and communications
  agent. Creates press releases, social media content,
  community updates, and partnership communications.
  Optimizes for brand voice consistency, key message
  clarity, and audience-appropriate tone.
- `content.md` convention pack: Language-agnostic
  writing standards shared by all content agents.
  Sections for Voice & Brand, Technical Documentation,
  Blog & Announcements, Public Relations, and
  Fact-Checking. Always deployed (like `default.md`
  and `severity.md`).
- `content-custom.md` convention pack stub:
  User-owned file for project-specific content rules.

### Modified Capabilities

- Scaffold engine: Adds 5 new embedded asset files
  (3 agents + 2 packs), updates `shouldDeployPack()`
  to always deploy `content`/`content-custom`, updates
  file count in tests.

### Removed Capabilities

None.

## Impact

- 5 new files:
  `.opencode/agents/divisor-scribe.md`,
  `.opencode/agents/divisor-herald.md`,
  `.opencode/agents/divisor-envoy.md`,
  `.opencode/unbound/packs/content.md`,
  `.opencode/unbound/packs/content-custom.md`
- 5 new scaffold assets (mirrors of the above)
- Updated `internal/scaffold/scaffold.go` —
  `shouldDeployPack()` updated to always deploy
  `content`/`content-custom` (language-agnostic)
- Updated `internal/scaffold/scaffold_test.go` —
  `expectedAssetPaths` count increases, pack
  deployment tests
- Updated `AGENTS.md` — agent list and Recent Changes
- Minimal Go logic change: 1 line in
  `shouldDeployPack()` to add the content pack to the
  always-deploy list

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The three agents communicate through well-defined
artifacts (Markdown documents, structured output). They
operate independently — each produces self-describing
content without runtime coupling to other agents.

### II. Composability First

**Assessment**: PASS

Each agent is independently usable. An engineer can
invoke `divisor-scribe` alone for documentation without
needing Herald or Envoy. No mandatory dependencies
between content agents or between content agents and
review agents.

### III. Observable Quality

**Assessment**: N/A

Content agents produce human-readable Markdown, not
machine-parseable JSON artifacts. This is appropriate
for their domain (prose content) and does not violate
the principle — Observable Quality applies to hero
artifacts exchanged between heroes, not to all agent
output.

### IV. Testability

**Assessment**: PASS

Each agent is testable in isolation via the scaffold
drift detection tests (embedded assets match canonical
sources) and can be invoked independently. No external
services required.
