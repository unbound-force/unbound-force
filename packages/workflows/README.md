# @unbound-force/workflows

Speckit pipeline commands for strategic features and OpenSpec commands
for tactical changes. Depends on **@unbound-force/review-council** for
shared packs and review tooling.

## Install

```bash
opkg install @unbound-force/workflows
```

Or run **`uf init`** with `opkg` available — installs this package and
review-council together.

## Commands

- **Speckit** — `speckit.constitution`, `speckit.specify`, `speckit.plan`,
  `speckit.tasks`, `speckit.implement`, and related pipeline steps
- **OpenSpec** — `opsx-propose`, `opsx-explore`, `opsx-apply`, `opsx-archive`
- **constitution-check** — hero constitution alignment

## Contents

- `agents/workflows/constitution-check.md`
- `commands/workflows/speckit.*.md`, `opsx-*.md`, `constitution-check.md`
