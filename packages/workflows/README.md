# @unbound-force/workflows

Spec-driven development workflows for AI-assisted software
engineering. Two tiers:

- **Speckit** (strategic): multi-phase pipeline for features with
  several user stories
- **OpenSpec** (tactical): lightweight workflow for bug fixes and
  small changes

## Install

```bash
opkg install @unbound-force/workflows
```

Or add to your project's `openpackage.yml`:

```yaml
dependencies:
- name: "@unbound-force/workflows"
  version: ^0.1.0
```

This also pulls in `@unbound-force/review-council` as a dependency.

## Speckit Commands

| Command | Purpose |
|:---|:---|
| `/speckit.constitution` | Create or update project constitution |
| `/speckit.specify` | Create feature specification |
| `/speckit.clarify` | Reduce spec ambiguity |
| `/speckit.plan` | Generate implementation plan |
| `/speckit.tasks` | Break plan into ordered tasks |
| `/speckit.analyze` | Cross-artifact consistency check |
| `/speckit.checklist` | Quality validation |
| `/speckit.implement` | Execute tasks |
| `/speckit.taskstoissues` | Convert tasks to GitHub Issues |
| `/speckit.testreview` | Test review pass |

## OpenSpec Commands

| Command | Purpose |
|:---|:---|
| `/opsx-propose` | Create change proposal with plan and tasks |
| `/opsx-explore` | Think through ideas (read-only exploration) |
| `/opsx-apply` | Implement tasks from a change |
| `/opsx-archive` | Archive a completed change |

## Other

| Command | Purpose |
|:---|:---|
| `/constitution-check` | Hero vs org constitution alignment |

## License

Apache-2.0
