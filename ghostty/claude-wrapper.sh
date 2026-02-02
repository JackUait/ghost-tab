#!/bin/bash
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"

TMUX_CMD="$(command -v tmux)"
LAZYGIT_CMD="$(command -v lazygit)"
BROOT_CMD="$(command -v broot)"
CLAUDE_CMD="$(command -v claude)"

# Load user projects from config file if it exists
PROJECTS_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/vibecode-editor/projects"

# Select working directory
if [ -n "$1" ] && [ -d "$1" ]; then
  cd "$1"
  shift
elif [ -z "$1" ]; then
  projects=()
  if [ -f "$PROJECTS_FILE" ]; then
    while IFS= read -r line; do
      [[ -z "$line" || "$line" == \#* ]] && continue
      projects+=("$line")
    done < "$PROJECTS_FILE"
  fi

  if [ ${#projects[@]} -gt 0 ]; then
    echo "Select project:"
    for i in "${!projects[@]}"; do
      name="${projects[$i]%%:*}"
      printf "  %d) %s\n" $((i+1)) "$name"
    done
    printf "  0) current directory\n"
    read -rn1 -p "> " choice
    echo
    if [[ "$choice" =~ ^[1-9][0-9]*$ ]] && [ "$choice" -le "${#projects[@]}" ]; then
      dir="${projects[$((choice-1))]#*:}"
      cd "$dir"
    fi
  fi
fi

export PROJECT_DIR="$(pwd)"
SESSION_NAME="dev-$(basename "$PROJECT_DIR")-$$"

# Background watcher: switch to Claude pane once it's ready
(
  while true; do
    sleep 0.5
    content=$("$TMUX_CMD" capture-pane -t "$SESSION_NAME:0.1" -p 2>/dev/null)
    if echo "$content" | grep -q '>'; then
      "$TMUX_CMD" select-pane -t "$SESSION_NAME:0.1"
      break
    fi
  done
) &
WATCHER_PID=$!

# Kill all processes in the tmux session when the terminal closes
cleanup() {
  # Get PIDs of all pane shell processes in this session, then kill their entire process trees
  for pane_pid in $("$TMUX_CMD" list-panes -s -t "$SESSION_NAME" -F '#{pane_pid}' 2>/dev/null); do
    pkill -TERM -P "$pane_pid" 2>/dev/null
  done
  kill $WATCHER_PID 2>/dev/null
  "$TMUX_CMD" kill-session -t "$SESSION_NAME" 2>/dev/null
}
trap cleanup EXIT HUP TERM INT

"$TMUX_CMD" new-session -s "$SESSION_NAME" -e "PATH=$PATH" -c "$PROJECT_DIR" \
  "$LAZYGIT_CMD; exec bash" \; \
  split-window -h -p 50 -c "$PROJECT_DIR" \
  "$CLAUDE_CMD $*; exec bash" \; \
  select-pane -t 0 \; \
  split-window -v -p 50 -c "$PROJECT_DIR" \
  "while true; do $BROOT_CMD $PROJECT_DIR; done" \; \
  split-window -v -p 30 -c "$PROJECT_DIR" \; \
  select-pane -t 3
