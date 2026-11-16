---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T23:27:07Z
identity: instruction-quality-20261115T232707-jay-flowers
tier: draft
as_of: "unbound-force@505dedf886fceeb2e552b7c1cf7633e2b6c36310"
---

Advisory instruction prose that prescribes an ongoing action — "run X in the background and poll the output file" — needs a completion criterion and a failure fallback even when the note is classified as applicability judgment. Judgment classification (DE-001/DE-005, design Decision 4 of idle-timeout-aborts-children-waiting-on) excuses the trigger condition (when to background), not the loop's termination (what marks the poll done) or its failure path: shell redirection (`> file`) creates the output file empty at launch, so file existence cannot signal completion, and a backgrounded command that fails returns no directly observed exit code, so the command's synchronous Error Handling section never fires. Without those two conditions, agents reading the same note diverge between polling file existence, polling content stability, or waiting forever on a failed launch.
