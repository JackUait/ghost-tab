#!/bin/bash
# iTerm2 terminal adapter using PlistBuddy.

# Return the path to iTerm2's preferences plist.
terminal_get_config_path() {
  echo "$HOME/Library/Preferences/com.googlecode.iterm2.plist"
}

# Return the path where the wrapper script should be.
terminal_get_wrapper_path() {
  echo "$HOME/.config/ghost-tab/wrapper.sh"
}

# Install iTerm2 via Homebrew cask.
terminal_install() {
  ensure_cask "iterm2" "iTerm"
}

# Create a "Ghost Tab" profile in iTerm2 that runs the wrapper.
# Args: plist_path wrapper_path
terminal_setup_config() {
  local plist_path="$1" wrapper_path="$2"

  local pb
  pb="$(command -v PlistBuddy 2>/dev/null || echo "/usr/libexec/PlistBuddy")"

  "$pb" -c "Add ':New Bookmarks:999:Name' string 'Ghost Tab'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Set ':New Bookmarks:999:Name' 'Ghost Tab'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Add ':New Bookmarks:999:Custom Command' string 'Yes'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Set ':New Bookmarks:999:Custom Command' 'Yes'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Add ':New Bookmarks:999:Command' string '$wrapper_path'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Set ':New Bookmarks:999:Command' '$wrapper_path'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Add ':New Bookmarks:999:Guid' string 'ghost-tab-profile'" "$plist_path" 2>/dev/null || true
  "$pb" -c "Set ':New Bookmarks:999:Guid' 'ghost-tab-profile'" "$plist_path" 2>/dev/null || true

  success "Created Ghost Tab profile in iTerm2"
  info "Set 'Ghost Tab' as your default profile in iTerm2 Preferences → Profiles"
}

# Remove the Ghost Tab profile from iTerm2.
terminal_cleanup_config() {
  local plist_path="$1"

  local pb
  pb="$(command -v PlistBuddy 2>/dev/null || echo "/usr/libexec/PlistBuddy")"

  "$pb" -c "Delete ':New Bookmarks:999'" "$plist_path" 2>/dev/null || true
  success "Removed Ghost Tab profile from iTerm2"
}
