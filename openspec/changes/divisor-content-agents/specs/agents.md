## ADDED Requirements

### Requirement: Divisor Scribe Agent

The project MUST include a `divisor-scribe.md` agent
file that serves as a technical documentation specialist.
The agent MUST:

- Use subagent mode with write and edit access enabled
- Use low temperature (0.1) for precision
- Include Step 0: Prior Learnings via Dewey
- Define documentation workflows for READMEs, AGENTS.md,
  spec descriptions, CLI help text, and API docs
- Produce output optimized for developer audiences
- Include an Out of Scope section excluding review,
  blog, and PR/comms work

#### Scenario: Engineer requests documentation

- **GIVEN** the divisor-scribe agent is available
- **WHEN** an engineer invokes it to document a feature
- **THEN** the agent produces technically accurate
  Markdown documentation with consistent structure,
  appropriate audience level, and cross-references to
  related specs or code

#### Scenario: Dewey provides prior documentation patterns

- **GIVEN** Dewey contains learnings about documentation
  style for this project
- **WHEN** the scribe agent queries prior learnings
- **THEN** the agent applies discovered patterns
  (heading style, terminology, depth) to new content

---

### Requirement: Divisor Herald Agent

The project MUST include a `divisor-herald.md` agent
file that serves as a blog and announcement writer.
The agent MUST:

- Use subagent mode with write and edit access enabled
- Use moderate temperature (0.4) for balanced creativity
- Include Step 0: Prior Learnings via Dewey
- Define workflows for release notes, blog posts,
  feature announcements, and changelog entries
- Produce output that is technically accurate but
  accessible to non-developer audiences
- Include an Out of Scope section excluding technical
  docs, review, and PR/comms work

#### Scenario: Engineer requests a release announcement

- **GIVEN** the divisor-herald agent is available
- **WHEN** an engineer invokes it with a feature spec
  or changelog
- **THEN** the agent produces a blog-style announcement
  with narrative flow, key highlights, and audience-
  appropriate language

---

### Requirement: Divisor Envoy Agent

The project MUST include a `divisor-envoy.md` agent
file that serves as a public relations and
communications specialist. The agent MUST:

- Use subagent mode with write and edit access enabled
- Use moderate-high temperature (0.5) for creative
  flexibility
- Include Step 0: Prior Learnings via Dewey
- Define workflows for press releases, social media
  content, community updates, and partnership comms
- Maintain consistent brand voice across outputs
- Include an Out of Scope section excluding technical
  docs, blog posts, and code review work

#### Scenario: Engineer requests social media content

- **GIVEN** the divisor-envoy agent is available
- **WHEN** an engineer invokes it with a feature or
  milestone description
- **THEN** the agent produces platform-appropriate
  social content with consistent brand voice and key
  messages

---

### Requirement: Content Convention Pack

The project MUST include a `content.md` convention pack
at `.opencode/unbound/packs/content.md` that defines
writing standards for all content agents. The pack MUST:

- Use YAML frontmatter with `pack_id: content` and
  `language: Any`
- Include sections for Voice & Brand (shared), Technical
  Documentation (Scribe), Blog & Announcements (Herald),
  Public Relations (Envoy), and Fact-Checking (shared)
- Use the same rule identifier format as existing packs
  (section prefix + number, e.g., VB-001, TD-001)
- Follow RFC 2119 severity language (MUST/SHOULD/MAY)

The project MUST also include a `content-custom.md` stub
for user-owned project-specific content rules.

#### Scenario: Content agent loads pack

- **GIVEN** the `content.md` convention pack exists
- **WHEN** a content agent (Scribe, Herald, or Envoy)
  is invoked
- **THEN** the agent references the pack's rules
  relevant to its domain section

#### Scenario: Pack is always deployed

- **GIVEN** the target project uses any language
- **WHEN** an engineer runs `uf init`
- **THEN** the `content.md` and `content-custom.md`
  packs are deployed alongside `default.md` and
  `severity.md` (language-agnostic, always deployed)

---

### Requirement: Scaffold Asset Registration

Each content agent file and convention pack file MUST
have a corresponding scaffold asset copy. The scaffold
drift detection tests MUST verify these copies match
their canonical sources.

#### Scenario: uf init deploys content agents and packs

- **GIVEN** the scaffold engine includes the 5 new
  files (3 agents + 2 packs) as embedded assets
- **WHEN** an engineer runs `uf init`
- **THEN** the 3 agent files are deployed to
  `.opencode/agents/` and the 2 pack files are
  deployed to `.opencode/unbound/packs/` in the
  target directory

#### Scenario: Drift detection catches mismatches

- **GIVEN** a canonical agent or pack file is modified
- **WHEN** the scaffold asset copy is not updated
- **THEN** the drift detection test fails

## MODIFIED Requirements

### Requirement: Scaffold Asset Count

The scaffold engine's expected asset path count
MUST increase to include the 5 new files (3 agents
+ 2 packs).

Previously: Asset count reflected 5 Divisor review
agents and 7 convention packs.

### Requirement: Pack Deployment Logic

`shouldDeployPack()` in `internal/scaffold/scaffold.go`
MUST be updated to always deploy `content` and
`content-custom` packs (language-agnostic, alongside
`default`, `default-custom`, and `severity`).

Previously: Only `default`, `default-custom`, and
`severity` were always-deploy packs.

## REMOVED Requirements

None.
