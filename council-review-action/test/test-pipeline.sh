#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# End-to-end tests for the council review diff pipeline:
#   prepare-diff.sh  → pr-diff-filtered.patch (noise removed)
#                    → pr-diff-annotated.patch (line-annotated)
#   filter-diff-lines.py → validates (path, line) pairs
#
# Run from the council-review-action directory:
#   bash test/test-pipeline.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../scripts" && pwd)"
WORK_DIR=$(mktemp -d)
trap 'rm -rf "${WORK_DIR}"' EXIT

PASS=0
FAIL=0

assert_eq() {
  local label="$1" expected="$2" actual="$3"
  if [[ "${expected}" == "${actual}" ]]; then
    echo "  PASS: ${label}"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: ${label}"
    echo "    expected: ${expected}"
    echo "    actual:   ${actual}"
    FAIL=$((FAIL + 1))
  fi
}

assert_contains() {
  local label="$1" needle="$2" haystack="$3"
  if echo "${haystack}" | grep -qF "${needle}"; then
    echo "  PASS: ${label}"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: ${label} — '${needle}' not found"
    FAIL=$((FAIL + 1))
  fi
}

assert_not_contains() {
  local label="$1" needle="$2" haystack="$3"
  if ! echo "${haystack}" | grep -qF "${needle}"; then
    echo "  PASS: ${label}"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: ${label} — '${needle}' unexpectedly found"
    FAIL=$((FAIL + 1))
  fi
}

# ── Test 1: New file annotation ──────────────────────────────
echo "Test 1: New file — lines annotated 1..N"

cat > "${WORK_DIR}/new-file.patch" << 'DIFF'
diff --git a/hello.py b/hello.py
new file mode 100644
index 0000000..abcdef1
--- /dev/null
+++ b/hello.py
@@ -0,0 +1,4 @@
+#!/usr/bin/env python3
+def main():
+    print("hello")
+main()
DIFF

(cd "${WORK_DIR}" && DIFF_PATH=new-file.patch bash "${SCRIPT_DIR}/prepare-diff.sh")

ANNOTATED=$(cat "${WORK_DIR}/pr-diff-annotated.patch")
assert_contains "L1 on first line"  "[L1] +#!/usr/bin/env python3" "${ANNOTATED}"
assert_contains "L2 on second line" "[L2] +def main():"             "${ANNOTATED}"
assert_contains "L3 on third line"  '[L3] +    print("hello")'      "${ANNOTATED}"
assert_contains "L4 on fourth line" "[L4] +main()"                  "${ANNOTATED}"
assert_not_contains "no annotation on hunk header" "[L" "$(grep '^@@' "${WORK_DIR}/pr-diff-annotated.patch")"

# ── Test 2: Modified file with context ───────────────────────
echo "Test 2: Modified file — context + added lines"

cat > "${WORK_DIR}/mod-file.patch" << 'DIFF'
diff --git a/lib/utils.py b/lib/utils.py
index abcdef1..abcdef2 100644
--- a/lib/utils.py
+++ b/lib/utils.py
@@ -10,6 +10,8 @@ def existing():
     pass
 
 def another():
+    # new comment
+    log("added")
     pass
 
 def last():
DIFF

(cd "${WORK_DIR}" && DIFF_PATH=mod-file.patch bash "${SCRIPT_DIR}/prepare-diff.sh")

ANNOTATED=$(cat "${WORK_DIR}/pr-diff-annotated.patch")
assert_contains "context at L10" "[L10]      pass"          "${ANNOTATED}"
assert_contains "context at L12" "[L12]  def another():"    "${ANNOTATED}"
assert_contains "added at L13"   "[L13] +    # new comment" "${ANNOTATED}"
assert_contains "added at L14"   '[L14] +    log("added")'  "${ANNOTATED}"
assert_contains "context at L15" "[L15]      pass"          "${ANNOTATED}"

# ── Test 3: Multiple files — line numbers reset per file ─────
echo "Test 3: Multi-file — line numbers reset per file"

cat > "${WORK_DIR}/multi.patch" << 'DIFF'
diff --git a/a.txt b/a.txt
new file mode 100644
--- /dev/null
+++ b/a.txt
@@ -0,0 +1,3 @@
+line one
+line two
+line three
diff --git a/b.txt b/b.txt
new file mode 100644
--- /dev/null
+++ b/b.txt
@@ -0,0 +1,2 @@
+alpha
+beta
DIFF

