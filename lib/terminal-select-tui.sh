#!/bin/bash
# Terminal selection TUI wrapper using ghost-tab-tui

# Interactive terminal selection.
# Returns 0 if selected, 1 if cancelled.
# Sets: _selected_terminal
select_terminal_interactive() {
  if ! command -v ghost-tab-tui &>/dev/null; then
    error "ghost-tab-tui binary not found. Please reinstall."
    return 1
  fi

  local result
  if ! result=$(ghost-tab-tui select-terminal 2>/dev/null); then
    return 1
  fi

  local selected
  if ! selected=$(echo "$result" | jq -r '.selected' 2>/dev/null); then
    error "Failed to parse terminal selection response"
    return 1
  fi

  if [[ -z "$selected" || "$selected" == "null" || "$selected" != "true" ]]; then
    return 1
  fi

  local terminal
  if ! terminal=$(echo "$result" | jq -r '.terminal' 2>/dev/null); then
    error "Failed to parse selected terminal"
    return 1
  fi

  if [[ -z "$terminal" || "$terminal" == "null" ]]; then
    error "TUI returned empty terminal selection"
    return 1
  fi

  _selected_terminal="$terminal"
  return 0
}
