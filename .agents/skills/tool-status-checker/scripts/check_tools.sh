#!/usr/bin/env bash
# Tool Status Checker Script for macOS / Linux Gemini CLI Environment

set -euo pipefail

# Resolve project root relative to this script
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
SANDBOX="$PROJECT_ROOT/sandbox"

check_tool() {
  local name="$1"
  local cmd="$2"
  local args="${3:---version}"

  printf "Checking %s..." "$name"
  if [ -x "$cmd" ] || command -v "$cmd" >/dev/null 2>&1; then
    ver=$("$cmd" $args 2>&1 | head -1) || true
    if [ -n "$ver" ]; then
      printf " [OK] (%s)\n" "$ver"
    else
      printf " [FAILED]\n"
    fi
  else
    printf " [NOT FOUND]\n"
  fi
}

echo "=== Core Tools Status ==="
check_tool "curl" "curl"
check_tool "uv"   "$SANDBOX/uv/uv"
check_tool "python" "$(find "$SANDBOX/python" -maxdepth 2 -name 'python3*' -type f 2>/dev/null | head -1)"
check_tool "cat"  "cat"

echo ""
echo "=== Environment Check ==="
if [ -d "$SANDBOX/git" ]; then
  echo "Git sandbox: [FOUND] at $SANDBOX/git"
else
  echo "Git sandbox: [NOT FOUND]"
fi
