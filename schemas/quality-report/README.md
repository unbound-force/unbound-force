# Quality Report Schema

The quality report payload captures Gaze's static analysis output:
per-function CRAP scores, coverage data, and actionable
recommendations.

## Producer

**Gaze** — the Tester hero.

## Consumers

- **Mx F** — uses quality data for metrics dashboards and coaching
- **Muti-Mind** — uses quality data for acceptance decisions
- **Cobalt-Crush** — uses quality feedback to improve code

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `summary` | object | Aggregate quality scores (total_functions, avg_coverage, avg_crap, crap_load) |
| `functions` | array | Per-function metrics (name, crap_score, complexity, coverage, contract_coverage, classification) |
| `coverage` | object | Aggregate coverage data (total_lines, covered_lines, percentage) |
| `recommendations` | array | Improvement suggestions (priority, description, target) |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
