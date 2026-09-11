## Why

When agents generate scaffolded issues for non-org-config
repositories (e.g., unbound-force/gaze#233), the issue
bodies contain org-config-specific language like
"organization-configuration repository", "Peribolos",
and "safe-settings". The text is LLM-generated, not
hardcoded in templates. Without guardrails, every
issue-generating component can produce content that is
inappropriate for the target repository's actual type.

This is tracked as GitHub issue #593. Related family of
hardcoded org-specific content bugs: #526, #527, #529,
#530.

## What Changes

Add Content Portability rules at two levels:

1. **Systemic**: New CP-001, CP-002, CP-003 rules in the
   default convention pack (`.opencode/uf/packs/default.md`)
   that all agents load. These rules prohibit
   org-config-specific terminology in LLM-generated issue
   content unless the target repository actually uses those
   tools.

2. **Targeted**: Content portability guardrails in four
   issue-generating components:
   - Triage child issue creation (`uf.triage-issue.md`)
   - Tasks-to-issues command (`speckit.taskstoissues.md`)
   - `uf.init` taskstoissues template (`uf.init.md`)
   - Divisor-curator agent (`divisor-curator.md`)

3. **Tests**: Four regression tests verifying the guardrails
   remain in place.

## Capabilities

### New Capabilities
- CP-001: Generated content MUST use language appropriate
  to the target repository's language and project type
- CP-002: Generated content MUST NOT reference tools
  specific to a single repository type unless the target
  repository actually uses them
- CP-003: Checklists and verification plans SHOULD describe
  checks in terms of the target repository's actual
  technology stack

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **Files**: `.opencode/uf/packs/default.md`,
  `.opencode/agents/divisor-curator.md`,
  `.opencode/commands/speckit.taskstoissues.md`,
  `.opencode/commands/uf.init.md`,
  `.opencode/commands/uf.triage-issue.md`,
  `internal/scaffold/assets/` (canonical copies),
  `internal/scaffold/scaffold_test.go`
- **Behaviors**: Issue-generating agents will use
  repository-appropriate language instead of assuming
  org-config context
- **Spec alignment**: Closes the gap identified in #593
  and the related family (#526, #527, #529, #530)

## Status

Retroactive. Implementation preceded this spec artifact.
Documented per AGENTS.md emergency hotfix exemption
(retroactively documented).
