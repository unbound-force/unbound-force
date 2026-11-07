---
tag: seam-testing
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:15Z
identity: seam-testing-20261107T160715-jay-flowers
tier: draft
---

When reviewing a CLI whose entry path depends on a language-runtime feature of uncertain floor-version support (e.g. Node's `import.meta.main`, whose availability across minor versions is easy to doubt), look for a seam test that execs the real entry file as a subprocess and asserts a specific non-zero exit code before flagging a speculative compatibility finding. On unbound-force PR #958, instruction-lint.test.ts's "exits 2 when the linter crashes" test execs `node .opencode/lib/instruction-lint.ts` directly and expects exit 2 from the `import.meta.main` block's catch handler — a green CI run of that test on the engines-floor Node proves the entry block executes, settling the availability question with evidence instead of a reviewer's version recall. The same pattern resolves type-stripping doubts for directly-executed .ts entry files.
