---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:09:47Z
identity: supply-chain-20261105T160947-jay-flowers
tier: draft
---

When a repository stops tracking a package-lock.json to end lock contention between competing writers (e.g., a scaffold tool and opencode both writing .opencode/), lock-based scanners (OSV-Scanner, Dependabot) lose transitive-tree visibility while package.json direct pins remain scannable. A CI step that generates an ephemeral lock (`npm install --package-lock-only --ignore-scripts`), scans it, and discards it restores transitive visibility without reintroducing a tracked lock. Gotcha: in node 22 CI environments where the bundled npm 10.9.9 crashes in Arborist on the vitest 5 peer graph, the ephemeral-lock generation must run through a pinned npm 11 (`npx -y npm@11.6.2`), because --package-lock-only performs the same ideal-tree construction that triggers the crash.
