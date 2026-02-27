#!/bin/bash
# Claude settings.json manipulation helpers.

# Merge statusLine into Claude settings.json (create if missing).
merge_claude_settings() {
  local path="$1"
  mkdir -p "$(dirname "$path")"
  if [ -f "$path" ]; then
    if grep -q '"statusLine"' "$path"; then
      success "Claude status line already configured"
    else
      sed -i '' '$ s/}$/,\n  "statusLine": {\n    "type": "command",\n    "command": "bash ~\/.claude\/statusline-wrapper.sh"\n  }\n}/' "$path"
      success "Added status line to Claude settings"
    fi
  else
    cat > "$path" << 'CSEOF'
{
  "statusLine": {
    "type": "command",
    "command": "bash ~/.claude/statusline-wrapper.sh"
  }
}
CSEOF
    success "Created Claude settings with status line"
  fi
}

# Add waiting indicator hooks (Stop + PreToolUse + UserPromptSubmit) to settings.json.
# Uses $GHOST_TAB_MARKER_FILE env var so hooks are safe outside Ghost Tab.
# Outputs "added", "upgraded", or "exists".
add_waiting_indicator_hooks() {
  local path="$1"
  mkdir -p "$(dirname "$path")"
  python3 - "$path" << 'PYEOF'
import json, sys, os

settings_path = sys.argv[1]

if os.path.exists(settings_path):
    try:
        with open(settings_path, "r") as f:
            settings = json.load(f)
    except (json.JSONDecodeError, ValueError):
        settings = {}
else:
    settings = {}

hooks = settings.setdefault("hooks", {})

stop_cmd = 'if [ -n "$GHOST_TAB_MARKER_FILE" ]; then touch "$GHOST_TAB_MARKER_FILE"; fi'
clear_cmd = 'if [ -n "$GHOST_TAB_MARKER_FILE" ]; then rm -f "$GHOST_TAB_MARKER_FILE"; fi'

marker = "GHOST_TAB_MARKER_FILE"

# Check if current Stop-based format is already installed
stop_list = hooks.get("Stop", [])
stop_exists = any(
    marker in h.get("command", "")
    for entry in stop_list
    for h in entry.get("hooks", [])
)

# Check if old Notification-based format exists (needs upgrade)
notif_list = hooks.get("Notification", [])
notif_exists = any(
    marker in h.get("command", "")
    for entry in notif_list
    for h in entry.get("hooks", [])
)

# Check if old Stop format without matcher exists (needs upgrade)
pre_list = hooks.get("PreToolUse", [])
old_stop_needs_upgrade = stop_exists and not any(
    entry.get("matcher") == "AskUserQuestion"
    for entry in pre_list
    if any(marker in h.get("command", "") for h in entry.get("hooks", []))
)

if stop_exists and not old_stop_needs_upgrade:
    # Current Stop format already installed
    print("exists")
    sys.exit(0)
elif notif_exists or old_stop_needs_upgrade:
    # Old format — remove ghost-tab hooks so they get re-added below
    for event in ["Stop", "Notification", "PreToolUse", "UserPromptSubmit"]:
        event_list = hooks.get(event, [])
        new_list = [
            entry for entry in event_list
            if not any(marker in h.get("command", "") for h in entry.get("hooks", []))
        ]
        if new_list:
            hooks[event] = new_list
        elif event in hooks:
            del hooks[event]
    action = "upgraded"
else:
    action = "added"

# Add Stop hook (fires immediately when Claude stops generating)
hooks.setdefault("Stop", []).append({
    "hooks": [{"type": "command", "command": stop_cmd}]
})

# Add PreToolUse hook with matcher for AskUserQuestion (creates marker — user input needed)
hooks.setdefault("PreToolUse", []).append({
    "matcher": "AskUserQuestion",
    "hooks": [{"type": "command", "command": stop_cmd}]
})

# Add PreToolUse catch-all hook (clears marker — Claude is actively working)
hooks.setdefault("PreToolUse", []).append({
    "hooks": [{"type": "command", "command": clear_cmd}]
})

# Add UserPromptSubmit hook (clears marker when user answers)
hooks.setdefault("UserPromptSubmit", []).append({
    "hooks": [{"type": "command", "command": clear_cmd}]
})

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print(action)
PYEOF
}

# Remove waiting indicator hooks from settings.json.
# Outputs "removed" or "not_found".
remove_waiting_indicator_hooks() {
  local path="$1"
  if [ ! -f "$path" ]; then
    echo "not_found"
    return 0
  fi
  python3 - "$path" << 'PYEOF'
import json, sys, os

settings_path = sys.argv[1]
marker = "GHOST_TAB_MARKER_FILE"

try:
    with open(settings_path, "r") as f:
        settings = json.load(f)
except (json.JSONDecodeError, ValueError, FileNotFoundError):
    print("not_found")
    sys.exit(0)

hooks = settings.get("hooks", {})
found = False

for event in ["Stop", "Notification", "PreToolUse", "UserPromptSubmit"]:
    event_list = hooks.get(event, [])
    new_list = [
        entry for entry in event_list
        if not any(marker in h.get("command", "") for h in entry.get("hooks", []))
    ]
    if len(new_list) != len(event_list):
        found = True
        if new_list:
            hooks[event] = new_list
        else:
            del hooks[event]

if not found:
    print("not_found")
    sys.exit(0)

if not hooks:
    del settings["hooks"]

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print("removed")
PYEOF
}
