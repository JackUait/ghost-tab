#!/bin/bash
# Tab title watcher — detects AI tool waiting state, updates terminal tab title.
# Depends on: tui.sh (set_tab_title, set_tab_title_waiting)

_TAB_TITLE_WATCHER_PID=""

# Check if the AI tool is waiting for user input.
# Usage: check_ai_tool_state <ai_tool> <session_name> <tmux_cmd> <marker_file> <pane_index>
# Outputs "waiting" or "active".
check_ai_tool_state() {
  local ai_tool="$1" session_name="$2" tmux_cmd="$3" marker_file="$4"
  local pane_index="${5:-3}"

  if [ "$ai_tool" = "claude" ]; then
    if [ -f "$marker_file" ]; then
      # Marker exists, but verify pane still shows a prompt.
      # Between user input and first tool call, the marker persists
      # even though Claude is actively working.
      local content last_line
      content=$("$tmux_cmd" capture-pane -t "$session_name:0.$pane_index" -p 2>/dev/null || true)
      last_line=$(echo "$content" | grep -v '^$' | tail -1)
      if echo "$last_line" | grep -qE '[>$❯]\s*$'; then
        echo "waiting"
      else
        echo "active"
      fi
    else
      echo "active"
    fi
  else
    local content last_line
    content=$("$tmux_cmd" capture-pane -t "$session_name:0.$pane_index" -p 2>/dev/null || true)
    last_line=$(echo "$content" | grep -v '^$' | tail -1)
    if echo "$last_line" | grep -qE '[>$❯]\s*$'; then
      echo "waiting"
    else
      echo "active"
    fi
  fi
}

# Start the tab title watcher background loop.
# Usage: start_tab_title_watcher <session_name> <ai_tool> <project_name> <tab_title_setting> <tmux_cmd> <marker_file> [pane_index]
start_tab_title_watcher() {
  local session_name="$1" ai_tool="$2" project_name="$3"
  local tab_title_setting="$4" tmux_cmd="$5" marker_file="$6"
  local pane_index="${7:-3}"

  (
    local was_waiting=false
    while true; do
      sleep 1.5
      local state
      state=$(check_ai_tool_state "$ai_tool" "$session_name" "$tmux_cmd" "$marker_file" "$pane_index")

      if [ "$state" = "waiting" ] && [ "$was_waiting" = false ]; then
        if [ "$tab_title_setting" = "full" ]; then
          set_tab_title_waiting "$project_name" "$ai_tool"
        else
          set_tab_title_waiting "$project_name"
        fi
        was_waiting=true
      elif [ "$state" = "active" ] && [ "$was_waiting" = true ]; then
        if [ "$tab_title_setting" = "full" ]; then
          set_tab_title "$project_name" "$ai_tool"
        else
          set_tab_title "$project_name"
        fi
        was_waiting=false
      fi
    done
  ) &
  _TAB_TITLE_WATCHER_PID=$!
}

# Stop the tab title watcher and clean up.
# Usage: stop_tab_title_watcher [marker_file]
stop_tab_title_watcher() {
  local marker_file="${1:-}"
  if [ -n "$_TAB_TITLE_WATCHER_PID" ]; then
    kill "$_TAB_TITLE_WATCHER_PID" 2>/dev/null || true
  fi
  if [ -n "$marker_file" ]; then
    rm -f "$marker_file"
  fi
}
