#!/bin/bash
# Screenshot helpers — locate the most recent screenshot so it can be injected
# into the AI pane.
#
# Why this exists: tmux delivers a drag-and-drop's paste to the *active* pane,
# not the pane the cursor is over (an external file drag never produces a tmux
# mouse event, so tmux cannot know the target pane). In ghost-tab's multi-pane
# layout the active pane is often lazygit or the spare shell, so a screenshot
# dropped onto the Claude pane lands elsewhere and Claude shows nothing. This
# lets a tmux binding inject the latest screenshot straight into the AI pane,
# bypassing drop routing entirely.

# gt_screenshot_dir — print the directory macOS saves screenshots to.
# Honors `com.apple.screencapture location`; falls back to ~/Desktop.
gt_screenshot_dir() {
  local loc
  loc="$(defaults read com.apple.screencapture location 2>/dev/null || true)"
  # Expand a leading ~ to $HOME.
  loc="${loc/#\~/$HOME}"
  if [ -n "$loc" ] && [ -d "$loc" ]; then
    printf '%s\n' "$loc"
  else
    printf '%s\n' "$HOME/Desktop"
  fi
}

# gt_latest_screenshot <dir> — print the newest image file in <dir>.
# Returns non-zero (printing nothing) when the dir is missing or has no images.
# Uses find+stat (not globbing) so it is robust across bash/zsh and when some
# image extensions have no matches.
gt_latest_screenshot() {
  local dir="$1"
  [ -d "$dir" ] || return 1

  local latest=""
  local line
  # Newest-first: macOS `stat -f '%m %N'` prints "<mtime-seconds> <path>".
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    latest="${line#* }"  # strip the leading "<mtime> "
    break
  done < <(find "$dir" -maxdepth 1 -type f \
            \( -iname '*.png' -o -iname '*.jpg' -o -iname '*.jpeg' \) \
            -exec stat -f '%m %N' {} + 2>/dev/null | sort -rn)

  [ -n "$latest" ] || return 1
  printf '%s\n' "$latest"
}

# gt_paste_latest_screenshot <session> [pane] — inject the latest screenshot's
# path into the AI pane as a bracketed paste (so Claude attaches it as an image).
# Defaults to pane index 1 (the AI pane in ghost-tab's layout).
gt_paste_latest_screenshot() {
  local tmux_cmd
  tmux_cmd="$(command -v tmux)" || return 1
  # Default to the session the binding fired in, and the AI pane (index 1).
  local session="${1:-$("$tmux_cmd" display-message -p '#{session_name}' 2>/dev/null)}"
  local pane="${2:-1}"
  [ -n "$session" ] || return 1

  local dir latest
  dir="$(gt_screenshot_dir)"
  latest="$(gt_latest_screenshot "$dir")" || {
    "$tmux_cmd" display-message "ghost-tab: no screenshot found in $dir" 2>/dev/null || true
    return 0
  }

  # Deliver the path to the AI pane as a bracketed paste (-p), regardless of
  # which pane is currently active.
  "$tmux_cmd" set-buffer -b gt-screenshot -- "$latest"
  "$tmux_cmd" paste-buffer -d -p -b gt-screenshot -t "${session}:0.${pane}"
  "$tmux_cmd" select-pane -t "${session}:0.${pane}" 2>/dev/null || true
}
