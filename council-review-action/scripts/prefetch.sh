#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Pre-fetch PR context for council review. Runs trusted gh commands
# so OpenCode can read the results via Read tool without Shell access.
#
# Review bodies and inline comments are kept at full length — human
# reviewer feedback is high-signal context that the AI reviewer needs
# to avoid duplicating findings or contradicting prior feedback. The
# total token cost (~25-30K tokens) is well within the 200K context.
#
# Required env: GH_TOKEN, META_PATH
set -euo pipefail

PR_NUMBER=$(jq -r '.number' "${META_PATH}")
REPO=$(jq -r '.repo' "${META_PATH}")

# CI check results
gh pr checks "${PR_NUMBER}" --repo "${REPO}" \
  --json name,state,description 2>/dev/null \
  > pr-checks.json || echo '[]' > pr-checks.json

# Existing reviews — full bodies preserved
gh api "repos/${REPO}/pulls/${PR_NUMBER}/reviews" \
  --paginate \
  --jq '[.[] | {
    user: .user.login,
    state,
    body,
    submitted_at
  }]' 2>/dev/null > pr-reviews.json || echo '[]' > pr-reviews.json

# Existing inline comments — full bodies, cap count at 50
gh api "repos/${REPO}/pulls/${PR_NUMBER}/comments" \
  --paginate \
  --jq '[.[] | {
    path,
    line,
    body,
    user: .user.login,
    created_at
  }] | .[:50]' 2>/dev/null > pr-review-comments.json || echo '[]' > pr-review-comments.json

# Linked issues from PR body — full bodies preserved
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
      body,
      labels: [.labels[]?.name]
    }')
    ISSUES_JSON=$(echo "${ISSUES_JSON}" | jq --argjson issue "${ISSUE}" '. + [$issue]')
  done
  echo "${ISSUES_JSON}" > pr-linked-issues.json
fi
