#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
#
# Extract structured review JSON from OpenCode text output.
# Scans backwards from the end of the text to find the last
# valid JSON object containing "summary" and "inline_comments".
import json
import sys

text = open("review_text.txt").read()
i = len(text)
while i > 0:
    i = text.rfind("{", 0, i)
    if i < 0:
        break
    try:
        obj = json.loads(text[i:])
        if "summary" in obj and "inline_comments" in obj:
            json.dump(obj, sys.stdout)
            sys.exit(0)
    except (json.JSONDecodeError, ValueError):
        pass
sys.exit(1)
