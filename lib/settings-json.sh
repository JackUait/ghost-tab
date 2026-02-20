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

# Add waiting indicator hooks (Stop + PreToolUse) to settings.json.
# Uses $GHOST_TAB_MARKER_FILE env var so hooks are safe outside Ghost Tab.
# Outputs "added" or "exists".
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
pre_tool_cmd = '_gt_in=$(cat); if [ -n "$GHOST_TAB_MARKER_FILE" ]; then if [[ "$_gt_in" == *AskUserQuestion* ]]; then touch "$GHOST_TAB_MARKER_FILE"; else rm -f "$GHOST_TAB_MARKER_FILE"; fi; fi'

# Check if already installed
stop_list = hooks.get("Stop", [])
already_exists = any(
    "GHOST_TAB_MARKER_FILE" in h.get("command", "")
    for entry in stop_list
    for h in entry.get("hooks", [])
)

if already_exists:
    print("exists")
    sys.exit(0)

# Add Stop hook
hooks.setdefault("Stop", []).append({
    "hooks": [{"type": "command", "command": stop_cmd}]
})

# Add PreToolUse hook (conditional: AskUserQuestion creates marker, others clear it)
hooks.setdefault("PreToolUse", []).append({
    "hooks": [{"type": "command", "command": pre_tool_cmd}]
})

# Add UserPromptSubmit hook (clears marker when user answers)
hooks.setdefault("UserPromptSubmit", []).append({
    "hooks": [{"type": "command", "command": clear_cmd}]
})

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print("added")
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

for event in ["Stop", "PreToolUse", "UserPromptSubmit"]:
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
