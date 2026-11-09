---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:15Z
identity: agent-permissions-20261108T224015-jay-flowers
tier: draft
---

Prefix-allow bash permission models validated with unanchored regex match multi-line commands on their first line: `node allowed-cli.ts --json\nrm -rf ~` still satisfies `node allowed-cli.ts*` because the unanchored `.*` finds a match in line one, and terminal metacharacter denies (`>*`, `;*`, `|*`, `$(*`, backtick) never see a plain second line. When hardening an agent's bash map with catch-all denies, enumerate the newline vector explicitly: assert in the permission-model test what happens to a command containing `\n`, and if the engine matches prefix allows across newlines, add an explicit guard (deny any pattern that spans lines, or reject multi-line commands before matching). PR #959 round 3: the DDP-FR-004 hardening test covers metacharacters but not newlines (Adversary LOW finding).
