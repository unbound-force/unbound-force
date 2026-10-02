---
tag: acceptance-criteria-review
author: jay-flowers
category: pattern
created_at: 2026-09-29T22:33:48Z
identity: acceptance-criteria-review-20260929T223348-jay-flowers
tier: draft
---

The review pipeline's acceptance-criteria evaluation uses prompt-based logic (agent instructions in markdown), not compiled Go code. Testing this logic requires prompt-based test fixtures rather than Go unit tests. The council-review-action pipeline test (test-pipeline.sh) validates shell-script behavior (build-prompt.sh, run-review.sh) with grep -qF assertions directly against files, not via piped echo commands which race under set -o pipefail.
