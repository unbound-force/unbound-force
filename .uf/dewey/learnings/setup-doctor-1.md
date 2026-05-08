---
tag: setup-doctor
category: pattern
created_at: 2026-05-05T21:30:33Z
identity: setup-doctor-1
tier: draft
---

When removing a tool from uf setup and uf doctor, the mxf removal pattern shows the clean approach: remove the installXxx function, remove the coreTools entry, renumber all subsequent steps (both comments and progress strings), remove dedicated test functions, and clean up LookPath stub entries in unrelated tests that included the tool for environment simulation. The hero availability check (agent file detection) is a separate concern from the binary distribution check and should be left intact. A stale comment referencing the removed tool in a different test was caught during code review — always grep for the tool name across all test files, not just the dedicated test functions.