(cd "${WORK_DIR}" && DIFF_PATH=multi.patch bash "${SCRIPT_DIR}/prepare-diff.sh")

ANNOTATED=$(cat "${WORK_DIR}/pr-diff-annotated.patch")
# a.txt lines 1-3
assert_contains "a.txt L1" "[L1] +line one"   "${ANNOTATED}"
assert_contains "a.txt L3" "[L3] +line three"  "${ANNOTATED}"
# b.txt resets to 1
assert_contains "b.txt L1" "[L1] +alpha"       "${ANNOTATED}"
assert_contains "b.txt L2" "[L2] +beta"        "${ANNOTATED}"

# ── Test 4: Noise filtering ─────────────────────────────────
echo "Test 4: Noise files filtered out"

cat > "${WORK_DIR}/noise.patch" << 'DIFF'
diff --git a/go.sum b/go.sum
new file mode 100644
--- /dev/null
+++ b/go.sum
@@ -0,0 +1,2 @@
+hash1
+hash2
diff --git a/main.go b/main.go
new file mode 100644
--- /dev/null
+++ b/main.go
@@ -0,0 +1,1 @@
+package main
DIFF

(cd "${WORK_DIR}" && DIFF_PATH=noise.patch bash "${SCRIPT_DIR}/prepare-diff.sh")

FILTERED=$(cat "${WORK_DIR}/pr-diff-filtered.patch")
assert_not_contains "go.sum excluded" "go.sum" "${FILTERED}"
assert_contains "main.go kept"  "main.go"      "${FILTERED}"

# ── Test 5: Deleted lines don't get annotations ─────────────
echo "Test 5: Deleted lines have no [L] prefix"

cat > "${WORK_DIR}/delete.patch" << 'DIFF'
diff --git a/f.txt b/f.txt
index abcdef1..abcdef2 100644
--- a/f.txt
+++ b/f.txt
@@ -5,7 +5,6 @@ header
 context
 keep
-removed line
 after
+added line
 end
DIFF

(cd "${WORK_DIR}" && DIFF_PATH=delete.patch bash "${SCRIPT_DIR}/prepare-diff.sh")

ANNOTATED=$(cat "${WORK_DIR}/pr-diff-annotated.patch")
assert_not_contains "no annotation on deleted" "[L" "$(grep '^-removed' "${WORK_DIR}/pr-diff-annotated.patch" || true)"
assert_contains "context L5 correct"  "[L5]  context"    "${ANNOTATED}"
assert_contains "context L6 correct"  "[L6]  keep"       "${ANNOTATED}"
assert_contains "context L7 correct"  "[L7]  after"      "${ANNOTATED}"
assert_contains "added L8 correct"    "[L8] +added line" "${ANNOTATED}"
assert_contains "context L9 correct"  "[L9]  end"        "${ANNOTATED}"

# ── Test 6: filter-diff-lines.py validation ──────────────────
echo "Test 6: filter-diff-lines.py — valid vs invalid lines"

cat > "${WORK_DIR}/filter-test.patch" << 'DIFF'
diff --git a/app.py b/app.py
new file mode 100644
--- /dev/null
+++ b/app.py
@@ -0,0 +1,3 @@
+import os
+def run():
+    pass
DIFF

cat > "${WORK_DIR}/review.json" << 'JSON'
{
  "summary": "Test review",
  "inline_comments": [
    {"path": "app.py", "line": 2, "body": "valid — on diff line"},
    {"path": "app.py", "line": 99, "body": "invalid — beyond file"},
    {"path": "nope.py", "line": 1, "body": "invalid — wrong file"}
  ]
}
JSON

RESULT=$(python3 "${SCRIPT_DIR}/filter-diff-lines.py" \
  "${WORK_DIR}/filter-test.patch" "${WORK_DIR}/review.json")

KEPT=$(echo "${RESULT}" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['inline_comments']))")
assert_eq "1 comment kept" "1" "${KEPT}"

SUMMARY=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin)['summary'])")
assert_contains "rescued invalid to summary" "app.py:99" "${SUMMARY}"
assert_contains "rescued wrong file to summary" "nope.py:1" "${SUMMARY}"

# ── Test 7: filter-diff-lines.py — modified file hunks ──────
echo "Test 7: filter-diff-lines.py — only hunk lines valid"

