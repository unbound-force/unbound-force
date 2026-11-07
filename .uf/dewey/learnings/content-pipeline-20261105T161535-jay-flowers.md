---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:15:35Z
identity: content-pipeline-20261105T161535-jay-flowers
tier: draft
---

Open website docs issues double as contract descriptions, and a PR that changes a documented contract must reference or update them even when the PR's own in-repo docs are complete. Verified instances during the PR #948 review (2026-11-05): website #310 (open) documents git_push's parameter surface as (setUpstream, branch), but PR 948 makes `branch` required and adds `remote`/`expected_sha` with BRANCH_MISMATCH/EXPECTED_SHA_UNKNOWN failure codes; website #317 and in-repo #791 (open) document skill resolution against .opencode/skills/ then .opencode/skill/ with no repo-root-file fallback, but PR 948 drops the legacy root and adds repo-root-file resolution. Website docs eventually written from those issue bodies would describe stale contracts. When reviewing a contract-changing PR, grep open docs issues for the parameter names and resolution rules the PR alters, and name the stale issues in the documentation-disposition finding.
