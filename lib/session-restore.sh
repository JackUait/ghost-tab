#!/bin/bash
# Session restore — snapshot alive Ghost Tab tmux sessions and reopen them
# after a reboot. Depends on: terminals/adapter.sh (load_terminal_adapter).

# Print the current macOS boot id (the kern.boottime sec value).
# Stable for one uptime; changes on every reboot. Empty on failure.
current_boot_id() {
  local out
  out="$(sysctl -n kern.boottime 2>/dev/null)" || return 0
  echo "$out" | sed -n 's/.*[^u]sec = \([0-9][0-9]*\).*/\1/p'
}
