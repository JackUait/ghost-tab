#!/bin/bash
# Input parsing helpers — no side effects on source.

# Interactive confirmation using ghost-tab-tui
# Usage: confirm_tui "Delete project 'foo'?"
# Returns: 0 if confirmed, 1 if cancelled
confirm_tui() {
  local msg="$1"

  if ! command -v ghost-tab-tui &>/dev/null; then
    # Fallback to simple bash prompt
    read -rp "$msg (y/N) " response </dev/tty
    [[ "$response" =~ ^[Yy]$ ]]
    return $?
  fi

  local result
  if ! result=$(ghost-tab-tui confirm "$msg" 2>/dev/null); then
    return 1
  fi

  local confirmed
  if ! confirmed=$(echo "$result" | jq -r '.confirmed' 2>/dev/null); then
    # Source tui.sh for error function if not already loaded
    if ! declare -F error &>/dev/null; then
      echo "ERROR: Failed to parse confirmation response" >&2
    else
      error "Failed to parse confirmation response"
    fi
    return 1
  fi

  # Validate against "null" string (learned from Task 3)
  if [[ "$confirmed" == "null" || -z "$confirmed" ]]; then
    return 1
  fi

  [[ "$confirmed" == "true" ]]
}
