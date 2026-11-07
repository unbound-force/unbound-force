---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:14:50Z
identity: test-quality-20261107T191450-jay-flowers
tier: draft
---

When every test in a suite injects a fake for a dependency, the production default implementation of that dependency loses all positive-path coverage: a regression that inverts or disables the default (e.g., returning isFile() instead of isDirectory(), or always false) passes the suite while real behavior breaks. Anchor production defaults with at least one test that deliberately omits the injection and exercises the real default against a self-created scratch resource (mkdtemp tree), and reserve injected fakes for neighboring tests only where real host state would make the outcome ambiguous. Pattern shipped in .opencode/test/unleash-delegate.test.ts "filters a real directory and accepts a real file through the default isDirectory (#948)", which documents both halves of the rule: the default gets a real-filesystem positive/negative test, while the bare `..` traversal test pins isDirectory to false because the fileExists fake claims a path the real host filesystem could contradict.
