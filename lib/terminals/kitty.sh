#!/bin/bash
# kitty terminal adapter.

# Return the path to kitty's config file.
terminal_get_config_path() {
  echo "$HOME/.config/kitty/kitty.conf"
}

# Return the path where the wrapper script should be.
terminal_get_wrapper_path() {
  echo "$HOME/.config/ghost-tab/wrapper.sh"
}

# Install kitty via Homebrew cask.
terminal_install() {
  ensure_cask "kitty" "kitty"
}

# Write or merge the wrapper command into kitty config.
# Args: config_path wrapper_path
terminal_setup_config() {
  local config_path="$1" wrapper_path="$2"
  local shell_line="shell $wrapper_path"

  if [ -f "$config_path" ] && grep -q '^shell[[:space:]]' "$config_path"; then
    sed -i '' 's|^shell[[:space:]].*|'"$shell_line"'|' "$config_path"
    success "Replaced existing shell line in config"
  else
    echo "$shell_line" >> "$config_path"
    success "Appended wrapper command to config"
  fi
}

# Remove ghost-tab shell line from kitty config.
terminal_cleanup_config() {
  local config_path="$1"
  if [ -f "$config_path" ]; then
    sed -i '' '/^shell[[:space:]]/d' "$config_path"
  fi
}
