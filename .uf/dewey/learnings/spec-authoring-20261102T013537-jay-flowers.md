---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:35:37Z
identity: spec-authoring-20261102T013537-jay-flowers
tier: draft
---

When authoring spec/design supersession or governance records, verify every verbatim quote byte-exact against its source and resolve every cited path from the repo root before writing it. Two failure modes observed in init-override-stubs-gitignore-sync iteration 2: (1) hand-transcribed quotes normalize Unicode punctuation — the proposal rendered the scaffold marker's em-dash (gitignoreMarker, scaffold.go:1426) as an ASCII double hyphen, making the quoted form un-greppable against the code; (2) citation paths are written from memory of the repo layout — design cited specs/gitignore.md, but the file lives at openspec/changes/archive/2026-04-28-gitignore-init/specs/gitignore.md after archival. Both are caught by a single grep/read of the cited source before publishing the record.
