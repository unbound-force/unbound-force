# Agent Interface Contract: Pinkman OSS Scout

## Invocation Interface

Pinkman is invoked via the `/scout` slash command in
OpenCode. The command file delegates to the `pinkman`
agent.

### Command: `/scout`

**Modes**:

| Mode      | Syntax                            | Description                                  |
|-----------|-----------------------------------|----------------------------------------------|
| discover  | `/scout <domain-keyword>`         | Discover OSI-approved projects by keyword    |
| trend     | `/scout --trend <category>`       | Find trending projects in a category         |
| audit     | `/scout --audit [manifest-path]`  | Audit dependencies from a manifest file      |
| report    | `/scout --report <project-url>`   | Generate recommendation report for a project |

**Default mode**: `discover` (when no flag is provided).

**Default manifest path**: `go.mod` (when `--audit` is
used without a path argument).

### Input Parameters

| Parameter       | Required | Default  | Description                              |
|-----------------|----------|----------|------------------------------------------|
| query/keyword   | Yes (discover/trend) | — | Domain keyword or category        |
| manifest-path   | No (audit) | `go.mod` | Path to dependency manifest file     |
| project-url     | Yes (report) | —   | Repository URL for the target project    |

### Output Format

All outputs are structured Markdown. Discover and trend
modes produce a result list. Audit mode produces a
health report table. Report mode produces a full
recommendation document.

#### Discover/Trend Result List

```markdown
## Scouting Results: <query>
**Mode**: discover | **Date**: YYYY-MM-DD
**Sources**: github.com, pkg.go.dev
**Results**: N compatible, M incompatible, K unknown

### Compatible Projects

#### 1. <project-name>
- **URL**: <repository-url>
- **License**: <spdx-id> (OSI-approved)
- **Compatibility**: <tier> (<verdict>)
- **Language**: <primary-language>
- **Stars**: N (↑X% in 90d)
- **Releases**: N in 6mo
- **Active contributors**: N in 90d
- **Direct dependencies**: dep-a v1.2.0, dep-b v3.0.0
- **Description**: <description>

[...repeat for each project...]

### Shared Dependencies
| Dependency | Projects Using It | Versions      | Conflict? |
|------------|-------------------|---------------|-----------|
| dep-a      | proj-1, proj-3    | v1.2.0, v1.3.0 | Yes     |
| dep-b      | proj-1, proj-2    | v3.0.0        | No        |

### Incompatible Projects (for awareness)
[...flagged projects with license explanation...]
```

#### Audit Result Table

```markdown
## Dependency Audit: <manifest-path>
**Date**: YYYY-MM-DD | **Dependencies**: N total

| Dependency | Current | Latest | Update? | License Changed? | Compatibility | Risk     |
|------------|---------|--------|---------|------------------|---------------|----------|
| dep-a      | v1.2.0  | v1.3.0 | Yes     | No               | compatible    | healthy  |
| dep-b      | v2.0.0  | v3.0.0 | Yes     | Yes (MIT→GPL-3.0)| incompatible  | critical |
| dep-c      | v0.5.0  | v0.5.0 | No      | No               | warning  |

### Risk Details
- **dep-b**: License changed from MIT to GPL-3.0
  (not OSI-incompatible but copyleft). Recommend staying
  on v2.0.0 or finding alternative.
- **dep-c**: No commits in 8 months. Issue backlog
  growing.
```

#### Recommendation Report

```markdown
---
producer: pinkman
version: "1.0.0"
timestamp: "<ISO-8601>"
query: "<project-url>"
mode: "report"
---

# Adoption Recommendation: <project-name>

## License Analysis
- **License**: <spdx-id>
- **OSI Status**: Approved / Not Approved / Unknown
- **Compatibility**: <tier> (<verdict>)
- **Verdict**: <explanation>

## Community Health
- **Stars**: N (growth: X% in 90d)
- **Forks**: N
- **Contributors**: N active in 90d
- **Health Score**: <assessment>

## Maintenance Signals
- **Last commit**: <date>
- **Release cadence**: N releases in 6mo
- **Risk level**: healthy / warning / critical
- **Indicators**: [list]

## Trend Trajectory
- **Star growth**: X% in 90d
- **Release velocity**: N in 6mo
- **Contributor growth**: N new in 90d

## Dependencies
- **Direct dependencies**: N total
- [list with versions]

## Dependency Overlap
- **Shared with previously evaluated projects**: [list]
- **Version conflicts**: [list or "none"]

## Relationship to Existing Dependencies
- **Already used by Unbound Force**: [list or "none"]
- **Shared transitive dependencies**: [deferred]

## Recommendation
**Verdict**: adopt / evaluate / defer / avoid
**Reason**: <justification>
```

## Persistence Contract

### Local Storage

Reports are written to `.uf/pinkman/reports/` with the
naming convention:
`YYYY-MM-DDTHH-MM-SS-<sanitized-query>.md`

The agent creates the directory if it does not exist.

### Dewey Integration (Optional)

When Dewey is available, the agent stores a structured
summary of each scouting session via
`dewey_store_learning` with mode-specific tags and
content prefixes (updated by
opsx/pinkman-dewey-enrichment):

- **tag**: `pinkman-<mode>` (e.g., `pinkman-discover`,
  `pinkman-trend`, `pinkman-audit`, `pinkman-report`).
  Hyphen-separated because `dewey_store_learning`
  strips `/` from tag values.
- **category**: `reference`
- **information**: Mode-specific prefix followed by
  structured prose with project names, license verdicts
  with compatibility tier/verdict, key metrics, and
  query context. Prefixes: `scouting-report:`,
  `trend-report:`, `dependency-audit:`,
  `adoption-report:`.

Primary discovery path for stored learnings is
`dewey_semantic_search` (content similarity). Tags
serve as filters via
`dewey_semantic_search_filtered(has_tag: ...)`. Do NOT
use `dewey_find_by_tag` for learning discovery.

Prior to scouting, the agent queries
`dewey_semantic_search` for past evaluations of the
same domain or project URL to avoid redundant work.

## Error Handling Contract

| Error Condition              | Agent Behavior                                         |
|------------------------------|--------------------------------------------------------|
| OSI site unreachable         | Use fallback license list, note in results             |
| GitHub rate-limited          | Report partial results, note which queries failed      |
| No manifest found (audit)    | Report "no manifest detected", skip dependency listing |
| Unknown license in project   | Classify as `unknown`, exclude from compatible list    |
| Custom/non-standard license  | Classify as `manual_review`, exclude from compatible   |
| No results for query         | Report "no projects found" with search criteria        |
| Dewey unavailable            | Skip Dewey integration silently, proceed with local    |
