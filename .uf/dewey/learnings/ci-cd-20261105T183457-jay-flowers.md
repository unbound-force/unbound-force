---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:34:57Z
identity: ci-cd-20261105T183457-jay-flowers
tier: draft
---

When a CI workflow must reproduce a tool version that a Makefile variable already pins (e.g. `PLUGIN_TEST_NPM ?= npx -y npm@11.6.2` in the Makefile vs the same literal in a ci_security.yml job), route the CI step through a make target that uses the variable, or at minimum add a cross-reference comment in both files naming the twin. Prose coordination ("coordinated with PLUGIN_TEST_NPM" in a design doc) drifts silently: when one pin is bumped, the other keeps generating stale artifacts (here, the throwaway lock OSV scan would keep using the old npm).
