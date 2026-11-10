---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:46:30Z
identity: error-handling-20261108T184630-jay-flowers
tier: draft
---

Command instruction that directs an agent to POST a JSON payload built from reviewer-authored or other external text must require a JSON-safe composer (jq -n --arg key "$value", or gh api --field) rather than hand-interpolating text into a JSON template: a quoted heredoc delimiter prevents shell injection but does not prevent JSON invalidity from double quotes, newlines, or backslashes in the text. Pair every API mutation step with a stated failure outcome whenever a protocol invariant makes later steps illegal on failure — in uf.address-feedback, the spec requires the issue to exist before the FILE reply links it, so a failed issue-creation POST must retry once and then stop and report, never post the reply. In-repo precedent: uf_finale_create_pr writes the PR body through a secure temp file; the uf.address-feedback reply block uses temp-file + --input.
