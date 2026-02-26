#!/bin/bash
# Auto-update for ghost-tab — git-based (native install) and Homebrew.

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

# check_for_update [share_dir]
#
# With share_dir: run a background git fetch + pull in share_dir.
#   If a new version is pulled, downloads the ghost-tab-tui binary and writes
#   a flag file at $XDG_CONFIG_HOME/ghost-tab/updated.
#   Does nothing if share_dir is not a git repo.
#
# Without share_dir: Homebrew-based check (legacy, used when installed via brew).
#   Reads/writes UPDATE_CACHE; sets _update_version if a newer version exists.
check_for_update() {
  if [ -n "${1:-}" ]; then
    _check_for_update_git "$1"
  else
    _check_for_update_brew
  fi
}

_check_for_update_git() {
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

_check_for_update_brew() {
  local cache_ts now age latest
  # Only check if brew is available (Homebrew install)
  command -v brew &>/dev/null || return 0

  # Read cache if it exists
  if [ -f "$UPDATE_CACHE" ]; then
    latest="$(sed -n '1p' "$UPDATE_CACHE")"
    cache_ts="$(sed -n '2p' "$UPDATE_CACHE")"
    now="$(date +%s)"
    age=$(( now - ${cache_ts:-0} ))
    # Use cached result if less than 24 hours old
    if [ "$age" -lt 86400 ]; then
      # Verify cached version is actually newer than installed
      if [ -n "$latest" ]; then
        installed="$(brew list --versions ghost-tab 2>/dev/null | awk '{print $2}')"
        if [ "$latest" != "$installed" ]; then
          _update_version="$latest"
        fi
      fi
      return
    fi
  fi

  # Spawn background check (non-blocking)
  (
    result="$(brew outdated --verbose --formula ghost-tab 2>/dev/null)"
    mkdir -p "$(dirname "$UPDATE_CACHE")"
    if [ -n "$result" ]; then
      # Extract new version: "ghost-tab (1.0.0) < 1.1.0" -> "1.1.0"
      new_ver="$(echo "$result" | sed -n 's/.*< *//p')"
      printf '%s\n%s\n' "$new_ver" "$(date +%s)" > "$UPDATE_CACHE.tmp"
      mv "$UPDATE_CACHE.tmp" "$UPDATE_CACHE"
    else
      printf '\n%s\n' "$(date +%s)" > "$UPDATE_CACHE.tmp"
      mv "$UPDATE_CACHE.tmp" "$UPDATE_CACHE"
    fi
  ) &
  disown
}
