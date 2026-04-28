# Quickstart: Pinkman OSS Scout

## Prerequisites

- OpenCode installed and configured
- `uf init` has been run in the target repository
  (deploys `pinkman.md` agent and `scout.md` command)

## Basic Usage

### Discover OSS Projects

Find OSI-approved open source projects for a domain:

```text
/scout static analysis Go
```

Returns a curated list of projects with license
verdicts, trend indicators, direct dependencies, and
shared dependency overlap.

### Track Trending Projects

Find trending projects in a category:

```text
/scout --trend MCP servers
```

Returns projects ranked by trend strength (star growth,
release velocity, contributor activity).

### Audit Existing Dependencies

Check dependency health for the current project:

```text
/scout --audit
```

Reads `go.mod` by default. Specify a different manifest:

```text
/scout --audit package.json
```

Returns a table of dependencies with version status,
license changes, and maintenance risk levels.

### Generate Adoption Report

Get a detailed recommendation for a specific project:

```text
/scout --report https://github.com/example/project
```

Returns a structured report with license analysis,
community health, trend trajectory, and an adopt/
evaluate/defer/avoid recommendation.

## Report Storage

Scouting reports are saved to `.uf/pinkman/reports/`.
When Dewey is available, results are also stored in the
knowledge graph for cross-session search.

## Customization

The `pinkman.md` agent file is user-owned. You can
customize:
- Default data sources and search behavior
- Report format and sections
- Domain keywords and category lists
- Trend indicator thresholds

Edit `.opencode/agents/pinkman.md` to modify.
