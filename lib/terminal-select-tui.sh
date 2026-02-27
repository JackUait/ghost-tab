#!/bin/bash
# Terminal selection TUI wrapper using ghost-tab-tui

# Interactive terminal selection.
# Returns 0 if selected, 1 if cancelled.
# Sets: _selected_terminal
#
# If GHOST_TAB_TERMINAL_PREF is set, reads current terminal from that file
# and passes it to the TUI via --current flag.
select_terminal_interactive() {
  if ! command -v ghost-tab-tui &>/dev/null; then
    error "ghost-tab-tui binary not found. Please reinstall."
    return 1
  fi

  while true; do
    # Build command args
    local args=("select-terminal")
    if [[ -n "$GHOST_TAB_TERMINAL_PREF" && -f "$GHOST_TAB_TERMINAL_PREF" ]]; then
      local current
      current="$(tr -d '[:space:]' < "$GHOST_TAB_TERMINAL_PREF")"
      if [[ -n "$current" ]]; then
        args+=("--current" "$current")
      fi
    fi

    local result
    # Capture output regardless of exit code — bubbletea may exit non-zero
    # during cleanup even when the user completed an action successfully.
    result=$(ghost-tab-tui "${args[@]}" 2>/dev/null) || true

    # If no output at all, the TUI failed to run
    if [[ -z "$result" ]]; then
      return 1
    fi

    local selected
    if ! selected=$(echo "$result" | jq -r '.selected' 2>/dev/null); then
      error "Failed to parse terminal selection response"
      return 1
    fi

    # Check for install action — run brew then loop back to TUI
    local action
    action=$(echo "$result" | jq -r '.action // empty' 2>/dev/null)
    if [[ "$action" == "install" ]]; then
      local cask terminal
      cask=$(echo "$result" | jq -r '.cask // empty' 2>/dev/null)
      terminal=$(echo "$result" | jq -r '.terminal // empty' 2>/dev/null)
      # Fall back to terminal name if cask field is missing
      if [[ -z "$cask" ]]; then
        cask="$terminal"
      fi
      if [[ -n "$cask" && "$cask" != "null" ]]; then
        echo ""
        info "Installing $cask via Homebrew..."
        if brew install --cask "$cask"; then
          success "Installed $cask"
        else
          error "Failed to install $cask"
        fi
        echo ""
      fi
      continue
    fi

    # User cancelled
    if [[ -z "$selected" || "$selected" == "null" || "$selected" != "true" ]]; then
      return 1
    fi

    # User selected an installed terminal
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
  done
}