cat > "${WORK_DIR}/partial.patch" << 'DIFF'
diff --git a/big.py b/big.py
index abcdef1..abcdef2 100644
--- a/big.py
+++ b/big.py
@@ -20,4 +20,5 @@ class Foo:
     pass
 
     def bar(self):
+        return True
         pass
DIFF

cat > "${WORK_DIR}/partial-review.json" << 'JSON'
{
  "summary": "Partial review",
  "inline_comments": [
    {"path": "big.py", "line": 23, "body": "on the added line"},
    {"path": "big.py", "line": 10, "body": "NOT in the hunk"},
    {"path": "big.py", "line": 50, "body": "beyond hunk range"}
  ]
}
JSON

RESULT=$(python3 "${SCRIPT_DIR}/filter-diff-lines.py" \
  "${WORK_DIR}/partial.patch" "${WORK_DIR}/partial-review.json")

KEPT=$(echo "${RESULT}" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['inline_comments']))")
assert_eq "1 valid hunk comment" "1" "${KEPT}"

# ── Test 8: extract-review-json.py — raw JSON ───────────────
echo "Test 8: extract-review-json.py — raw JSON input"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
{"summary": "Looks good.", "inline_comments": [{"path": "a.py", "line": 1, "body": "Nice"}]}
TXT

RESULT=$(cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py")
SUMMARY=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin)['summary'])")
assert_eq "raw JSON summary" "Looks good." "${SUMMARY}"

# ── Test 9: extract-review-json.py — markdown fences ────────
echo "Test 9: extract-review-json.py — JSON in code fences"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
Here is my review:

```json
{"summary": "Fenced review.", "inline_comments": []}
```

Thank you.
TXT

RESULT=$(cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py")
SUMMARY=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin)['summary'])")
assert_eq "fenced JSON summary" "Fenced review." "${SUMMARY}"

# ── Test 10: extract-review-json.py — extra text around JSON ─
echo "Test 10: extract-review-json.py — JSON with surrounding text"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
I reviewed the PR carefully. Here are my findings:

{"summary": "Embedded in text.", "inline_comments": [{"path": "x.go", "line": 5, "body": "test"}]}

That concludes my review.
TXT

RESULT=$(cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py")
COUNT=$(echo "${RESULT}" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['inline_comments']))")
assert_eq "extracted 1 inline comment" "1" "${COUNT}"

# ── Test 11: extract-review-json.py — invalid input ──────────
echo "Test 11: extract-review-json.py — no valid JSON exits 1"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
This has no JSON at all, just plain text review.
TXT

