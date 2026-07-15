#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
#
# Filter inline review comments to only lines visible in the PR
# diff. The GitHub Pull Request Review API rejects comments on
# lines outside diff hunks with HTTP 422 "Line could not be
# resolved". This script parses the unified diff to build a set
# of valid (path, line) pairs, then filters the review JSON.
#
# Usage: python3 filter-diff-lines.py <diff-file> <review-json>
# Output: filtered review JSON to stdout
import json
import re
import sys


def parse_diff_lines(diff_path):
    """Extract valid (path, line) pairs from a unified diff.

    For the RIGHT side (new version), valid lines are:
    - Added lines ('+' prefix): the new-file line number
    - Context lines (' ' prefix): the new-file line number
    Both are commentable via the GitHub Review API.
    """
    valid = set()
    current_path = None
    new_line = 0

    with open(diff_path) as f:
        for raw in f:
            line = raw.rstrip("\n")

            if line.startswith("diff --git"):
                match = re.search(r" b/(.+)$", line)
                if match:
                    current_path = match.group(1)
                continue

            hunk = re.match(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@", line)
            if hunk:
                new_line = int(hunk.group(1))
                continue

            if current_path is None:
                continue

            if line.startswith("+") and not line.startswith("+++"):
                valid.add((current_path, new_line))
                new_line += 1
            elif line.startswith("-") and not line.startswith("---"):
                pass  # deleted lines don't increment new_line
            elif not line.startswith("\\"):
                valid.add((current_path, new_line))
                new_line += 1

    return valid


def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <diff-file> <review-json>",
              file=sys.stderr)
        sys.exit(1)

    diff_path, review_path = sys.argv[1], sys.argv[2]

    valid_lines = parse_diff_lines(diff_path)

    with open(review_path) as f:
        review = json.load(f)

    original = review.get("inline_comments", [])
    filtered = []
    rescued = []

    for c in original:
        if (c.get("path"), c.get("line")) in valid_lines:
            filtered.append(c)
        else:
            rescued.append(c)

    if rescued:
        print(f"Filtered {len(rescued)}/{len(original)} comments "
              f"(lines not in diff) — appending to summary",
              file=sys.stderr)
        extra = "\n\n### Additional findings (not on diff lines)\n\n"
        for c in rescued:
            path = c.get("path", "unknown")
            line = c.get("line", "?")
            body = c.get("body", "")
            extra += f"**{path}:{line}** — {body}\n\n"
        review["summary"] = review.get("summary", "") + extra

    review["inline_comments"] = filtered
    json.dump(review, sys.stdout)


if __name__ == "__main__":
    main()
