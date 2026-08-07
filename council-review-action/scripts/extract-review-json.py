#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
#
# Extract structured review JSON from OpenCode text output.
# Handles both raw JSON and JSON inside markdown code fences.
# Scans backwards from the end of the text to find the last
# valid JSON object containing "summary" and "inline_comments".
#
# Usage: python3 extract-review-json.py [input-file]
# Default input file: review_text.txt
import json
import re
import sys

# Size limit: reject inputs > 2 MB to prevent quadratic parsing
# on adversarial input. Typical review output is < 100 KB.
MAX_INPUT_BYTES = 2 * 1024 * 1024

# Iteration limit: cap the number of parse attempts to prevent
# unbounded CPU usage on inputs with many '{' or '}' characters.
MAX_PARSE_ATTEMPTS = 500

input_file = sys.argv[1] if len(sys.argv) > 1 else "review_text.txt"

with open(input_file, encoding="utf-8", errors="replace") as f:
    text = f.read(MAX_INPUT_BYTES + 1)

if len(text) > MAX_INPUT_BYTES:
    print(f"Input exceeds {MAX_INPUT_BYTES} bytes, skipping",
          file=sys.stderr)
    sys.exit(1)

# Strip markdown code fences (```json ... ``` or ``` ... ```)
text = re.sub(r"```(?:json)?\s*\n?", "", text)

attempts = 0
i = len(text)
while i > 0:
    i = text.rfind("{", 0, i)
    if i < 0:
        break
    candidate = text[i:]
    # Find the matching closing brace by trying progressively
    # shorter substrings from each '}' found from the end.
    j = len(candidate)
    while j > 0:
        j = candidate.rfind("}", 0, j)
        if j < 0:
            break
        attempts += 1
        if attempts > MAX_PARSE_ATTEMPTS:
            print(f"Exceeded {MAX_PARSE_ATTEMPTS} parse attempts",
                  file=sys.stderr)
            sys.exit(1)
        try:
            obj = json.loads(candidate[: j + 1])
            if "summary" in obj and "inline_comments" in obj:
                json.dump(obj, sys.stdout)
                sys.exit(0)
        except (json.JSONDecodeError, ValueError):
            pass
sys.exit(1)
