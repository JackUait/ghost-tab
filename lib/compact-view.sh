#!/bin/bash
# shellcheck disable=SC2059  # Intentional: ANSI escape variables in printf format strings
# Compact view: show changed files with +/- line counts instead of lazygit.
# Refreshes every 2 seconds. Ctrl-C to exit.

compact_view() {
  local project_dir="${1:-.}"

  # Need to be in a git repo
  if ! git -C "$project_dir" rev-parse --is-inside-work-tree &>/dev/null; then
    echo "Not a git repository"
    exec "$SHELL"
    return
  fi

  trap 'printf "\033[?25h\033[0m"; exit 0' INT TERM
  printf "\033[?25l" # hide cursor

  # ANSI helpers
  local dim="\033[90m"
  local bold="\033[1m"
  local cyan="\033[36m"
  local green="\033[32m"
  local red="\033[31m"
  local yellow="\033[33m"
  local bright="\033[97m"
  local reset="\033[0m"
  local dimline="\033[2m"

  # Show parent_dir/filename, truncating parent dir if needed.
  # Usage: format_file <path> <max_width>
  format_file() {
    local p="$1" max="$2"
    local fname="${p##*/}"
    local dir="${p%/*}"
    # Get just the immediate parent directory name
    local parent="${dir##*/}"
    local short="${parent}/${fname}"
    if [ ${#short} -le "$max" ]; then
      printf '%s' "$short"
    elif [ ${#fname} -ge "$max" ]; then
      # Filename alone exceeds max
      printf '%.*s…' "$((max - 1))" "$fname"
    else
      # Truncate parent dir to fit
      local keep=$((max - ${#fname} - 1))  # 1 for "/"
      printf '%.*s/%s' "$keep" "$parent" "$fname"
    fi
  }

  while true; do
    # Capture pane width outside subshell.
    # tput cols may return wrong value in tmux; query tmux directly.
    local w
    if [ -n "${TMUX:-}" ] && command -v tmux &>/dev/null; then
      w=$(tmux display-message -p '#{pane_width}' 2>/dev/null || tput cols 2>/dev/null || echo 80)
    else
      w=$(tput cols 2>/dev/null || echo 80)
    fi

    output=$(
      cd "$project_dir" || exit 1

      # Inner content width (2-space padding each side)
      local iw=$((w - 4))
      [ "$iw" -lt 20 ] && iw=20

      # Branch + ahead/behind
      local branch ahead_behind=""
      branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "detached")
      if git rev-parse '@{u}' &>/dev/null 2>&1; then
        local counts ahead behind
        counts=$(git rev-list --left-right --count "HEAD...@{u}" 2>/dev/null)
        if [ -n "$counts" ]; then
          ahead=$(echo "$counts" | cut -f1)
          behind=$(echo "$counts" | cut -f2)
          [ "$ahead" -gt 0 ] && ahead_behind=" ${cyan}↑${ahead}${reset}"
          [ "$behind" -gt 0 ] && ahead_behind="${ahead_behind} ${yellow}↓${behind}${reset}"
        fi
      fi

      # Gather data
      local staged unstaged untracked
      staged=$(git diff --cached --numstat 2>/dev/null)
      unstaged=$(git diff --numstat 2>/dev/null)
      untracked=$(git ls-files --others --exclude-standard 2>/dev/null)

      # Count totals
      local n_staged=0 n_unstaged=0 n_untracked=0
      [ -n "$staged" ] && n_staged=$(echo "$staged" | wc -l | tr -d ' ')
      [ -n "$unstaged" ] && n_unstaged=$(echo "$unstaged" | wc -l | tr -d ' ')
      [ -n "$untracked" ] && n_untracked=$(echo "$untracked" | wc -l | tr -d ' ')

      # Clear screen
      printf "\033[2J\033[H"

      # ── Header ──
      printf " ${bold}${bright}%s${reset}" "$branch"
      [ -n "$ahead_behind" ] && printf "%s" "$ahead_behind"
      printf "\n"

      # Separator line
      printf " ${dimline}"
      printf '%.*s' "$iw" '─'
      printf "${reset}\n"

      local has_content=0
      # Available width for file names after "+NNN -NNN  " prefix (13 chars)
      local file_width=$((iw - 13))
      [ "$file_width" -lt 10 ] && file_width=10

      # ── Staged ──
      if [ -n "$staged" ]; then
        printf " ${green}${bold}●${reset} ${green}staged${reset}  ${dim}(%s)${reset}\n" "$n_staged"
        while IFS=$'\t' read -r added deleted file; do
          has_content=1
          local display
          display=$(format_file "$file" "$file_width")
          printf "   ${green}%4s${reset} ${dim}-${reset}${red}%4s${reset}  %s\n" "+${added}" "${deleted}" "$display"
        done <<< "$staged"
        printf "\n"
      fi

      # ── Modified ──
      if [ -n "$unstaged" ]; then
        printf " ${yellow}${bold}●${reset} ${yellow}modified${reset}  ${dim}(%s)${reset}\n" "$n_unstaged"
        while IFS=$'\t' read -r added deleted file; do
          has_content=1
          local display
          display=$(format_file "$file" "$file_width")
          printf "   ${green}%4s${reset} ${dim}-${reset}${red}%4s${reset}  %s\n" "+${added}" "${deleted}" "$display"
        done <<< "$unstaged"
        printf "\n"
      fi

      # ── Untracked ──
      if [ -n "$untracked" ]; then
        printf " ${dim}${bold}○${reset} ${dim}new${reset}  ${dim}(%s)${reset}\n" "$n_untracked"
        while IFS= read -r file; do
          has_content=1
          local display
          display=$(format_file "$file" "$file_width")
          printf "   ${dim}%s${reset}\n" "$display"
        done <<< "$untracked"
        printf "\n"
      fi

      # Empty state
      if [ "$has_content" -eq 0 ]; then
        printf " ${dim}no changes${reset}\n\n"
      fi

      # ── Summary bar ──
      printf " ${dimline}"
      printf '%.*s' "$iw" '─'
      printf "${reset}\n"
      printf " "
      [ "$n_staged" -gt 0 ] && printf " ${green}${n_staged} staged${reset}"
      [ "$n_unstaged" -gt 0 ] && printf " ${yellow}${n_unstaged} mod${reset}"
      [ "$n_untracked" -gt 0 ] && printf " ${dim}${n_untracked} new${reset}"
      [ "$has_content" -eq 0 ] && printf " ${dim}clean${reset}"
      printf "\n"
    )

    printf "%s" "$output"
    sleep 2
  done
}
