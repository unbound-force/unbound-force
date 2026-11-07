---
tag: input-validation
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:58:43Z
identity: input-validation-20261107T165843-jay-flowers
tier: draft
---

When a validation guard cites an external model (git's remote-name rules, a platform's identifier grammar) as its basis, port the model's complete rejection set for the guarded failure mode, not just its most familiar rule. In unbound-force PR #948, RemoteNameSchema modeled itself on git's remote-name validation but ported only the '/' rejection; '.' and '..' passed (they contain no slash), yet git resolves both as local repository paths when used as push targets ('.' = the current repo, '..' = the parent directory), and 'git push <path>' executes the target repository's hooks with the pusher's privileges. The guard's own doc comment described the exact failure mode the omitted rules protect against. Practice: for each rejection rule in the cited model, port it or record why it is out of scope, and enumerate the model's boundary values ('.'/'..' for path-like resolution) as INVALID_INPUT test rows. Corollary for this repo: fixes to .opencode/plugins/uf-workflow/index.ts must also update the byte-identical scaffold mirror internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts, because drift tests enforce identity and a canonical-only fix fails the suite.
