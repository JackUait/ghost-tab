#!/bin/bash
# npm-based update check for ghost-tab.

# Show update-available notification if a previous background check found a newer version.
# Deletes the flag after displaying.
notify_if_update_available() {
  local config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  local flag="${config_home}/ghost-tab/update-available"
  [ -f "$flag" ] || return 0

  local version
  version="$(cat "$flag")"
  rm -f "$flag"
  echo "  ↑ Update available: v${version} — run 'npx ghost-tab' to update"
}

# Run a background check against the npm registry.
# If a newer version exists, writes a flag file for notify_if_update_available.
# Args: install_dir (where .version marker lives)
check_for_update() {
  local install_dir="$1"
  local config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  local flag="${config_home}/ghost-tab/update-available"

  # Need npm and a local version to compare
  command -v npm &>/dev/null || return 0
  local local_version
  local_version="$(cat "$install_dir/.version" 2>/dev/null | tr -d '[:space:]')"
  [ -n "$local_version" ] || return 0

  (
    local remote_version
    remote_version="$(npm view ghost-tab version 2>/dev/null | tr -d '[:space:]')" || return
    [ -n "$remote_version" ] || return
    [ "$local_version" = "$remote_version" ] && return

    mkdir -p "${config_home}/ghost-tab"
    echo "$remote_version" > "$flag"
  ) &
  disown
}
