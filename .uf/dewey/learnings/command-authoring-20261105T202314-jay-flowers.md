---
tag: command-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:23:14Z
identity: command-authoring-20261105T202314-jay-flowers
tier: draft
---

When adding or strengthening an ordering or keying MUST in an agent command file, sweep every instruction site that performs the governed call — including early-file session-resume guards and companion tracking sections — not just the section being rewritten. Resume-path instructions execute precisely when the rest of the command (including the new rule) is no longer in context, so an unfixed resume step silently overrides the new rule and reintroduces the failure the change was meant to fix. Precedent: PR #955 / issue #954 — /uf.address-feedback gained \"Never call get_progress or checkpoint from another branch\" while its SESSION-RESUME GUARD (line 18) and TodoWrite resume instruction (line 75) still directed unconditional pre-checkout get_progress calls, keying resume state to whatever branch the shell occupied.
