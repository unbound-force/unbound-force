---
tag: cli-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-08T20:00:25Z
identity: cli-tooling-20261108T200025-jay-flowers
tier: draft
---

When consolidating CLI plumbing between sibling CLIs into a shared library in the same PR, the extraction checklist must include the output-safety boundary, not just plumbing: PR 959 created cli-common.ts (findRoot, git, CliIO, runMain) while instruction-lint.ts kept its escapeControlChars print-site boundary module-private, so the new tool-inventory.ts shipped text-mode print sites interpolating file-derived fields (tool descriptions parsed from plugin source, .uf/scripts and .opencode/lib first-comment summaries) unescaped — invisible at authoring time because JSON mode escapes via JSON.stringify and the precedent CLI's boundary made the requirement look satisfied. Determinism framing: control-character escaping passes the twice-test (fixed procedure, same output every time), so it belongs in the shared tested library, applied once per CLI at a single print-site wrapper (`const print = (s) => io.out(escapeControlChars(s))`) rather than wrapped per interpolated field; move it with a re-export so existing test imports (instruction-lint.test.ts imports escapeControlChars from ../lib/instruction-lint.js) keep pinning the behavior contract through the move.
