# Quickstart: Specification Framework

**Spec**: 003-specification-framework
**Date**: 2026-03-08

This guide covers installing the unified specification
framework (Speckit + OpenSpec) in a new or existing
repository.

## Prerequisites

- Git repository (any project, not just Unbound Force heroes)
- Node.js >= 20.19.0 (for OpenSpec CLI)

## Step 1: Install the Unbound CLI

```bash
# Homebrew (recommended)
brew install unbound-force/tap/unbound

# Or via Go
go install github.com/unbound-force/unbound-force/\
cmd/unbound@latest
```

## Step 2: Scaffold Your Repository

From your repository root:

```bash
unbound init
```

This creates:

```text
your-repo/
+-- .specify/
|   +-- templates/          # 6 Speckit templates
|   +-- scripts/bash/       # 5 Speckit scripts
+-- .opencode/
|   +-- command/            # 10 OpenCode commands
|   +-- agents/             # Constitution check agent
+-- openspec/
    +-- specs/              # OpenSpec behavior contracts
    +-- changes/            # Active tactical changes
    +-- schemas/
    |   +-- unbound-force/  # Custom schema
    +-- config.yaml         # OpenSpec configuration
```

Each file includes a version marker:
`<!-- scaffolded by unbound v1.0.0 -->`

## Step 3: Install OpenSpec CLI

```bash
npm install -g @fission-ai/openspec@latest
```

Then initialize OpenSpec integration with OpenCode:

```bash
openspec init --tools opencode --profile core
```

This adds OpenSpec skills and commands to `.opencode/`.

## Step 4: Configure Your Project

Create `.specify/config.yaml` with your project details:

```yaml
language: go
framework: ""
build_command: "go build ./..."
test_command: "go test ./..."
integration_patterns: "RESTful APIs"
project_type: cli
```

Edit `openspec/config.yaml` to verify the constitution
reference is correct:

```yaml
schema: unbound-force

context: |
  This project follows the Unbound Force org constitution
  v1.0.0. Three core principles govern all work:
  I. Autonomous Collaboration
  II. Composability First
  III. Observable Quality

rules:
  proposal:
    - MUST include Constitution Alignment section
```

## Step 5: Create Your First Constitution

If your repo doesn't have a constitution yet:

```
/speckit.constitution
```

This creates `.specify/memory/constitution.md` following the
org constitution template.

## Step 6: Choose Your Workflow

### For Strategic Work (Speckit)

Use when: new features with 3+ user stories, architectural
changes, cross-repo impact, constitution changes.

```text
/speckit.specify <description>  # Create spec
/speckit.clarify                # Reduce ambiguity
/speckit.plan                   # Generate plan
/speckit.tasks                  # Generate tasks
/speckit.analyze                # Check consistency
/speckit.checklist              # Quality validation
/speckit.implement              # Execute tasks
```

### For Tactical Work (OpenSpec)

Use when: bug fixes, small enhancements (<3 stories),
maintenance tasks, single-repo changes.

```text
/opsx:propose <description>    # Create proposal + plan
/opsx:apply                    # Implement tasks
/opsx:archive                  # Archive completed change
```

## Upgrading

When a new version is released:

```bash
brew upgrade unbound
# or: go install ...@latest

unbound init
```

Re-running `unbound init` on an existing repo:
- **User-owned files** (templates, scripts, agents, config):
  Skipped if they already exist
- **Tool-owned files** (speckit commands, OpenSpec schema):
  Updated if content has changed
- Use `unbound init --force` to overwrite everything

## Decision Guide

| Your Work | Tool | Command |
|-----------|------|---------|
| New hero architecture | Speckit | `/speckit.specify` |
| Cross-repo spec | Speckit | `/speckit.specify` |
| Constitution change | Speckit | `/speckit.constitution` |
| Bug fix | OpenSpec | `/opsx:propose` |
| Small enhancement | OpenSpec | `/opsx:propose` |
| Maintenance task | OpenSpec | `/opsx:propose` |
| Refactoring | OpenSpec | `/opsx:propose` |
| Unsure? | OpenSpec | Start tactical, escalate if needed |
