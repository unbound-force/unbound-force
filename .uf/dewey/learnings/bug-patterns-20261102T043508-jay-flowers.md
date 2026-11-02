---
tag: bug-patterns
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:35:08Z
identity: bug-patterns-20261102T043508-jay-flowers
tier: draft
---

Cobra CLI triage signature: a bug report pairing "Error: &lt;message&gt;" with full command usage output, for a correctly invoked command, almost always means a runtime error returned from RunE with SilenceUsage unset on the root command — Cobra prints usage for every returned error unless silenced. Fix pattern: set SilenceUsage on the root command, keep usage output on the flag-parse/argument-error path, and lock the behavior in with a regression test asserting stderr on an injected runtime failure contains the error but not "Usage:", while --help and invalid-flag invocations still print usage. Verified on unbound-force issue #577 (uf doctor, v0.17.0): the website troubleshooting docs direct users to run uf doctor in broken environments, so the misleading usage output lands exactly on the audience doctor serves. Applies to every Unbound Force Go CLI (uf, gaze, dewey, replicator), all built on spf13/cobra.
