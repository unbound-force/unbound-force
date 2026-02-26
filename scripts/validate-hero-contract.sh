#!/usr/bin/env bash
# Hero Interface Contract Validation Script
# Version: 1.0.0
#
# Checks a repository against the Hero Interface Contract
# defined in specs/002-hero-interface-contract/contract.md.
#
# Usage: bash scripts/validate-hero-contract.sh /path/to/repo
#
# Exit codes:
#   0 - All required checks pass (PASS)
#   1 - One or more required checks fail (FAIL)
#   2 - Usage error (no path provided)

set -euo pipefail

CONTRACT_VERSION="1.0.0"

# --- Determine script directory for schema paths ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MANIFEST_SCHEMA="$REPO_ROOT/schemas/hero-manifest/v1.0.0.schema.json"

# --- Colors (if terminal supports them) ---
if [[ -t 1 ]]; then
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  YELLOW='\033[0;33m'
  NC='\033[0m'
else
  GREEN=''
  RED=''
  YELLOW=''
  NC=''
fi

# --- Counters ---
REQUIRED_TOTAL=0
REQUIRED_PASS=0
OPTIONAL_TOTAL=0
OPTIONAL_PASS=0

# --- Usage ---
if [[ $# -lt 1 ]]; then
  echo "Usage: bash scripts/validate-hero-contract.sh /path/to/repo"
  exit 2
fi

REPO_PATH="$1"

if [[ ! -d "$REPO_PATH" ]]; then
  echo "Error: '$REPO_PATH' is not a directory"
  exit 2
fi

# Resolve to absolute path
REPO_PATH="$(cd "$REPO_PATH" && pwd)"

# --- Check Functions ---

check_required() {
  local description="$1"
  local result="$2"  # "pass" or "fail"
  local detail="${3:-}"

  REQUIRED_TOTAL=$((REQUIRED_TOTAL + 1))
  if [[ "$result" == "pass" ]]; then
    REQUIRED_PASS=$((REQUIRED_PASS + 1))
    printf "  ${GREEN}[PASS]${NC} %s\n" "$description"
  else
    if [[ -n "$detail" ]]; then
      printf "  ${RED}[FAIL]${NC} %s (%s)\n" "$description" "$detail"
    else
      printf "  ${RED}[FAIL]${NC} %s\n" "$description"
    fi
  fi
}

check_optional() {
  local description="$1"
  local result="$2"  # "pass" or "warn"
  local detail="${3:-}"

  OPTIONAL_TOTAL=$((OPTIONAL_TOTAL + 1))
  if [[ "$result" == "pass" ]]; then
    OPTIONAL_PASS=$((OPTIONAL_PASS + 1))
    printf "  ${GREEN}[PASS]${NC} %s\n" "$description"
  else
    if [[ -n "$detail" ]]; then
      printf "  ${YELLOW}[WARN]${NC} %s (%s)\n" "$description" "$detail"
    else
      printf "  ${YELLOW}[WARN]${NC} %s\n" "$description"
    fi
  fi
}

# --- Header ---
echo "Hero Interface Contract Validation"
echo "==================================="
echo "Repository: $REPO_PATH"
echo "Contract version: $CONTRACT_VERSION"
echo ""

# --- Required Checks ---
echo "Required Checks:"

# 1. .specify/memory/constitution.md exists
if [[ -f "$REPO_PATH/.specify/memory/constitution.md" ]]; then
  check_required ".specify/memory/constitution.md exists" "pass"
else
  check_required ".specify/memory/constitution.md exists" "fail"
fi

# 2. .specify/templates/ exists and populated
if [[ -d "$REPO_PATH/.specify/templates/" ]]; then
  template_count=$(find "$REPO_PATH/.specify/templates/" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
  if [[ "$template_count" -ge 1 ]]; then
    check_required ".specify/templates/ exists and populated" "pass"
  else
    check_required ".specify/templates/ exists and populated" "fail" "directory exists but empty"
  fi
else
  check_required ".specify/templates/ exists and populated" "fail"
fi

# 3. .specify/scripts/bash/ exists and populated
if [[ -d "$REPO_PATH/.specify/scripts/bash/" ]]; then
  script_count=$(find "$REPO_PATH/.specify/scripts/bash/" -maxdepth 1 -name "*.sh" -type f 2>/dev/null | wc -l | tr -d ' ')
  if [[ "$script_count" -ge 1 ]]; then
    check_required ".specify/scripts/bash/ exists and populated" "pass"
  else
    check_required ".specify/scripts/bash/ exists and populated" "fail" "directory exists but empty"
  fi
else
  check_required ".specify/scripts/bash/ exists and populated" "fail"
fi

# 4. .opencode/ exists
if [[ -d "$REPO_PATH/.opencode/" ]]; then
  check_required ".opencode/ exists" "pass"
else
  check_required ".opencode/ exists" "fail"
fi

# 5. .opencode/command/ exists
if [[ -d "$REPO_PATH/.opencode/command/" ]]; then
  check_required ".opencode/command/ exists" "pass"
else
  check_required ".opencode/command/ exists" "fail"
fi

# 6. specs/ exists
if [[ -d "$REPO_PATH/specs/" ]]; then
  check_required "specs/ exists" "pass"
else
  check_required "specs/ exists" "fail"
fi

# 7. AGENTS.md exists
if [[ -f "$REPO_PATH/AGENTS.md" ]]; then
  check_required "AGENTS.md exists" "pass"
else
  check_required "AGENTS.md exists" "fail"
fi

# 8. LICENSE exists
if [[ -f "$REPO_PATH/LICENSE" ]]; then
  check_required "LICENSE exists" "pass"
else
  check_required "LICENSE exists" "fail"
fi

# 9. README.md exists
if [[ -f "$REPO_PATH/README.md" ]]; then
  check_required "README.md exists" "pass"
else
  check_required "README.md exists" "fail"
fi

# 10. .unbound-force/hero.json exists
if [[ -f "$REPO_PATH/.unbound-force/hero.json" ]]; then
  check_required ".unbound-force/hero.json exists" "pass"

  # 11. hero.json is valid JSON
  if python3 -m json.tool "$REPO_PATH/.unbound-force/hero.json" > /dev/null 2>&1; then
    check_required "hero.json is valid JSON" "pass"

    # 12. hero.json contains required fields
    required_fields=("name" "display_name" "role" "version" "description" "repository" "parent_constitution_version" "artifacts_produced" "artifacts_consumed" "opencode_agents" "opencode_commands" "dependencies")
    missing_fields=()
    for field in "${required_fields[@]}"; do
      if ! python3 -c "import json,sys; d=json.load(open(sys.argv[1])); assert '$field' in d" "$REPO_PATH/.unbound-force/hero.json" 2>/dev/null; then
        missing_fields+=("$field")
      fi
    done
    if [[ ${#missing_fields[@]} -eq 0 ]]; then
      check_required "hero.json contains required fields" "pass"
    else
      check_required "hero.json contains required fields" "fail" "missing: ${missing_fields[*]}"
    fi

    # 12b. hero.json validates against manifest schema (if schema file exists)
    if [[ -f "$MANIFEST_SCHEMA" ]]; then
      schema_result=""
      # Try python3 jsonschema first (best validation)
      if python3 -c "from jsonschema import validate" 2>/dev/null; then
        schema_result=$(python3 -c "
import json, sys
from jsonschema import validate, ValidationError
try:
    schema = json.load(open(sys.argv[1]))
    manifest = json.load(open(sys.argv[2]))
    validate(instance=manifest, schema=schema)
    print('pass')
except ValidationError as e:
    print('fail:' + e.message[:120])
except Exception as e:
    print('fail:' + str(e)[:120])
" "$MANIFEST_SCHEMA" "$REPO_PATH/.unbound-force/hero.json" 2>&1)
      # Fallback: try ajv (Node.js) if available
      elif command -v ajv &>/dev/null; then
        if ajv validate -s "$MANIFEST_SCHEMA" -d "$REPO_PATH/.unbound-force/hero.json" &>/dev/null; then
          schema_result="pass"
        else
          schema_result="fail:ajv validation error"
        fi
      else
        schema_result="skip"
      fi

      if [[ "$schema_result" == "pass" ]]; then
        check_optional "hero.json validates against manifest schema" "pass"
      elif [[ "$schema_result" == "skip" ]]; then
        check_optional "hero.json validates against manifest schema" "warn" "no validator available (install python3 jsonschema or ajv)"
      else
        detail="${schema_result#fail:}"
        check_optional "hero.json validates against manifest schema" "warn" "$detail"
      fi
    fi
  else
    check_required "hero.json is valid JSON" "fail" "JSON parse error"
    check_required "hero.json contains required fields" "fail" "cannot check - invalid JSON"
  fi
else
  check_required ".unbound-force/hero.json exists" "fail"
  # Skip JSON validity and field checks
  REQUIRED_TOTAL=$((REQUIRED_TOTAL + 2))
  printf "  ${RED}[FAIL]${NC} hero.json is valid JSON (file missing)\n"
  printf "  ${RED}[FAIL]${NC} hero.json contains required fields (file missing)\n"
fi

# 13. constitution contains parent_constitution reference
if [[ -f "$REPO_PATH/.specify/memory/constitution.md" ]]; then
  if grep -qi "parent.constitution" "$REPO_PATH/.specify/memory/constitution.md" 2>/dev/null; then
    check_required "constitution contains parent_constitution ref" "pass"
  else
    check_required "constitution contains parent_constitution ref" "fail" "no parent_constitution reference found"
  fi
else
  check_required "constitution contains parent_constitution ref" "fail" "constitution file missing"
fi

# --- Optional Checks ---
echo ""
echo "Optional Checks:"

# .opencode/agents/ exists
if [[ -d "$REPO_PATH/.opencode/agents/" ]]; then
  check_optional ".opencode/agents/ exists" "pass"
else
  check_optional ".opencode/agents/ exists" "warn" "no agents provided"
fi

# .github/workflows/ exists
if [[ -d "$REPO_PATH/.github/workflows/" ]]; then
  check_optional ".github/workflows/ exists" "pass"
else
  check_optional ".github/workflows/ exists" "warn" "no CI workflows"
fi

# Agent naming convention check (if agents exist)
if [[ -d "$REPO_PATH/.opencode/agents/" ]]; then
  non_compliant=()
  # Get the hero name from manifest or directory name
  hero_name=""
  if [[ -f "$REPO_PATH/.unbound-force/hero.json" ]]; then
    hero_name=$(python3 -c "import json,sys; print(json.load(open(sys.argv[1])).get('name',''))" "$REPO_PATH/.unbound-force/hero.json" 2>/dev/null || true)
  fi

  if [[ -n "$hero_name" ]]; then
    for agent_file in "$REPO_PATH/.opencode/agents/"*.md; do
      if [[ -f "$agent_file" ]]; then
        basename_file=$(basename "$agent_file" .md)
        # Check if agent name starts with hero name prefix
        # Also allow "reviewer-" prefix for The Divisor agents (deployed in other repos)
        if [[ ! "$basename_file" == "${hero_name}-"* ]] && [[ ! "$basename_file" == "reviewer-"* ]]; then
          non_compliant+=("$basename_file")
        fi
      fi
    done
  fi

  if [[ ${#non_compliant[@]} -gt 0 ]]; then
    check_optional "Agent naming convention ({hero}-{function})" "warn" "non-compliant: ${non_compliant[*]}"
  else
    check_optional "Agent naming convention ({hero}-{function})" "pass"
  fi
fi

# --- Summary ---
echo ""
if [[ $REQUIRED_PASS -eq $REQUIRED_TOTAL ]]; then
  printf "Overall: ${GREEN}PASS${NC} (%d/%d required, %d/%d optional)\n" \
    "$REQUIRED_PASS" "$REQUIRED_TOTAL" "$OPTIONAL_PASS" "$OPTIONAL_TOTAL"
  exit 0
else
  printf "Overall: ${RED}FAIL${NC} (%d/%d required, %d/%d optional)\n" \
    "$REQUIRED_PASS" "$REQUIRED_TOTAL" "$OPTIONAL_PASS" "$OPTIONAL_TOTAL"
  exit 1
fi
