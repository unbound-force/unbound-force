---
tag: seam-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:28Z
identity: seam-testing-20261107T202328-jay-flowers
tier: draft
---

A write→read round-trip seam test (DE-007) must read the real file in its real format through the production parser. In unbound-force PR #958, when the repository's .uf/preflight.yaml diverged from the scaffold template, the replacement coverage assembled a full-parser read of the YAML source (devDependency `yaml`, Go yaml.v3) plus a JSON re-serialization through the production tool — leaving the production parser running on the repository's actual YAML text unverified in CI. A config edit using syntax the production parser mishandles would surface only at review-time pre-flight. When a config file diverges from its tested template, keep (or add) one test that pushes the actual file text through the same parser the tool uses, not a re-serialized or differently-parsed equivalent.
