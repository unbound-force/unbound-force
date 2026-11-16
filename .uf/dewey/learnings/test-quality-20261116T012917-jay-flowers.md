---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:29:17Z
identity: test-quality-20261116T012917-jay-flowers
tier: draft
as_of: "unbound-force@1677fd12d4ed8969d5ac87110e35ca030b9326b1"
evidence:
  - "gh:unbound-force/unbound-force#1048"
  - "dewey:.uf/artifacts/dispatch/a504842e-57da-41a4-926a-884c37671964.json"
---

When pinning a security structural ban regex across spec delta, design, and tasks, include the exact canonical pattern verbatim in all three artifacts. For the heredoc ban, `<<-?\s*\$?\\?\s*['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?` requires the second `\s*` after `\\?` to match `<<\ 'EOF'`; dropping it creates a false-negative on issue #1048. Verify table positives against the spec's canonical spelling, not a paraphrase.
