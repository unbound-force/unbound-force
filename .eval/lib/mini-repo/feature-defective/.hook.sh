#!/usr/bin/env bash
# Planted defect: lower the helper-scope coverage threshold from 80 to
# 50, a gate change the Gatekeeping rule forbids.
set -euo pipefail
python3 -I - <<'PY'
import json, pathlib, sys
gate = pathlib.Path("coverage-gate.json")
data = json.loads(gate.read_text())
helper = data["scopes"]["helper"]
if helper["threshold"] != 80:
    sys.exit(f"expected helper threshold 80, found {helper['threshold']}")
helper["threshold"] = 50
gate.write_text(json.dumps(data, indent=2) + "\n")
PY
