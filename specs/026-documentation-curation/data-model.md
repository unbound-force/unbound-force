# Data Model: Documentation Curation

**Branch**: `026-documentation-curation` | **Date**: 2026-04-11

## Overview

This feature introduces no new data models, schemas, or
persistent state. All artifacts are Markdown agent files
and their scaffold asset copies. The Curator's outputs
are:

1. **Review findings** — standard Divisor output format
   (text, not persisted beyond the review session)
2. **GitHub issues** — created in `unbound-force/website`
   via `gh issue create` (persisted by GitHub, not by
   this project)

## Agent File Structure

### divisor-curator.md

```yaml
# Frontmatter
description: string  # One-line role summary
mode: "subagent"     # Always subagent for Divisor agents
model: string        # Model identifier
temperature: 0.2     # Slightly higher than standard 0.1
tools:
  read: true         # Read PR diff and project files
  write: false       # No file creation
  edit: false        # No file modification
  bash: true         # Exception: gh CLI operations only
  webfetch: false    # No web access
```

```markdown
# Sections
- Role declaration with exclusive domain
- Bash Access Restriction (unique to Curator)
- Step 0: Prior Learnings (Dewey integration)
- Source Documents
- Code Review Mode
  - Review Scope
  - Audit Checklist
    1. Documentation Gap Detection
    2. Blog Opportunity Identification
    3. Tutorial Opportunity Identification
    4. Duplicate Issue Check
- Spec Review Mode (minimal — documentation
  completeness is primarily a code review concern)
- Output Format
- Decision Criteria
- Out of Scope
```

## GitHub Issue Templates

The Curator files issues with these structures:

### Documentation Issue (label: `docs`)

```markdown
Title: docs: <brief description of what changed>
Body:
  ## What Changed
  <description of the code change>

  ## Why Documentation Matters
  <user impact of the undocumented change>

  ## Suggested Updates
  <specific pages/sections that need updating>

  ## References
  - PR: <link to the triggering PR>
  - Spec: <link to the relevant spec, if any>
```

### Blog Issue (label: `blog`)

```markdown
Title: blog: <suggested topic>
Body:
  ## Topic
  <one-line topic description>

  ## Angle
  <suggested narrative angle>

  ## Key Points
  - <point 1>
  - <point 2>
  - <point 3>

  ## References
  - PR: <link to the triggering PR>
  - Spec: <link to the relevant spec, if any>
```

### Tutorial Issue (label: `tutorial`)

```markdown
Title: tutorial: <suggested topic>
Body:
  ## Topic
  <one-line topic description>

  ## Target Audience
  <who should read this tutorial>

  ## Suggested Structure
  1. <step 1>
  2. <step 2>
  3. <step 3>

  ## Prerequisites
  <what the reader needs before starting>

  ## References
  - PR: <link to the triggering PR>
  - Spec: <link to the relevant spec, if any>
```

## Finding Severity Mapping

| Finding Type | Severity | Blocking | Agent |
|-------------|----------|----------|-------|
| Missing AGENTS.md update | MEDIUM | Yes | Guard |
| Missing website docs issue | HIGH | Yes | Curator |
| Missing blog issue (significant change) | MEDIUM | Yes | Curator |
| Missing tutorial issue (workflow change) | MEDIUM | Yes | Curator |
| Internal-only change (no docs needed) | N/A | N/A | Both skip |
| Docs already updated + issue filed | N/A | N/A | Both pass |

## File Inventory

### New Files (1)

| File | Type | Owner | Lines (est.) |
|------|------|-------|-------------|
| `.opencode/agents/divisor-curator.md` | Agent | User | ~250 |

### Modified Files (3)

| File | Change | Owner |
|------|--------|-------|
| `.opencode/agents/divisor-guard.md` | +Documentation Completeness checklist | User |
| `.opencode/command/review-council.md` | +Curator row in reference table | Tool |
| `internal/scaffold/scaffold_test.go` | +1 entry in `expectedAssetPaths` | N/A |

### Scaffold Asset Copies (3)

| Live File | Asset Copy |
|-----------|-----------|
| `.opencode/agents/divisor-curator.md` | `internal/scaffold/assets/opencode/agents/divisor-curator.md` |
| `.opencode/agents/divisor-guard.md` | `internal/scaffold/assets/opencode/agents/divisor-guard.md` |
| `.opencode/command/review-council.md` | `internal/scaffold/assets/opencode/command/review-council.md` |

### Documentation Updates (1)

| File | Change |
|------|--------|
| `AGENTS.md` | +Curator in Heroes table, +Recent Changes entry, +Project Structure update |
