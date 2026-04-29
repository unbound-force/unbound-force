# @unbound-force/review-council

AI code review council — 9 reviewer personas audit your code
in parallel for security, architecture, testing, operations,
intent drift, and documentation completeness.

## Install

```bash
opkg install @unbound-force/review-council
```

Or add to your project's `openpackage.yml`:

```yaml
dependencies:
- name: "@unbound-force/review-council"
  version: ^0.1.0
```

## What You Get

| Persona | Agent | Focus |
|:---|:---|:---|
| The Guard | `divisor-guard` | Intent drift, zero-waste, constitution alignment |
| The Architect | `divisor-architect` | Structure, conventions, DRY, patterns |
| The Adversary | `divisor-adversary` | Secrets, CVEs, error handling, injection safety |
| The Operator | `divisor-sre` | Deployment, performance, dependencies, observability |
| The Tester | `divisor-testing` | Test architecture, coverage, assertions, isolation |
| The Curator | `divisor-curator` | Documentation gaps, blog/tutorial opportunities |
| The Scribe | `divisor-scribe` | Technical documentation (READMEs, API docs) |
| The Herald | `divisor-herald` | Blog posts, release notes, announcements |
| The Envoy | `divisor-envoy` | Press releases, social media, community updates |

Plus 2 commands (`/review-council`, `/review-pr`) and 3 convention
packs (`severity`, `default`, `default-custom`).

## License

Apache-2.0
