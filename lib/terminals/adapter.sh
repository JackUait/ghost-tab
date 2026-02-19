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
