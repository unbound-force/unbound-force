---
tag: opsx-extract-feedback-triage-artifact-tool
author: jay-flowers
category: gotcha
created_at: 2026-11-16T08:43:40Z
identity: opsx-extract-feedback-triage-artifact-tool-20261116T084340-jay-flowers
tier: draft
---

On resuming /uf.unleash after a crash, uf_unleash_detect_resume can report the spec-review step incomplete even when the dispatch artifact on disk, relay_spec_review_gate, and the checkpoint record all show a decided APPROVE WITH ADVISORIES iteration — observed on opsx/extract-feedback-triage-artifact-tool after a context-thrash abort: detect_resume returned spec_review false with resume_from spec_review, while relay_spec_review_gate returned gate_paused false with undecided_finding_ids empty against the approving artifact 8f64702d. The reconciliation that worked: trust the purpose-built gate relay and the persisted dispatch artifact (the command's own "artifacts authoritative" rule) over the step boolean, and never re-dispatch a decided approving iteration — re-running an approved review wastes a full council run and risks a contradictory REQUEST CHANGES verdict on identical artifacts, which would then trigger a human decision gate for no reason. Checking the relay before re-entering a review step on resume is a one-call insurance policy against that whole failure chain.
