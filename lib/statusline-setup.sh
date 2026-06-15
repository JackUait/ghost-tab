#!/bin/bash
# Statusline setup — install ccstatusline, copy configs and scripts.
# Depends on: tui.sh (success, warn, info), settings-json.sh (merge_claude_settings)

# Check whether npm is available. Extracted for testability.
_has_npm() { command -v npm &>/dev/null; }

# Minimum ccstatusline version. ccstatusline >= 2.2.x reads the model's real
# context_window_size from Claude Code's JSON, so the context percentage tracks
# the actual window (e.g. 1M for Opus [1m]) instead of a hardcoded 200K. Older
# installs are upgraded; newer installs are left untouched. Bump to raise the
# floor.
_CCSTATUSLINE_MIN_VERSION="2.2.21"

# Print the globally-installed ccstatusline version, or empty if not installed.
_ccstatusline_version() {
  npm list -g ccstatusline --depth=0 2>/dev/null \
    | sed -n 's/.*ccstatusline@\([0-9][0-9.]*\).*/\1/p' \
    | head -1
}

# True if version $1 is strictly older than version $2 (semver-aware via sort -V).
_version_lt() {
  [[ "$1" != "$2" ]] && [[ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -1)" == "$1" ]]
}

# Install and configure the Claude Code status line.
# Usage: setup_statusline <share_dir> <claude_settings_path> <home_dir>
setup_statusline() {
  local share_dir="$1" claude_settings_path="$2" home_dir="$3"

  # Check for npm, install Node.js LTS if needed
  if ! _has_npm; then
    info "Installing Node.js LTS..."
    if brew install node@22 &>/dev/null; then
      export PATH="/opt/homebrew/opt/node@22/bin:$PATH"
      success "Node.js LTS installed"
    else
      warn "Node.js installation failed — skipping status line setup"
      return 0
    fi
  fi

  if ! _has_npm; then
    return 0
  fi

  # Install or upgrade ccstatusline to at least the minimum version. A stale
  # install would keep reporting context % against a hardcoded 200K window, so
  # the floor ensures the percentage reflects the model's real context window.
  # Newer-than-minimum installs are left as-is (never downgraded).
  local installed_version
  installed_version="$(_ccstatusline_version)"

  if [[ -n "$installed_version" ]] && ! _version_lt "$installed_version" "$_CCSTATUSLINE_MIN_VERSION"; then
    success "ccstatusline already up to date ($installed_version)"
  else
    if [[ -n "$installed_version" ]]; then
      info "Updating ccstatusline ($installed_version -> $_CCSTATUSLINE_MIN_VERSION)..."
    else
      info "Installing ccstatusline..."
    fi
    if npm install -g "ccstatusline@$_CCSTATUSLINE_MIN_VERSION" &>/dev/null; then
      success "ccstatusline installed"
    else
      warn "Failed to install ccstatusline — skipping status line setup"
      return 0
    fi
  fi

  if npm list -g ccstatusline &>/dev/null; then
    # Create ccstatusline config
    mkdir -p "$home_dir/.config/ccstatusline"
    cp "$share_dir/templates/ccstatusline-settings.json" "$home_dir/.config/ccstatusline/settings.json"
    success "Created ccstatusline config"

    # Create statusline scripts
    mkdir -p "$home_dir/.claude"
    cp "$share_dir/templates/statusline-command.sh" "$home_dir/.claude/statusline-command.sh"
    cp "$share_dir/templates/statusline-wrapper.sh" "$home_dir/.claude/statusline-wrapper.sh"
    cp "$share_dir/lib/statusline.sh" "$home_dir/.claude/statusline-helpers.sh"
    chmod +x "$home_dir/.claude/statusline-command.sh"
    chmod +x "$home_dir/.claude/statusline-wrapper.sh"
    success "Created statusline scripts"

    # Update Claude settings.json
    merge_claude_settings "$claude_settings_path"
  fi
}
