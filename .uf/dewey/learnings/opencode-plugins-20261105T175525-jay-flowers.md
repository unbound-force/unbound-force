---
tag: opencode-plugins
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:55:25Z
identity: opencode-plugins-20261105T175525-jay-flowers
tier: draft
---

In this repo, the intermittent appearance of `.opencode/package-lock.json` after `make plugin-test` (despite the recipe's `--no-package-lock` install and a no-lock assertion) is caused by CONCURRENT OPENCODE HOST PROCESSES re-installing the live `.opencode/` plugins with their own npm invocation (no `--no-package-lock`) whenever `node_modules` churns (e.g. `rm -rf node_modules`) or plugin files change. Verification recipe that isolated it: (1) replace the lock path with a directory and run the vitest suite — no EISDIR/ENOTDIR errors proves the suite never writes it; (2) run the install command alone in a scratch dir with the exact manifest — no lock proves npm honors the flag; (3) stat the lock mtimes AFTER a clean install — mtimes seconds AFTER the command finished with no concurrent command of your own proves an external writer; (4) `ps aux` shows the opencode hosts. npm itself never wrote a visible lock in any controlled run (up-to-date path, fresh-install path, four scratch-manifest variants). Disposition: the recipe and its no-lock assertion are correct; CI (no host, clean checkout) is deterministic; locally the assertion can race with hosts re-installing plugins — remove the stray lock (`rm -f .opencode/package-lock.json`), do NOT remove node_modules (that retriggers the host), and re-run. The lock is gitignored so the repo is never contaminated. Key tell: the host's lock tree differs from the recipe's (different byte size and package count than `npm install --package-lock-only` output).
