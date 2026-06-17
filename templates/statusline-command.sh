#!/bin/bash
input=$(cat)
cwd=$(echo "$input" | sed -n 's/.*"current_dir":"\([^"]*\)".*/\1/p')

if git -C "$cwd" rev-parse --git-dir > /dev/null 2>&1; then
  repo_name=$(basename "$cwd")
  branch=$(git -C "$cwd" --no-optional-locks rev-parse --abbrev-ref HEAD 2>/dev/null)
  printf '\033[01;36m%s\033[00m | \033[01;32m%s\033[00m' \
    "$repo_name" "$branch"
else
  printf '\033[01;36m%s\033[00m' "$(basename "$cwd")"
fi
