---
tag: git-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:45:24Z
identity: git-tooling-20261105T144524-jay-flowers
tier: draft
---

git ignores `--force-if-includes` when `--force-with-lease` carries an explicit expected value — the flag only protects the ref-only lease form (`--force-with-lease=<ref>`). So a force-push tool needs three lease modes: (1) explicit `expected_sha` verified as a local commit via `rev-parse --verify <sha>^{commit}` (fail EXPECTED_SHA_UNKNOWN before pushing), which callers must record BEFORE any fetch that could move the remote-tracking ref; (2) default ref-only lease plus `--force-if-includes`, which rejects   [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
