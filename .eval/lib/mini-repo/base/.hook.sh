# Register the base packages in the files the snapshot owns, so the
# base commit is internally consistent and the feature diff carries
# only the feature. Fails if upstream moved the anchors these edits
# rely on, rather than silently producing an inconsistent base.
set -euo pipefail
python3 -I - <<'PY'
import json, pathlib, sys

agents = pathlib.Path("AGENTS.md")
text = agents.read_text()
anchor = "│   ├── homebrew/"
if anchor not in text:
    sys.exit("AGENTS.md project tree no longer has the internal/homebrew line")
lines = text.split("\n")
i = next(n for n, l in enumerate(lines) if l.startswith(anchor))
lines[i + 1:i + 1] = [
    "│   ├── provenance/                   # Artifact source provenance records",
    "│   ├── refs/                         # Git reference helpers",
]
agents.write_text("\n".join(lines))

gate = pathlib.Path("coverage-gate.json")
data = json.loads(gate.read_text())
helper = data["scopes"]["helper"]["packages"]
helper += ["internal/provenance", "internal/refs"]
gate.write_text(json.dumps(data, indent=2) + "\n")
PY
