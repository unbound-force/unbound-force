---
tag: go-path-safety
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T15:09:15Z
identity: go-path-safety-20261102T150915-yvonne-devlin
tier: draft
---

filepath.Glob treats runtime-supplied directory components as patterns; metacharacters can cause false artifact-not-found failures or discarded ErrBadPattern. Escape root components or enumerate directories directly.
