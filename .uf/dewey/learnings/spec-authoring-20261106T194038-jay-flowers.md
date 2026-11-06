---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-06T19:40:38Z
identity: spec-authoring-20261106T194038-jay-flowers
tier: draft
---

When a verification check iterates a package allow-list, name the list in the requirement text. Requirement wording like "agreement between the package.json pins and the installed versions" reads as every pin including devDependencies, while the implementation may check only a runtime allow-list scoped by a design decision — the requirement then over-promises. In a repository that declares its own devDependencies, those pins end up with no doctor verification path unless the wording is tightened ("each declared runtime pin") or the check is extended. Observed: stop-shipping-plugin-lock SC-FR-005 (spec.md:139) vs checkManifestInstallConsistency iterating reviewPluginPackages in internal/doctor/checks.go.
