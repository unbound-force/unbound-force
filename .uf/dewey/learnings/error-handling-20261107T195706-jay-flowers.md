---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:57:06Z
identity: error-handling-20261107T195706-jay-flowers
tier: draft
---

git localizes fatal messages via gettext, so classifying a benign exit by matching English stderr text (e.g., "fatal: not a git repository") is locale-sensitive — a translated environment takes the failure branch instead of the skip. Locale-stable alternatives: git rev-parse --is-inside-work-tree (returns true/false regardless of locale) or forcing LC_ALL=C on the invocation. When stderr-signature matching is kept anyway, route the mismatch to a visible warning with the raw command output appended so the degradation stays observable instead of hiding the cause.
