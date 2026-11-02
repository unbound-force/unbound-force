---
tag: cli-architecture
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T15:33:23Z
identity: cli-architecture-20261102T153323-yvonne-devlin
tier: draft
---

Mirroring an established package pattern copies its vestigial surface. internal/gate's Options struct carries Format and Stdout fields that Run() defaults but never consumes — formatting responsibility actually lives at the CLI boundary (cmd/unbound-force/gate.go's format switch calling gate.FormatJSON/FormatText). Verified on opsx/uf-gate (2026-11-02): grep shows o.Format/o.Stdout appear only inside defaults() in BOTH internal/gate and internal/doctor, the pattern source — the gate package faithfully mirrors doctor, dead fields included. When cloning a structural pattern into a new package, audit each mirrored Options field for a live consumer in the new context; unconsumed config fields mislead API consumers about where responsibility lives and invite callers to set values that silently do nothing.
