---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:13Z
identity: test-quality-20261107T180013-jay-flowers
tier: draft
---

A test fixture whose comment claims to mirror a canonical artifact becomes a false mirror under partial updates: when the canonical pins change, some fixture values get swept and others don't, leaving a fixture that is internally self-consistent (declared == installed, every test green) while no longer mirroring anything — yet the comment still says it does. PR #957's doctor fixture bumped zod to 4.6.6 but left @opencode-ai/plugin at 1.4.10 against a canonical 1.19.0 pin. Either pin the fixture to the canonical source it names (read and compare, or derive it), or word the comment to claim only the shape being exercised (package names, section layout), never the values.
