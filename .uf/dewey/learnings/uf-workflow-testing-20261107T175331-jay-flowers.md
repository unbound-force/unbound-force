---
tag: uf-workflow-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:53:31Z
identity: uf-workflow-testing-20261107T175331-jay-flowers
tier: draft
---

In uf-workflow delegate tests, resolveSkillReference accepts a reference via `fileExists(abs) && !isDirectory(abs)`: a test that injects fileExists=true for a path but leaves isDirectory at the statSync-backed production default makes its outcome depend on host filesystem state at that path — the && short-circuit hides the un-injected dependency in every test whose fileExists returns false. Two deterministic shapes, chosen by test purpose: tests exercising the real default pass a mkdtemp scratch tree as cwd (the #948 sibling pattern); tests guarding the bare-`..` traversal rejection must inject an isDirectory double that reports the would-be target (resolve(cwd, '..')) as a non-directory — a literal scratch-tree port leaves cwd's parent a real directory, so the directory filter deterministically masks removal of the traversal guard and the test passes for the wrong reason on every host.
