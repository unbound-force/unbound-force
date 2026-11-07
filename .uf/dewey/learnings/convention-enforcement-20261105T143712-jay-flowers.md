---
tag: convention-enforcement
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:37:12Z
identity: convention-enforcement-20261105T143712-jay-flowers
tier: draft
---

A .gitignore entry added in the same change that establishes a MUST-not-exist contract masks violations of that contract: the ignore line removes `git status` as a detection channel, so the path that produces the artifact needs its own executable absence assertion. Observed on PR 957 (stop-shipping-plugin-lock): `.opencode/.gitignore` gained `package-lock.json` while SC-FR-003 forbids tracking or shipping the lock, leaving the `make plugin-test` recipe guarded only by the `--no-package-lock` flag. In Make, such an assertion must be a standalone recipe line — every recipe line starts a fresh shell in the Makefile directory, so an assertion chained after `cd <dir> && ...` resolves repo-root-relative paths against <dir> (e.g., `.opencode/.opencode/package-lock.json`) and passes vacuously against a path that never exists.
