#!/bin/bash
# shellcheck source=../lib/statusline.sh
source "$(dirname "$0")/../lib/statusline.sh" 2>/dev/null \
  || source ~/.claude/statusline-helpers.sh 2>/dev/null \
  || true

input=$(cat)
git_info=$(echo "$input" | bash ~/.claude/statusline-command.sh)
context_pct=$(echo "$input" | npx ccstatusline 2>/dev/null)
model_name=$(echo "$input" | sed -n 's/.*"display_name":"\([^"]*\)".*/\1/p')

# Find parent Claude Code process and get total tree memory usage
pid=$PPID
mem_label=""
while [ -n "$pid" ] && [ "$pid" != "1" ]; do
  comm=$(ps -o comm= -p "$pid" 2>/dev/null | xargs basename 2>/dev/null)
  if [ "$comm" = "claude" ]; then
    if type get_tree_rss_kb &>/dev/null; then
      mem_kb=$(get_tree_rss_kb "$pid")
    else
      mem_kb=$(ps -o rss= -p "$pid" 2>/dev/null | tr -d ' ')
    fi
    if [ -n "$mem_kb" ] && [ "$mem_kb" -gt 0 ] 2>/dev/null; then
      mem_mb=$((mem_kb / 1024))
      if [ "$mem_mb" -ge 1024 ]; then
        mem_gb=$(echo "scale=1; $mem_mb / 1024" | bc)
        mem_label="${mem_gb}G"
      else
        mem_label="${mem_mb}M"
      fi
    fi
    break
  fi
  pid=$(ps -o ppid= -p "$pid" 2>/dev/null | tr -d ' ')
done

line=$(printf '%s | %s' "$git_info" "$context_pct")
if [ -n "$mem_label" ]; then
  line="$line$(printf ' | \033[01;35m%s\033[00m' "$mem_label")"
fi
if [ -n "$model_name" ]; then
  line="$line$(printf ' | \033[01;34m%s\033[00m' "$model_name")"
fi
printf '%s' "$line"
