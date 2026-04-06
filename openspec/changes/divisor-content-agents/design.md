## Context

The project has 5 Divisor review agents
(`divisor-adversary`, `divisor-architect`,
`divisor-guard`, `divisor-sre`, `divisor-testing`) that
follow a consistent structural pattern:

- YAML frontmatter (description, mode, model,
  temperature, tools)
- Role heading with domain description
- Step 0: Prior Learnings (Dewey integration)
- Mode-specific workflows (Code Review / Spec Review)
- Severity guidelines (reference shared severity.md)
- Out of Scope section

This change adds 3 content-creation agents that follow
the same file structure but serve a different function:
content production rather than code/spec review.

The project also has 6 convention packs (`default.md`,
`default-custom.md`, `go.md`, `go-custom.md`,
`severity.md`, `typescript.md`, `typescript-custom.md`)
deployed by the scaffold engine. Language-specific packs
are conditional; `default`, `default-custom`, and
`severity` are always deployed. The website repo has a
`markdown.md` pack with writing quality standards for
docs and blog content — this change adds a similar
`content.md` pack to the meta repo for content agents.

## Goals / Non-Goals

### Goals

- Create 3 agent files that OpenCode discovers via the
  `divisor-*.md` naming pattern
- Create 1 shared convention pack (`content.md`) with
  writing standards for all content agents
- Each agent has a distinct voice and audience focus
- Each agent references the `content.md` pack for its
  relevant sections
- Each agent integrates with Dewey for prior knowledge
  retrieval (consistent with existing agents)
- Scaffold engine deploys agents and packs via `uf init`
- Drift detection tests catch unsynchronized copies

### Non-Goals

- No review-council integration — these are invoked
  directly, not through `/review-council`
- No new hero architecture — these are agents under
  the existing Divisor namespace
- No structured JSON output schema — content agents
  produce prose, not machine-parseable artifacts

## Decisions

### D1: Agent naming follows Divisor pattern

Agents are named `divisor-scribe.md`,
`divisor-herald.md`, `divisor-envoy.md` to leverage
OpenCode's automatic agent discovery. This means they
appear in the agent list alongside review agents.

**Trade-off**: Content agents are not reviewers, but
the `divisor-` prefix groups them with the review
council. This is acceptable because the Divisor is the
team's "communication specialist" hero — both review
and content are communication functions.

### D2: Subagent mode with write access

Unlike review agents (which have `write: false`,
`edit: false`, `bash: false`), content agents need
write access to create documentation files. The
frontmatter sets:

```yaml
tools:
  write: true
  edit: true
  bash: false
```

Bash remains disabled — content agents should not
execute shell commands.

### D3: Temperature tuning per agent

- Scribe: `temperature: 0.1` — technical docs require
  precision and consistency
- Herald: `temperature: 0.4` — blog posts benefit from
  moderate creativity while maintaining accuracy
- Envoy: `temperature: 0.5` — PR comms need the most
  creative freedom for engaging language

### D4: No severity.md reference

Content agents do not produce severity-graded findings.
They produce documents. The shared `severity.md`
convention pack is not referenced — it applies only to
review agents.

### D5: One unified content convention pack

A single `content.md` pack with sections for each
content domain (Technical Documentation, Blog &
Announcements, Public Relations) plus shared sections
(Voice & Brand, Fact-Checking, Formatting). This
follows the `markdown.md` precedent from the website
repo, which combines docs and blog rules in one pack.

Each agent references the pack and focuses on its
relevant sections. The pack is always deployed
(language-agnostic, like `default.md` and `severity.md`)
because content writing rules apply in any project.

`content-custom.md` provides a user-owned stub for
project-specific content rules (same pattern as
`default-custom.md`, `go-custom.md`).

`shouldDeployPack()` in `scaffold.go` is updated to
always deploy `content` and `content-custom` alongside
`default`, `default-custom`, and `severity`.

### D6: Dewey integration follows established pattern

All 3 agents include Step 0: Prior Learnings using
`dewey_semantic_search`, identical to the existing
review agents. This ensures content agents benefit
from institutional memory (prior blog posts, docs
patterns, communication style guides).

## Risks / Trade-offs

### Risk: Naming confusion

Engineers may expect `divisor-*` agents to be review
agents. Mitigated by clear role descriptions in
frontmatter and the first paragraph of each agent file.

### Risk: Scaffold asset count growth

Adding 5 files (3 agents + 2 packs) increases the
embedded asset count. This is minimal overhead (~200
lines per agent, ~300 lines for the content pack) and
follows the same pattern as existing agents and packs.

### Trade-off: No review-council integration

Content agents are invoked directly (`/task` with
agent type) rather than through `/review-council`.
This keeps the review council focused on its core
function (code/spec review) and avoids overloading
it with content creation concerns.
