---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:07:57Z
identity: dependency-management-20261110T190757-jay-flowers
tier: draft
---

A security override that forces a nested exact-pinned transitive onto a different version changes resolved code, not just audit metadata: `@opencode-ai/plugin@1.19.0` declares `effect` exactly at 4.0.0-beta.83, and the override moves it to 4.0.0-beta.103 — beta-channel packages can change API between pre-releases. A green throwaway-lock OSV scan proves only vulnerability absence, not API compatibility; functional validation of the consuming suite (`make plugin-test` on the plugin code that actually loads the parent package) is the required compensating evidence before merging an override pin.
