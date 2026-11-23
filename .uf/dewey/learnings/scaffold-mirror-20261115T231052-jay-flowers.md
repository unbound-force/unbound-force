---
tag: scaffold-mirror
author: jay-flowers
category: gotcha
created_at: 2026-11-15T23:10:52Z
identity: scaffold-mirror-20261115T231052-jay-flowers
tier: draft
---

Scaffold mirror consistency chain (unbound-force): when a canonical OpenCode plugin (e.g. .opencode/plugins/uf-workflow/index.ts) gains an import of a new lib module (../../lib/<name>.js), the embedded mirror cp alone is NOT enough to make `make test` green. TestEmbeddedAssets_MatchSource/TestSCFR001 force mirror == canonical (the cp), but TestStageReviewPluginSources_CoversImportedLibModules (internal/scaffold/plugin_activation_test.go) walks the embedded plugin sources' lib imports and requires each referenced module to (1) exist as an embedded asset under internal/scaffold/assets/opencode/lib/, (2) be staged by stageReviewPluginSources, (3) be byte-identical, and (4) be isToolOwned + isDivisorAsset (Go list registrations in scaffold.go and plugin_activation.go). Therefore canonical plugin registration and the full lib mirror + Go list updates must land as one consistent unit (or be planned in the same phase); an intermediate state with only the plugin mirror copied fails the activation closure test. Symptom: "lib module <name> referenced by staged sources has no embedded asset opencode/lib/<name>.ts: file does not exist".
