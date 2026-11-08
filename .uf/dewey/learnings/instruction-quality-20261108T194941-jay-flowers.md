---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T19:49:41Z
identity: instruction-quality-20261108T194941-jay-flowers
tier: draft
---

Replacing a command line inside a fenced code block can double the closing fence: if the edit adds a new closing fence directly above the unchanged original, the file carries two consecutive ``` lines and fence parity flips for everything below — code renders as prose, prose as code, and fence-aware tooling (instruction-lint's fenced map, the IL-042 delegation scanner) misclassifies regions file-wide. PR 959 did this to divisor-linguist.md:109-110 during the Step 1 command replacement. After any fence-adjacent edit, verify fences still alternate open/close from the section start, or render the section to confirm structure.
