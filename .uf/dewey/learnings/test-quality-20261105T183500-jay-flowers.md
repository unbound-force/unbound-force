---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:35:00Z
identity: test-quality-20261105T183500-jay-flowers
tier: draft
---

When a check folds multiple failure causes into one condition with a shared observable outcome (e.g. `json.Unmarshal err != nil || field == ""` both mapping to the same "unreadable" message), a single fixture that trips the failure proves the message contract but silently covers only one disjunct. To keep branch intent visible, either give each cause its own fixture (malformed JSON vs empty field vs injected read error) or name the test after the cause it actually exercises — otherwise test names overstate coverage and a refactor that breaks one disjunct passes silently.
