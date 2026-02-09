#!/bin/bash
# Notification setup — sound and tab spinner hooks.
# Depends on: tui.sh (success, warn), settings-json.sh (add_sound_notification_hook, add_spinner_hooks)

# Add sound notification hook to Claude settings.
# Usage: setup_sound_notification <settings_path> <sound_command>
setup_sound_notification() {
  local settings_path="$1" sound_command="$2"
  local result
  result="$(add_sound_notification_hook "$settings_path" "$sound_command")"
  if [ "$result" = "added" ]; then
    success "Sound notification configured"
  elif [ "$result" = "exists" ]; then
    success "Sound notification already configured"
  else
    warn "Failed to configure sound notification"
    return 1
  fi
}

# Copy spinner scripts and add hooks to Claude settings.
# Usage: setup_tab_spinner <share_dir> <settings_path> <start_cmd> <stop_cmd> <home_dir>
setup_tab_spinner() {
  local share_dir="$1" settings_path="$2" start_cmd="$3" stop_cmd="$4" home_dir="$5"

  mkdir -p "$home_dir/.claude"

  cp "$share_dir/templates/tab-spinner-start.sh" "$home_dir/.claude/tab-spinner-start.sh"
  cp "$share_dir/templates/tab-spinner-stop.sh" "$home_dir/.claude/tab-spinner-stop.sh"
  chmod +x "$home_dir/.claude/tab-spinner-start.sh"
  chmod +x "$home_dir/.claude/tab-spinner-stop.sh"
  success "Created spinner scripts"

  if add_spinner_hooks "$settings_path" "$start_cmd" "$stop_cmd"; then
    success "Tab animation hooks configured"
  else
    warn "Failed to configure tab animation hooks"
    return 1
  fi
}
