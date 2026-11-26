---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-26T16:35:31Z
identity: test-quality-20261126T163531-jay-flowers
tier: draft
as_of: "unbound-force@56992cd6fed3f692274a5ff3aba690f32f457308"
---

Contract tests that pin multi-line shell-block inputs (e.g., a workflow's `with:` install-check commands) must assert every command line of the block, not just the first. A partial strings.Contains assertion — pinning the `test -L` symlink existence check but not the `--version` run — leaves the functional half of the contract unprotected: drift that drops or breaks the unpinned command passes silently. Loop over each command in the block (mirroring other multi-command assertions in the same file) and use exact-value pins for single-value inputs. Observed in internal/homebrew/workflow_test.go's tap_install_check assertion vs. the go_install_check exact pin in the same PR.
