---
tag: openspec-deltas
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:48:45Z
identity: openspec-deltas-20261107T164845-jay-flowers
tier: draft
---

An OpenSpec delta scenario that pins exact remediation string text (stop-shipping-plugin-lock SC-FR-004: "remediation names npm install --ignore-scripts --omit=dev --no-package-lock") couples the spec scenario to the Go remediation constant: any later correction of the remediation text requires amending the delta scenario within the same open change, or the fix is blocked by the pin. Pin the invariant the remediation must convey — the install command plus any runtime requirement — rather than a bare command string, so runtime-floor updates (such as an npm 11 requirement for devDependencies-carrying targets) do not need spec surgery to keep the guidance truthful.
