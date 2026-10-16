---
tag: scaffold-parity
author: jay-flowers
category: pattern
created_at: 2026-10-16T15:11:33Z
identity: scaffold-parity-20261016T151133-jay-flowers
tier: draft
---

Identical git diff index pairs across canonical and scaffold-mirror files prove byte-identical parity, and mismatched pairs are immediate drift evidence — treat the Go drift test suite (internal/scaffold) as the CI backstop rather than the only check. This pattern from the fix-tool-args-schema-mismatch review (uf-workflow and uf-finale plugin mirrors) gives reviewers a zero-cost parity verification inside any unified diff.
