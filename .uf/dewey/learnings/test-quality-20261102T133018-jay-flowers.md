---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-02T13:30:18Z
identity: test-quality-20261102T133018-jay-flowers
tier: draft
---

CLI-mode test coverage that exercises only rejection paths (flag conflicts, help text) leaves the flag→Options wiring untested: scaffold-level tests construct the mode field directly, so a dropped wiring line in the command layer ships a full-behavior regression under the flag with the entire suite green. Observed reviewing unbound-force PR #918 (2026-11-02): `uf init --overrides` had a conflict matrix, cobra-surface rejection tests, and help-text assertions at the cmd layer plus exhaustive Options.OverridesOnly behavior coverage in package scaffold — but no cmd-layer happy-path invocation, so removing `OverridesOnly: p.overridesOnly` from runInit's Options literal would make `--overrides` silently run a full scaffold, uncaught by any test. Rule: every new CLI mode needs at least one happy-path test through the command surface (runXxx or cobra Execute) asserting the mode's defining observable contract — for --overrides, the exact entry set — not just its rejection paths. Such tests are cheap whenever the mode early-exits before sub-tool or plugin activation, and the assertion machinery usually already exists at the scaffold layer to reuse.
