---
tag: model-matrix
author: jay-flowers
category: gotcha
created_at: 2026-11-15T04:32:30Z
identity: model-matrix-20261115T043230-jay-flowers
tier: draft
---

In the kestrel workspace, uf_delegate_command children resolved through the model-matrix classifier to a provider endpoint at models.corp.redhat.com that is unreachable from this environment — every child failed twice with child_message_error "Cannot connect to API" before doing any work. Passing an explicit model override (enmaas/rits/zai-org/glm-5-3, the host session's own working provider) on each delegation made children connect immediately and complete normally. When matrix-resolved delegated children fail with connection errors in this workspace, retry once, then pass the explicit working model rather than retrying matrix resolution; the override is recorded in the delegation audit trail (selection_source: explicit-model).
