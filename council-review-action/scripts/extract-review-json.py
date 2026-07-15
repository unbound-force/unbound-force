#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
#
# Extract structured review JSON from OpenCode text output.
# Handles both raw JSON and JSON inside markdown code fences.
# Scans backwards from the end of the text to find the last
# valid JSON object containing "summary" and "inline_comments".
import json
import re
import sys

text = open("review_text.txt").read()

# Strip markdown code fences (```json ... ``` or ``` ... ```)
text = re.sub(r"```(?:json)?\s*\n?", "", text)

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
        try:
            obj = json.loads(candidate[: j + 1])
            if "summary" in obj and "inline_comments" in obj:
                json.dump(obj, sys.stdout)
                sys.exit(0)
        except (json.JSONDecodeError, ValueError):
            pass
sys.exit(1)
