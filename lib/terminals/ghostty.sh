#!/bin/bash
# Ghostty terminal adapter.

# Return the path to Ghostty's config file.
terminal_get_config_path() {
  echo "$HOME/.config/ghostty/config"
}

# Return the path where the wrapper script should be.
terminal_get_wrapper_path() {
  echo "$HOME/.config/ghost-tab/wrapper.sh"
}

# Install Ghostty: check for the app, open download page if missing.
terminal_install() {
  local app_path="${GHOSTTY_APP_PATH:-/Applications/Ghostty.app}"
  if [ -d "$app_path" ]; then
    success "Ghostty found"
    return 0
  fi

  info "Ghostty not found. Opening download page..."
  open "https://ghostty.org/download"
  echo ""
  echo "  Download and install Ghostty from the page that just opened."
  echo "  Press Enter when installation is complete."
  read -r < /dev/tty

  if [ ! -d "$app_path" ]; then
    error "Ghostty still not found at $app_path"
    info "Install Ghostty and re-run: ghost-tab --terminal"
    return 1
  fi
  success "Ghostty installed"
}

# Write or merge the wrapper command into Ghostty config.
# Args: config_path wrapper_path
terminal_setup_config() {
  local config_path="$1" wrapper_path="$2"
  local wrapper_line="command = $wrapper_path"

  if [ -f "$config_path" ] && grep -q '^command[[:space:]]*=' "$config_path"; then
    sed -i '' 's|^command[[:space:]]*=.*|'"$wrapper_line"'|' "$config_path"
    success "Replaced existing command line in config"
  else
    echo "$wrapper_line" >> "$config_path"
    success "Appended wrapper command to config"
  fi
}

# Remove ghost-tab command line from Ghostty config.
terminal_cleanup_config() {
  local config_path="$1"
  if [ -f "$config_path" ]; then
    sed -i '' '/^command[[:space:]]*=/d' "$config_path"
  fi
}
