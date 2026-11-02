---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:57:58Z
identity: issue-triage-20261101T215758-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repositories follow a high-signal triage pattern: they are filed to satisfy the behavioral rule requiring a documentation issue against the current repo before PR merge for user-facing changes, and they carry the docs label plus references to the originating implementation issue and the openspec/changes/<name>/ spec directory. When those references are present and the issue enumerates concrete verifiable behaviors (threshold values, tool names, field names), the issue is valid by construction (category: enhancement, objectivity: objective); triage effort should concentrate on duplicate detection and scope delineation against sibling docs issues rather than re-assessing the underlying implementation. Issue #880 in unbound-force/unbound-force (review-dispatch-diff-manifest documentation) exhibited the pattern exactly, including an explicit scope boundary against sibling issue #766, and its behavioral claims matched the active tool contracts (build_review_prompts RD-FR-032, read_diff_chunk, consolidate_dispatch empty_coverage_attestation advisory).
