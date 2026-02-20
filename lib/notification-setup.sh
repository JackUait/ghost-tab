#!/bin/bash
# Notification setup — sound hooks.
# Depends on: tui.sh (success, warn), settings-json.sh (add_sound_notification_hook)

# Play notification sound if enabled for the given AI tool.
# Reads sound preference from features JSON and plays via afplay in background.
# Usage: play_notification_sound <ai_tool> <config_dir>
play_notification_sound() {
  local ai_tool="$1" config_dir="$2"
  local sound_name
  sound_name="$(get_sound_name "$ai_tool" "$config_dir")"
  if [[ -n "$sound_name" ]]; then
    afplay "/System/Library/Sounds/${sound_name}.aiff" &
  fi
}

# Add sound notification hook to Claude settings.
# Usage: setup_sound_notification <settings_path> <sound_command> [config_dir]
setup_sound_notification() {
  local settings_path="$1" sound_command="$2" config_dir="${3:-}"
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
  if [[ -n "$config_dir" ]]; then
    set_claude_notif_channel "$config_dir"
  fi
}

# Set Claude Code's preferredNotifChannel to terminal_bell to prevent
# double sounds (ghost-tab hook + built-in notification).
# Saves the previous value to <config_dir>/prev-notif-channel for restoration.
# Usage: set_claude_notif_channel <config_dir>
set_claude_notif_channel() {
  local config_dir="$1"
  if ! command -v claude &>/dev/null; then
    return 0
  fi
  mkdir -p "$config_dir"
  local prev
  prev="$(CLAUDECODE="" claude config get preferredNotifChannel 2>/dev/null || true)"
  echo "$prev" > "$config_dir/prev-notif-channel"
  CLAUDECODE="" claude config set preferredNotifChannel terminal_bell 2>/dev/null || true
}

# Restore Claude Code's preferredNotifChannel from saved value.
# If no saved value exists, does nothing.
# Usage: restore_claude_notif_channel <config_dir>
restore_claude_notif_channel() {
  local config_dir="$1"
  local saved_file="$config_dir/prev-notif-channel"
  if [ ! -f "$saved_file" ]; then
    return 0
  fi
  if ! command -v claude &>/dev/null; then
    return 0
  fi
  local prev
  prev="$(tr -d '[:space:]' < "$saved_file")"
  if [[ -n "$prev" ]]; then
    CLAUDECODE="" claude config set preferredNotifChannel "$prev" 2>/dev/null || true
  else
    CLAUDECODE="" claude config set preferredNotifChannel "" 2>/dev/null || true
  fi
  rm -f "$saved_file"
}

# Check if sound notifications are enabled for the given AI tool.
# Usage: is_sound_enabled <tool> <config_dir>
# Outputs "true" or "false".
is_sound_enabled() {
  local tool="$1" config_dir="$2"
  local features_file="$config_dir/${tool}-features.json"
  if [ -f "$features_file" ]; then
    local val
    val="$(python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    print('false' if d.get('sound') is False else 'true')
except Exception:
    print('true')
" "$features_file" 2>/dev/null)"
    echo "${val:-true}"
  else
    echo "true"
  fi
}

# Get the sound name for the given AI tool.
# Returns the sound name (e.g. "Bottle") or empty string if sound is disabled.
# Usage: get_sound_name <tool> <config_dir>
get_sound_name() {
  local tool="$1" config_dir="$2"
  local features_file="$config_dir/${tool}-features.json"
  if [ -f "$features_file" ]; then
    python3 -c "
import json, sys
try:
    d = json.load(open(sys.argv[1]))
    if d.get('sound') is False:
        print('')
    else:
        print(d.get('sound_name', 'Bottle'))
except Exception:
    print('Bottle')
" "$features_file" 2>/dev/null
  else
    echo "Bottle"
  fi
}

# Set the sound name for the given AI tool.
# Usage: set_sound_name <tool> <config_dir> <name>
set_sound_name() {
  local tool="$1" config_dir="$2" name="$3"
  local features_file="$config_dir/${tool}-features.json"
  mkdir -p "$config_dir"
  python3 -c "
import json, sys
path = sys.argv[1]
name = sys.argv[2]
try:
    d = json.load(open(path))
except Exception:
    d = {}
d['sound_name'] = name
with open(path, 'w') as f:
    json.dump(d, f)
    f.write('\n')
" "$features_file" "$name"
}

# Set sound feature flag for the given AI tool.
# Usage: set_sound_feature_flag <tool> <config_dir> <true|false>
set_sound_feature_flag() {
  local tool="$1" config_dir="$2" enabled="$3"
  local features_file="$config_dir/${tool}-features.json"
  mkdir -p "$config_dir"
  python3 -c "
import json, sys, os
path = sys.argv[1]
enabled = sys.argv[2] == 'true'
try:
    d = json.load(open(path))
except Exception:
    d = {}
d['sound'] = enabled
with open(path, 'w') as f:
    json.dump(d, f)
    f.write('\n')
" "$features_file" "$enabled"
}

# Remove sound notification hook from Claude settings.
# Usage: remove_sound_notification <settings_path> <sound_command> [config_dir]
remove_sound_notification() {
  local settings_path="$1" sound_command="$2" config_dir="${3:-}"
  local result
  result="$(remove_sound_notification_hook "$settings_path" "$sound_command")"
  echo "$result"
  if [[ -n "$config_dir" ]]; then
    restore_claude_notif_channel "$config_dir"
  fi
}

# Toggle sound notification for the given AI tool.
# Usage: toggle_sound_notification <tool> <config_dir> <settings_path>
# Reads current state, flips it, applies the change.
toggle_sound_notification() {
  local tool="$1" config_dir="$2" settings_path="$3"
  local current
  current="$(is_sound_enabled "$tool" "$config_dir")"
  local sound_command="afplay /System/Library/Sounds/Bottle.aiff &"

  if [[ "$current" == "true" ]]; then
    # Disable
    set_sound_feature_flag "$tool" "$config_dir" false
    case "$tool" in
      claude)
        remove_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications disabled"
  else
    # Enable
    set_sound_feature_flag "$tool" "$config_dir" true
    case "$tool" in
      claude)
        setup_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications enabled"
  fi
}

# Apply sound notification state for the given AI tool.
# Usage: apply_sound_notification <tool> <config_dir> <settings_path> <sound_name>
# If sound_name is empty, disables sound. Otherwise enables with that sound.
apply_sound_notification() {
  local tool="$1" config_dir="$2" settings_path="$3" sound_name="$4"

  if [[ -z "$sound_name" ]]; then
    # Disable sound
    set_sound_feature_flag "$tool" "$config_dir" false
    # Remove any existing afplay hook
    case "$tool" in
      claude)
        remove_sound_notification "$settings_path" "afplay /System/Library/Sounds/" "$config_dir"
        ;;
    esac
    success "Sound notifications disabled"
  else
    # Enable sound with specific name
    set_sound_feature_flag "$tool" "$config_dir" true
    set_sound_name "$tool" "$config_dir" "$sound_name"
    local sound_command="afplay /System/Library/Sounds/${sound_name}.aiff &"
    case "$tool" in
      claude)
        # Remove old hook first (any afplay sound), then add new one
        remove_sound_notification "$settings_path" "afplay /System/Library/Sounds/" "$config_dir"
        setup_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications enabled"
  fi
}
