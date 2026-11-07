---
tag: skill-migration
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:01:58Z
identity: skill-migration-20261107T160158-jay-flowers
tier: draft
---

Retiring a filesystem root requires sweeping every consumer in the same change, not just the primary resolver. PR #948 retired the legacy .opencode/skill/ root in uf_delegate_command's SKILL_ROOTS and added the uf init migration, but internal/doctor/checks.go still scans both roots (`for _, skillBase := range []string{"skill", "skills"}` at ~line 1159), leaving a vestigial read path that validates a directory the toolchain now considers retired. Post-migration it degrades to a no-op (os.Stat fails, scan skipped), so it is harmless — but a thorough retirement greps all consumers (resolvers, validators, doctor checks, scaffold assets, .gitignore, AGENTS.md structure, test fixtures) and either updates them or files a tracked follow-up in the same change.
