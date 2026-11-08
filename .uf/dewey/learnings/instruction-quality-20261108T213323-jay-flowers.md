---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:23Z
identity: instruction-quality-20261108T213323-jay-flowers
tier: draft
---

instruction-lint --base REF --changed scopes findings by file, not by line: a pre-existing finding anchored at an unchanged line of a changed file surfaces in delta mode. In PR 959, IL-023 ("appropriate" in a normative sentence) fired at divisor-curator.md:169 — verified present at the merge base — while the change touched only that file's frontmatter bash block. The reviewer must reconcile the file-scoped linter output against the caller's changed-line confinement: verify the finding exists at the merge base (or trust pre-flight evidence that did), then document the exclusion in the review summary with that evidence. Reporting it violates the confinement rule; dropping it silently violates the never-drop-a-linter-finding-silently rule; the documented exclusion satisfies both.
