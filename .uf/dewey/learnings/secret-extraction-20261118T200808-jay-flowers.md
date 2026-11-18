---
tag: secret-extraction
author: jay-flowers
category: pattern
created_at: 2026-11-18T20:08:08Z
identity: secret-extraction-20261118T200808-jay-flowers
tier: draft
as_of: "unbound-force@cd2520c42c9cd75797136ab6ba23bbdd6ca6bc29"
---

When a helper extracts a single credential out of a larger aggregate secret, validate the JSON type of the value, not just the rendered text: `jq -r` renders numbers, arrays, and objects into output that prefix checks for reference forms miss, so a wrong-typed value is exported as a plausible-looking credential and the misconfiguration surfaces later as an unattributed authentication failure. Pair the type check with an explicit ambiguity check (more than one provider matching the selector) rather than silently taking `.[0]`.
