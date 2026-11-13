---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:00:05Z
identity: supply-chain-20261107T180005-jay-flowers
tier: draft
---

Removing a tracked npm lockfile weakens transitive vulnerability scanning (OSV-Scanner/Dependabot lose the lock to read), and --no-audit on the install commands silences npm's own advisory report. A working compensating control, shipped in unbound-force PR #957: generate a throwaway lock in CI with the same exact-pinned npm the local install path uses (npx -y npm@<exact> install --package-lock-only --ignore-scripts --no-audit --no-fund), scan it (osv-scanner scan source --lockfile=...), and never track or ship it. Two enforcement details make it durable: a twin-pin drift test asserting the CI workflow's npm version equals the Makefile's PLUGIN_TEST_NPM (otherwise the scanned tree diverges from the tested tree), and a post-run `test ! -e <lock>` assertion in the Make target (otherwise concurrent tooling re-creates the lock locally and trips the no-lock-shipped gate). Hero repos adopting the lockless posture must copy the scan job — the removal commit's changelog entry is the delivery vehicle for that instruction.
