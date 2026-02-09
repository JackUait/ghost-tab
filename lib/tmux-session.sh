#!/bin/bash
# Tmux session helpers — build launch command, cleanup.
# Depends on: process.sh (kill_tree)

# Build the AI tool launch command string.
# Usage: build_ai_launch_cmd <tool> <claude_cmd> <codex_cmd> <copilot_cmd> <opencode_cmd> [extra_args_or_project_dir]
build_ai_launch_cmd() {
  local tool="$1" claude_cmd="$2" codex_cmd="$3" copilot_cmd="$4" opencode_cmd="$5"
  shift 5
  local extra="$*"

  case "$tool" in
    codex)
      echo "$codex_cmd --cd \"$extra\""
      ;;
    copilot)
      echo "$copilot_cmd"
      ;;
    opencode)
      echo "$opencode_cmd \"$extra\""
      ;;
    *)
      if [ -n "$extra" ]; then
        echo "$claude_cmd $extra"
      else
        echo "$claude_cmd"
      fi
      ;;
  esac
}

# Clean up a tmux session: kill watcher, TERM pane trees, KILL survivors, destroy session.
cleanup_tmux_session() {
  local session_name="$1" watcher_pid="$2" tmux_cmd="$3"

  kill "$watcher_pid" 2>/dev/null || true

  local pane_pid
  for pane_pid in $("$tmux_cmd" list-panes -s -t "$session_name" -F '#{pane_pid}' 2>/dev/null); do
    kill_tree "$pane_pid" TERM
  done

  sleep 0.3
  for pane_pid in $("$tmux_cmd" list-panes -s -t "$session_name" -F '#{pane_pid}' 2>/dev/null); do
    kill_tree "$pane_pid" KILL
  done

  "$tmux_cmd" kill-session -t "$session_name" 2>/dev/null || true
}
