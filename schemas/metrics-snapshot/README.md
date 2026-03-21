# Metrics Snapshot Schema

The metrics snapshot payload captures Mx F's point-in-time
collection of all computed engineering metrics: velocity, cycle
time, lead time, defect rate, and health indicators.

## Producer

**Mx F** — the Manager hero.

## Consumers

- **Muti-Mind** — uses metrics for backlog prioritization
- **Cobalt-Crush** — monitors quality trends
- **The Divisor** — tracks review iteration trends

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 collection timestamp |
| `velocity` | number | Sprint velocity |
| `cycle_time` | object | Cycle time stats (avg, median, p90, p99) |
| `lead_time` | number | Lead time in days |
| `defect_rate` | number | Defect rate (0.0-1.0) |
| `review_iterations` | number | Average review iterations |
| `ci_pass_rate` | number | CI pass rate (0.0-1.0) |
| `backlog_health` | object | Backlog health (total, ready, stale) |
| `flow_efficiency` | number | Flow efficiency ratio |
| `sources_collected` | array | Data sources used |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `health_indicators` | array | Traffic-light health assessments |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
