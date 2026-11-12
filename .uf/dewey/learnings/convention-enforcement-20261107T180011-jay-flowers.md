---
tag: convention-enforcement
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:00:11Z
identity: convention-enforcement-20261107T180011-jay-flowers
tier: draft
---

When a check family injects a filesystem seam (doctor's Options.ReadFile with an os.ReadFile fallback), every new helper added to that family must thread the seam even when its tests can happily use real t.TempDir files. A helper that calls os.ReadFile directly still passes today's tests, but the seam contract erodes one helper at a time, and the next test that simulates an environment through the injected reader (missing, malformed, or unreadable manifest) silently misses the new read path. The failure mode is a confusingly passing or failing test, never a loud one. Review heuristic: in packages whose Options struct carries a ReadFile (or equivalent) field, grep newly added file-reading helpers for direct os.ReadFile calls.
