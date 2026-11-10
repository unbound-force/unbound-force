---
tag: injection-safety
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:07Z
identity: injection-safety-20261110T152107-jay-flowers
tier: draft
---

When splicing user-controlled strings into a structured config file, route every value through a real encoder rather than manual quoting: internal/scaffold's jsonString (json.Encoder with SetEscapeHTML(false)) makes JSON-structure injection impossible by construction — a version payload like `1.0.0", "zod": "4.2.0` renders as one escaped string value instead of injecting a duplicate key that could downgrade a security pin. Two companion rules make the splice safe and stable: search splice anchors only in the tool-owned (embedded, fixed) bytes so user content cannot move the splice boundaries, and render merged keys in sorted order so byte-identity skip logic and merge idempotence hold across repeated runs.
