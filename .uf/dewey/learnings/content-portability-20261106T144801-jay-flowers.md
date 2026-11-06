---
tag: content-portability
author: jay-flowers
category: pattern
created_at: 2026-11-06T14:48:01Z
identity: content-portability-20261106T144801-jay-flowers
tier: draft
---

OpenSpec MODIFIED-requirement deltas restate the full requirement, so external repository references inside carried-forward scenarios can look like content-portability violations while actually being accurate pre-existing spec content. Verify before flagging: (1) check the referenced org/repo against the repository's actual `uses:` pins, and (2) diff the delta's "Previously" paragraph and scenarios against the live `openspec/specs/<capability>/spec.md` to separate carried-forward text from newly introduced text. Applied on unbound-force PR #972 (fix-pinact-gaze-pin): the ci capability delta's dependabot scenario names `complytime/org-infra`, which looked foreign to the unbound-force meta repo, but `.github/workflows/ci_crapload.yml:26` actually consumes `complytime/org-infra/.github/workflows/reusable_crapload_analysis.yml` and the scenario matches the existing spec verbatim — so no portability finding was warranted. This extends the grep-verify discipline from documentation-gap triage to spec-delta portability checks.
