---
tag: test-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:04Z
identity: test-gates-20261110T195004-jay-flowers
tier: draft
---

A repo-wide regex gate over command templates that rejects `<< identifier` matches anywhere in the file — prose, code fences, any language — so left-shift or stream-insertion examples in future template documentation (e.g., `1 << flags`, `cout << value`) trip a payload-heredoc gate with a misleading error message. Scope such content gates to bash/sh code fences or to command positions (line-leading or after `;`/`&&`) so documentation examples do not fail payload-specific checks.
