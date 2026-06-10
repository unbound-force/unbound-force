#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Pre-fetch PR context for council review. Runs trusted gh commands
# so Claude can read the results via Read tool without Shell access.
#
# Required env: GH_TOKEN, META_PATH
set -euo pipefail

PR_NUMBER=$(jq -r '.number' "${META_PATH}")
REPO=$(jq -r '.repo' "${META_PATH}")

# CI check results
gh pr checks "${PR_NUMBER}" --repo "${REPO}" \
  --json name,state,description 2>/dev/null \
  > pr-checks.json || echo '[]' > pr-checks.json

# Existing reviews (truncate long bodies)
gh api "repos/${REPO}/pulls/${PR_NUMBER}/reviews" \
  --jq '[.[] | {
    user: .user.login,
    state,
    body: (.body | if length > 500 then .[:500] + "..." else . end),
    submitted_at
  }]' 2>/dev/null > pr-reviews.json || echo '[]' > pr-reviews.json

# Existing inline comments (cap at 20)
gh api "repos/${REPO}/pulls/${PR_NUMBER}/comments" \
  --jq '[.[] | {
    path,
    line,
    body: (.body | if length > 300 then .[:300] + "..." else . end),
    user: .user.login,
    created_at
  }] | .[:20]' 2>/dev/null > pr-review-comments.json || echo '[]' > pr-review-comments.json

# Linked issues from PR body
PR_BODY=$(gh pr view "${PR_NUMBER}" --repo "${REPO}" \
  --json body --jq '.body // ""' 2>/dev/null) || PR_BODY=""
ISSUE_NUMS=$(echo "${PR_BODY}" | \
  grep -ioE '(fixes|closes|resolves)\s+#([0-9]+)' | \
  grep -oE '[0-9]+' | head -5) || true

echo '[]' > pr-linked-issues.json
if [[ -n "${ISSUE_NUMS}" ]]; then
  ISSUES_JSON="[]"
  for num in ${ISSUE_NUMS}; do
    ISSUE=$(gh issue view "${num}" --repo "${REPO}" \
      --json number,title,body,labels 2>/dev/null) || continue
    ISSUE=$(echo "${ISSUE}" | jq '{
      number,
      title,
      body: (.body | if length > 2000 then .[:2000] + "..." else . end),
      labels: [.labels[]?.name]
    }')
    ISSUES_JSON=$(echo "${ISSUES_JSON}" | jq --argjson issue "${ISSUE}" '. + [$issue]')
  done
  echo "${ISSUES_JSON}" > pr-linked-issues.json
fi