if (cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py" > /dev/null 2>&1); then
  echo "  FAIL: should have exited 1"
  FAIL=$((FAIL + 1))
else
  echo "  PASS: exits 1 on invalid input"
  PASS=$((PASS + 1))
fi

# ── Test 12: extract-review-json.py — missing keys ──────────
echo "Test 12: extract-review-json.py — JSON without required keys"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
{"status": "ok", "message": "not a review"}
TXT

if (cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py" > /dev/null 2>&1); then
  echo "  FAIL: should reject JSON missing summary/inline_comments"
  FAIL=$((FAIL + 1))
else
  echo "  PASS: rejects JSON without required keys"
  PASS=$((PASS + 1))
fi

# ── Test 13: build-prompt.sh — generates prompt with title ───
echo "Test 13: build-prompt.sh — prompt contains PR title"

mkdir -p "${WORK_DIR}/prompt-test"
echo '{"title": "feat: add dark mode toggle"}' \
  > "${WORK_DIR}/prompt-test/meta.json"

(cd "${WORK_DIR}/prompt-test" && \
  META_PATH=meta.json AGENT_MODE=multi \
  bash "${SCRIPT_DIR}/build-prompt.sh")

PROMPT=$(cat "${WORK_DIR}/prompt-test/review_prompt.txt")
assert_contains "PR title in prompt" "feat: add dark mode toggle" "${PROMPT}"
assert_contains "output format section" "OUTPUT FORMAT" "${PROMPT}"
assert_contains "annotated patch reference" "pr-diff-annotated.patch" "${PROMPT}"
assert_contains "line annotation guidance" "[L<N>]" "${PROMPT}"

# ── Test 14: build-prompt.sh — title truncation ─────────────
echo "Test 14: build-prompt.sh — long title truncated at 200 chars"

LONG_TITLE=$(python3 -c "print('x' * 300)")
jq -n --arg t "${LONG_TITLE}" '{title: $t}' \
  > "${WORK_DIR}/prompt-test/meta-long.json"

(cd "${WORK_DIR}/prompt-test" && \
  META_PATH=meta-long.json AGENT_MODE=single \
  bash "${SCRIPT_DIR}/build-prompt.sh")

TITLE_IN_PROMPT=$(grep "^PR Title:" "${WORK_DIR}/prompt-test/review_prompt.txt" | sed 's/^PR Title: //')
TITLE_LEN=${#TITLE_IN_PROMPT}
if [[ "${TITLE_LEN}" -le 200 ]]; then
  echo "  PASS: title truncated to ${TITLE_LEN} chars"
  PASS=$((PASS + 1))
else
  echo "  FAIL: title is ${TITLE_LEN} chars (expected ≤200)"
  FAIL=$((FAIL + 1))
fi

# ── Test 15: build-prompt.sh — security instructions ────────
echo "Test 15: build-prompt.sh — contains security instructions"

PROMPT=$(cat "${WORK_DIR}/prompt-test/review_prompt.txt")
assert_contains "untrusted input warning" "untrusted input" "${PROMPT}"
assert_contains "no shell commands" "Do NOT run shell commands" "${PROMPT}"
assert_contains "no subagents" "Do NOT spawn subagents" "${PROMPT}"

# ── Test 16: parse-output.sh — no input file ────────────────
echo "Test 16: parse-output.sh — no input produces comment fallback"

PARSE_DIR=$(mktemp -d)
(
  cd "${PARSE_DIR}"
  GITHUB_OUTPUT="${PARSE_DIR}/gh_output"
  touch "${GITHUB_OUTPUT}"
  DIFF_PATH=pr-diff-filtered.patch \
    SCRIPT_DIR="${SCRIPT_DIR}" \
    GITHUB_OUTPUT="${GITHUB_OUTPUT}" \
    bash "${SCRIPT_DIR}/parse-output.sh"
)
MODE=$(grep "review_mode=" "${PARSE_DIR}/gh_output" | tail -1 | cut -d= -f2)
assert_eq "fallback mode on no input" "comment" "${MODE}"
rm -rf "${PARSE_DIR}"

# ── Test 17: parse-output.sh — direct JSON input ────────────
echo "Test 17: parse-output.sh — direct JSON passes through"

PARSE_DIR=$(mktemp -d)
cat > "${PARSE_DIR}/review_raw.txt" << 'JSON'
{"summary": "Direct JSON.", "inline_comments": [{"path": "a.go", "line": 1, "body": "ok"}]}
JSON
cat > "${PARSE_DIR}/pr-diff-filtered.patch" << 'DIFF'
diff --git a/a.go b/a.go
new file mode 100644
--- /dev/null
+++ b/a.go
@@ -0,0 +1,2 @@
+package main
+func main() {}
DIFF
(
  cd "${PARSE_DIR}"
  GITHUB_OUTPUT="${PARSE_DIR}/gh_output"
  touch "${GITHUB_OUTPUT}"
  DIFF_PATH=pr-diff-filtered.patch \
    SCRIPT_DIR="${SCRIPT_DIR}" \
    GITHUB_OUTPUT="${GITHUB_OUTPUT}" \
    bash "${SCRIPT_DIR}/parse-output.sh"
)
MODE=$(grep "review_mode=" "${PARSE_DIR}/gh_output" | tail -1 | cut -d= -f2)
SUMMARY=$(jq -r '.summary' "${PARSE_DIR}/review_output.json")
assert_eq "direct JSON mode" "inline" "${MODE}"
assert_eq "direct JSON summary" "Direct JSON." "${SUMMARY}"
rm -rf "${PARSE_DIR}"

# ── Test 18: parse-output.sh — plain text fallback ──────────
echo "Test 18: parse-output.sh — plain text falls back to comment"

PARSE_DIR=$(mktemp -d)
echo "Just some plain text review." > "${PARSE_DIR}/review_raw.txt"
(
  cd "${PARSE_DIR}"
  GITHUB_OUTPUT="${PARSE_DIR}/gh_output"
  touch "${GITHUB_OUTPUT}"
  DIFF_PATH=pr-diff-filtered.patch \
    SCRIPT_DIR="${SCRIPT_DIR}" \
    GITHUB_OUTPUT="${GITHUB_OUTPUT}" \
    bash "${SCRIPT_DIR}/parse-output.sh"
)
MODE=$(grep "review_mode=" "${PARSE_DIR}/gh_output" | tail -1 | cut -d= -f2)
assert_eq "plain text mode" "comment" "${MODE}"
rm -rf "${PARSE_DIR}"

# ── Test 19: empty diff produces empty filtered output ──────
echo "Test 19: Empty diff — no crash"

EMPTY_DIR=$(mktemp -d)
touch "${EMPTY_DIR}/empty.patch"
(cd "${EMPTY_DIR}" && DIFF_PATH=empty.patch bash "${SCRIPT_DIR}/prepare-diff.sh")
FILTERED_SIZE=$(wc -c < "${EMPTY_DIR}/pr-diff-filtered.patch" | tr -d ' ')
rm -rf "${EMPTY_DIR}"

# Empty diff should produce a file (possibly with just a newline)
if [[ "${FILTERED_SIZE}" -le 1 ]]; then
  echo "  PASS: empty diff handled"
  PASS=$((PASS + 1))
else
  echo "  FAIL: expected empty output, got ${FILTERED_SIZE} bytes"
  FAIL=$((FAIL + 1))
fi

# ── Test 20: extract-review-json.py — multiple JSON objects ──
echo "Test 20: extract-review-json.py — picks last valid JSON"

cat > "${WORK_DIR}/review_text.txt" << 'TXT'
{"status": "partial", "inline_comments": []}
Some text in between.
{"summary": "Final review.", "inline_comments": [{"path": "b.py", "line": 1, "body": "ok"}]}
TXT

RESULT=$(cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py")
SUMMARY=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin)['summary'])")
assert_eq "picks last valid JSON" "Final review." "${SUMMARY}"

# ── Test 21: extract-review-json.py — size limit ────────────
echo "Test 21: extract-review-json.py — rejects oversized input"

python3 -c "print('{' * 3000000)" > "${WORK_DIR}/review_text.txt"

if (cd "${WORK_DIR}" && python3 "${SCRIPT_DIR}/extract-review-json.py" > /dev/null 2>&1); then
  echo "  FAIL: should reject oversized input"
  FAIL=$((FAIL + 1))
else
  echo "  PASS: rejects oversized input"
  PASS=$((PASS + 1))
fi

# ── Test 22: filter-diff-lines.py — deeper assertion ────────
echo "Test 22: filter-diff-lines.py — rescued comments in summary"

cat > "${WORK_DIR}/depth-test.patch" << 'DIFF'
diff --git a/main.go b/main.go
new file mode 100644
--- /dev/null
+++ b/main.go
@@ -0,0 +1,2 @@
+package main
+func main() {}
DIFF

cat > "${WORK_DIR}/depth-review.json" << 'JSON'
{
  "summary": "Base summary.",
  "inline_comments": [
    {"path": "main.go", "line": 1, "body": "valid"},
    {"path": "main.go", "line": 99, "body": "out of range"},
    {"path": "other.go", "line": 1, "body": "wrong file"}
  ]
}
JSON

RESULT=$(python3 "${SCRIPT_DIR}/filter-diff-lines.py" \
  "${WORK_DIR}/depth-test.patch" "${WORK_DIR}/depth-review.json")

KEPT=$(echo "${RESULT}" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['inline_comments']))")
SUMMARY=$(echo "${RESULT}" | python3 -c "import sys,json; print(json.load(sys.stdin)['summary'])")

assert_eq "1 valid comment kept" "1" "${KEPT}"
assert_contains "rescued has out-of-range" "main.go:99" "${SUMMARY}"
assert_contains "rescued has wrong-file" "other.go:1" "${SUMMARY}"
assert_contains "base summary preserved" "Base summary." "${SUMMARY}"

# ── Test 23: build-prompt.sh — JSON schema fields present ───
echo "Test 23: build-prompt.sh — JSON schema fields in prompt"

PROMPT=$(cat "${WORK_DIR}/prompt-test/review_prompt.txt")
assert_contains "schema has summary field" '"summary"' "${PROMPT}"
assert_contains "schema has inline_comments" '"inline_comments"' "${PROMPT}"
assert_contains "schema has path field" '"path"' "${PROMPT}"
assert_contains "schema has line field" '"line"' "${PROMPT}"
assert_contains "schema has body field" '"body"' "${PROMPT}"

# ── Summary ──────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Results: ${PASS} passed, ${FAIL} failed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ "${FAIL}" -gt 0 ]]; then
  exit 1
fi
