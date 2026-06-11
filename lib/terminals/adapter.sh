#!/bin/bash
# Terminal adapter loader — sources the correct adapter for a terminal.

# Load the adapter for the given terminal identifier.
# After calling this, terminal_setup_config, terminal_get_config_path, etc. are available.
load_terminal_adapter() {
  local terminal="$1"
  local adapter_dir
  adapter_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

  local adapter_file="$adapter_dir/${terminal}.sh"
  if [ ! -f "$adapter_file" ]; then
    error "Unknown terminal: $terminal"
    return 1
  fi

  # shellcheck disable=SC1090  # Dynamic adapter loading
  source "$adapter_file"
}

# Remove ghost-tab config from the previously selected terminal when switching.
# Reads the old terminal from the preference file; no-op when there is no
# previous preference, the terminal is unchanged, or the adapter is unknown.
# Runs in a subshell so the old adapter cannot clobber the loaded adapter.
# Usage: cleanup_previous_terminal <pref_file> <new_terminal>
cleanup_previous_terminal() {
  local pref_file="$1" new_terminal="$2"
  local old_terminal
  old_terminal="$(load_terminal_preference "$pref_file")"

  [ -z "$old_terminal" ] && return 0
  [ "$old_terminal" = "$new_terminal" ] && return 0

  (
    load_terminal_adapter "$old_terminal" >/dev/null 2>&1 || exit 0
    terminal_cleanup_config "$(terminal_get_config_path)"
  )
}
