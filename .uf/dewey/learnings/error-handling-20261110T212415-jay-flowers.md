---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:24:15Z
identity: error-handling-20261110T212415-jay-flowers
tier: draft
---

Classifying subprocess failures by matching an error-message signature requires pinning the child's locale and anchoring the signature to a line start. An unanchored strings.Contains match lets a target path that embeds the phrase inside another error spoof the skip classification, and a translated locale rewrites the message entirely. Durable pattern (unbound-force internal/doctor checkTrackedPluginLock): run the child as 'env LC_ALL=C <tool> ...' so gettext ignores LANGUAGE and emits C-locale messages, and match with a multiline-anchored regex like (?m)^fatal: not a git repository\\b so the signature must start a line. Test both directions: a spoofed mid-line occurrence must surface as a real failure (warn), and a genuine signature following preamble output must still classify as the skip.
