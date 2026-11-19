---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:28:36Z
identity: dependencies-20261110T202836-jay-flowers
tier: draft
---

An npm `overrides` entry forces EVERY copy of the package to the pinned version — including the copy a direct dependency declares at a newer version later. A security override added to force a vulnerable transitive onto a fixed version (effect 4.0.0-beta.103 overriding @opencode-ai/plugin's declared 4.0.0-beta.83) silently downgrades the nested copy once the direct dependency ships its own fixed version, and nothing flags the skew: manifest-vs-install checks read direct pins, and OSV scans fire only on advisories affecting the overridden version. When adding an override, record a retirement criterion (drop it when the direct dependency's declared version supersedes the override) and extend any operator verification commands (`npm ls <pkg>`) to include the overridden package.
