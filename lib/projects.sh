#!/bin/bash
# Project file helpers — pure, no side effects on source.

# Extracts project name from a "name:path" line.
parse_project_name() {
  echo "${1%%:*}"
}

# Extracts project path from a "name:path" line.
# Uses non-greedy match so paths with colons work.
parse_project_path() {
  echo "${1#*:}"
}

# Reads a projects file and outputs valid lines (skips blanks and comments).
# Usage: mapfile -t projects < <(load_projects "$file")
load_projects() {
  local file="$1" line
  [ ! -f "$file" ] && return
  while IFS= read -r line; do
    [[ -z "$line" || "$line" == \#* ]] && continue
    echo "$line"
  done < "$file"
}

# Expands ~ to $HOME at the start of a path.
path_expand() {
  echo "${1/#\~/$HOME}"
}

# Truncates a path to max_width chars with ... in the middle.
path_truncate() {
  local path="$1" max_width="$2"
  if [ "${#path}" -le "$max_width" ]; then
    echo "$path"
  else
    local half=$(( (max_width - 3) / 2 ))
    echo "${path:0:$half}...${path: -$half}"
  fi
}
