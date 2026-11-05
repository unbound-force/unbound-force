---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-04T22:58:27Z
identity: review-dispatch-20261104T225827-jay-flowers
tier: draft
---

When a plugin tool's result outgrows the host's tool-output window (host truncation of large JSON, file-read tools cutting long lines), do not relocate the truncation — persist the data once under a strict schema and serve bounded pages through a purpose-built reader. The read_dispatch_artifact design (unbound-force PR #949, RD-FR-057) is the reference shape: derive the file path from a UUID-validated correlation_id so no caller-supplied path exists; open with O_NOFOLLOW|O_NONBLOCK (feature-detecting both f
