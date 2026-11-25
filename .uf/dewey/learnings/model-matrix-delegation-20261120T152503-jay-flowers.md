---
tag: model-matrix-delegation
author: jay-flowers
category: gotcha
created_at: 2026-11-20T15:25:03Z
identity: model-matrix-delegation-20261120T152503-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
evidence:
  - "dewey:.uf/model-matrix.override.yaml"
  - "dewey:.opencode/plugins/uf-workflow/index.ts"
  - "session:ses_e068c7546ffe2JLVHqo3mOdohC"
  - "session:ses_e067caae2ffeCGw3dzt1te2pDq"
---

Delegating spec-authoring work through uf_delegate_command failed four consecutive times on this machine (2026-11-20, branch opsx/intent-satisfaction-check, issue #563): child sessions were created with the matrix-selected identity cobalt-crush-dev @ Muse-Spark-Open-816B-A42B-NVFP4 (per the .uf/model-matrix.override.yaml spec-authoring profile), but their message parts actually executed as the default build agent on host model glm-5-3. The delegation's model_mismatch gate correctly failed the one run that completed (child reported glm-5-3, expected Muse-Spark), and the other three attempts aborted with MessageAbortedError mid-run — twice leaving partial uncommitted spec artifacts on the branch. The EnMaaS provider serves Muse-Spark directly (verified with a /v1/chat/completions call), and the review-dispatch path demonstrably runs Muse-Spark children, so the gap is specific to the uf-workflow child spawn/prompt-submission path losing the requested agent+model. The critical lesson: never work around model_mismatch by passing an explicit model matching the child's actual model — that spoofs the governance check (a Gatekeeping violation). Treat the mismatch as blocking, diagnose the divergence (compare the session's creation identity against part-level agent/model in the session DB), report it, and leave any partial artifacts uncommitted for human review rather than laundering them past the gate.
