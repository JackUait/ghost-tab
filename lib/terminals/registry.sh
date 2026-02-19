#!/bin/bash
# Terminal registry — lists supported terminals and manages preference.

# Print supported terminal identifiers, one per line.
get_supported_terminals() {
  echo "ghostty"
  echo "iterm2"
  echo "wezterm"
  echo "kitty"
}

# Return the human-readable display name for a terminal identifier.
get_terminal_display_name() {
  local terminal="$1"
  case "$terminal" in
    ghostty) echo "Ghostty" ;;
    iterm2)  echo "iTerm2" ;;
    wezterm) echo "WezTerm" ;;
    kitty)   echo "kitty" ;;
    *)       echo "$terminal" ;;
  esac
}

# Read saved terminal preference from file. Prints the terminal name or empty.
load_terminal_preference() {
  local pref_file="$1"
  if [ -f "$pref_file" ]; then
    tr -d '[:space:]' < "$pref_file"
  fi
}

# Save terminal preference to file.
save_terminal_preference() {
  local terminal="$1" pref_file="$2"
  mkdir -p "$(dirname "$pref_file")"
  echo "$terminal" > "$pref_file"
}

# Detect if user has a legacy Ghostty-only installation.
# Prints "ghostty" if detected, empty otherwise.
detect_legacy_ghostty_setup() {
  local old_wrapper="$HOME/.config/ghostty/claude-wrapper.sh"
  local pref_file="${XDG_CONFIG_HOME:-$HOME/.config}/ghost-tab/terminal"

  if [ -f "$old_wrapper" ] || [ -L "$old_wrapper" ]; then
    if [ ! -f "$pref_file" ]; then
      echo "ghostty"
      return 0
    fi
  fi
  return 1
}
