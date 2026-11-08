---
tag: cli-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-08T20:00:28Z
identity: cli-tooling-20261108T200028-jay-flowers
tier: draft
---

When a second runnable CLI joins .opencode/lib/, the change must complete the shared-library extraction surface in the same increment: every helper both CLIs need — including print-hygiene contracts like control-character escaping — moves to cli-common.ts, and the new CLI adopts the sibling's established boundary shape rather than per-field wrapping. The boundary shape that works: one `const print = (s) => io.out(escapeControlChars(s))` wrapper for text-mode output (JSON branches stay on raw io.out, mirroring instruction-lint's printText). Because escapeControlChars keeps tab and newline, whole-line wrapping preserves pinned listing layouts (`name  [plugin]  file:line\n  description`) and existing text-output assertions stay green. Two gotchas observed in PR 959: (1) whitespace collapsing (`one()`'s `replace(/\s+/g, " ")`) is NOT control-character sanitization — ESC/OSC/C1 sequences pass through it, so a description field that looks cleaned still needs the escape boundary; (2) test-injected io writers bypass runMain's process writers, so the boundary must live inside each CLI's main()/print path to be test-covered — instruction-lint additionally cannot use runMain at all (pre-Node-24.2 import.meta.main comparison). Extends learning security-20261107T160712 (sanitize once at the print site) with the sibling-CLI onboarding rule: review a new lib CLI by diffing its print boundary against its sibling's, not just against its own feature spec.
