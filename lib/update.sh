#!/bin/bash
# Git-based auto-update for ghost-tab.

# Show update notification if a previous background update wrote a flag file.
# Deletes the flag after displaying.
notify_if_updated() {
  local config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  local flag="${config_home}/ghost-tab/updated"
  [ -f "$flag" ] || return 0

  local version
  version="$(cat "$flag")"
  rm -f "$flag"
  echo "  ↑ Updated to v${version}"
}

# Run a background git fetch + pull in share_dir.
# If a new version is pulled, downloads the ghost-tab-tui binary and writes a flag file.
# Args: share_dir
check_for_update() {
  local share_dir="$1"
  local config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  local flag="${config_home}/ghost-tab/updated"

  # Only works if share_dir is a git repo
  [ -d "$share_dir/.git" ] || return 0

  (
    local local_ref remote_ref
    git -C "$share_dir" fetch origin main --quiet 2>/dev/null || return

    local_ref="$(git -C "$share_dir" rev-parse HEAD 2>/dev/null)"
    remote_ref="$(git -C "$share_dir" rev-parse origin/main 2>/dev/null)"
    [ "$local_ref" = "$remote_ref" ] && return

    git -C "$share_dir" pull --rebase --quiet origin main 2>/dev/null || return

    local new_version arch
    new_version="$(tr -d '[:space:]' < "$share_dir/VERSION" 2>/dev/null)" || return
    [ -n "$new_version" ] || return

    arch="$(uname -m)"
    local bin_url="https://github.com/JackUait/ghost-tab/releases/download/v${new_version}/ghost-tab-tui-darwin-${arch}"
    mkdir -p "$HOME/.local/bin"
    curl -fsSL -o "$HOME/.local/bin/ghost-tab-tui" "$bin_url" 2>/dev/null && \
      chmod +x "$HOME/.local/bin/ghost-tab-tui" || true

    mkdir -p "${config_home}/ghost-tab"
    echo "$new_version" > "$flag"
  ) &
  disown
}
