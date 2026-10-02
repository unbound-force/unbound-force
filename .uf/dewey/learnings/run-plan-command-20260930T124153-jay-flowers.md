---
tag: run-plan-command
author: jay-flowers
category: gotcha
created_at: 2026-09-30T12:41:53Z
identity: run-plan-command-20260930T124153-jay-flowers
tier: draft
---

When adding a new OpenCode command to the Unbound Force project, three changes are required: (1) create the canonical file at .opencode/commands/uf.<name>.md, (2) copy it byte-identical to internal/scaffold/assets/opencode/commands/uf.<name>.md, (3) add the path to expectedAssetPaths in internal/scaffold/scaffold_test.go in alphabetical order, and (4) update the comment count (e.g., "OpenCode commands (10)" to "(11)"). Additionally, cmd/unbound-force/main_test.go TestRunInit_FreshDir has a hardcoded file count assertion (e.g., "45 files processed") that must be incremented by 1 for each new scaffold asset. The stray internal/scaffold/AGENTS.md file can appear during scaffold operations and must be removed before staging.
