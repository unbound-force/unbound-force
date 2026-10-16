---
tag: opencode-plugins
author: jay-flowers
category: gotcha
created_at: 2026-10-16T13:27:09Z
identity: opencode-plugins-20261016T132709-jay-flowers
tier: draft
---

When auditing OpenCode plugin tools for the parameters:/args: SDK defect, only factories that actually declare a `parameters:` key are broken. Parameterless tools (no declaration at all) are correct — the SDK treats a missing `args` key as a no-argument tool. Prior learning opencode-plugin-sdk-20261015T234420-jay-flowers overcounts the defect as "14 legacy tools in uf-workflow"; that figure is all non-persist factories including 6 parameterless ones. The verified count is 8 parameterized factories in uf-workflow plus 3 in uf-finale (grep 'parameters:' under .opencode/plugins). Spec reviews of changes that enumerate code sites should independently grep-verify the enumeration against source — miscounts in prior learnings propagate into requirement text (FR counts, task lists) if trusted uncritically.
