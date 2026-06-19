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

# gt_latest_screenshot <dir>... — print the newest image file across the given
# dirs. Returns non-zero (printing nothing) when no dir exists or none has an
# image. Uses find+stat (not globbing) so it is robust across bash/zsh and when
# some image extensions have no matches. Multiple dirs let the injector search
# both the saved location (Desktop) and the screencaptureui temp dir holding a
# just-taken, not-yet-saved floating-thumbnail screenshot.
gt_latest_screenshot() {
  local dirs=() d
  for d in "$@"; do
    [ -d "$d" ] && dirs+=("$d")
  done
  [ ${#dirs[@]} -gt 0 ] || return 1

  local latest=""
  local line
  # Newest-first: macOS `stat -f '%m %N'` prints "<mtime-seconds> <path>".
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    latest="${line#* }"  # strip the leading "<mtime> "
    break
  done < <(find "${dirs[@]}" -maxdepth 1 -type f \
            \( -iname '*.png' -o -iname '*.jpg' -o -iname '*.jpeg' \) \
            -exec stat -f '%m %N' {} + 2>/dev/null | sort -rn)

  [ -n "$latest" ] || return 1
  printf '%s\n' "$latest"
}

# gt_screenshot_temp_dirs — print the screencaptureui TemporaryItems dirs that
# hold a just-taken screenshot while its floating thumbnail is still showing
# (before it is saved to the screenshot location). A real OS drag of that
# thumbnail is intermittently broken (macOS hands the terminal an empty,
# promise-only payload), so reading the file straight from here and injecting
# its path bypasses the drag entirely. Base is overridable for tests.
gt_screenshot_temp_dirs() {
  local base="${GT_SCREENSHOT_TEMP_BASE:-$(getconf DARWIN_USER_TEMP_DIR 2>/dev/null)TemporaryItems}"
  local d
  for d in "$base"/NSIRD_screencaptureui_*; do
    [ -d "$d" ] && printf '%s\n' "$d"
  done
}

# _gt_pick_marked_pane — read "<index> <flag>" lines on stdin and print the
# index whose flag is "1" (the AI pane, marked with the @gt_ai pane option).
# Prints nothing and returns non-zero when no pane is marked.
_gt_pick_marked_pane() {
  local idx flag
  while read -r idx flag; do
    if [ "$flag" = "1" ]; then
      printf '%s\n' "$idx"
      return 0
    fi
  done
  return 1
}

# gt_ai_pane <tmux_cmd> <session> — print the AI pane index. Prefers the pane
# marked with @gt_ai=1 (set by wrapper.sh at session creation), so it is robust
# to tmux pane renumbering and non-default layouts. When no pane is marked (e.g.
# a session created by an older ghost-tab), falls back to the full-height pane
# on the right edge of the layout — where the AI tool lives — and only then to
# index 1 as a last resort.
gt_ai_pane() {
  local tmux_cmd="$1" session="$2" idx
  idx="$("$tmux_cmd" list-panes -t "${session}:0" -F '#{pane_index} #{@gt_ai}' 2>/dev/null | _gt_pick_marked_pane)" || idx=""
  if [ -z "$idx" ]; then
    # The AI pane spans the full height on the right (at_right & at_top & at_bottom).
    idx="$("$tmux_cmd" list-panes -t "${session}:0" \
            -F '#{pane_index} #{pane_at_right} #{pane_at_top} #{pane_at_bottom}' 2>/dev/null \
          | awk '$2=="1" && $3=="1" && $4=="1"{print $1; exit}')"
  fi
  [ -n "$idx" ] || idx=1
  printf '%s\n' "$idx"
}

# gt_focus_ai_pane_when_ready <tmux_cmd> <session> — poll the AI pane until it
# shows a shell/AI prompt, then make it the active pane. Resolves the AI pane via
# gt_ai_pane (marker/geometry) on every poll, so it is correct under any tmux
# pane-base-index — a hardcoded index only matches the AI pane at base-index 0,
# and would otherwise re-focus the wrong pane just after launch.
gt_focus_ai_pane_when_ready() {
  local tmux_cmd="$1" session="$2" pane content
  while true; do
    sleep 0.5
    pane="$(gt_ai_pane "$tmux_cmd" "$session")"
    content="$("$tmux_cmd" capture-pane -t "${session}:0.${pane}" -p 2>/dev/null)"
    # All three tools show a prompt character when ready.
    if printf '%s' "$content" | grep -qE '[>$❯]'; then
      "$tmux_cmd" select-pane -t "${session}:0.${pane}"
      break
    fi
  done
}

# gt_paste_latest_screenshot <session> [pane] — inject the latest screenshot's
# path into the AI pane as a bracketed paste (so Claude attaches it as an image).
# Resolves the AI pane via the @gt_ai marker when no pane is given.
gt_paste_latest_screenshot() {
  local tmux_cmd
  tmux_cmd="$(command -v tmux)" || return 1
  # Default to the session the binding fired in.
  local session="${1:-$("$tmux_cmd" display-message -p '#{session_name}' 2>/dev/null)}"
  [ -n "$session" ] || return 1
  local pane="${2:-$(gt_ai_pane "$tmux_cmd" "$session")}"

  # Search the saved location AND the screencaptureui temp dirs, so a screenshot
  # taken moments ago (still a floating thumbnail, not yet on disk in the saved
  # dir) is found too.
  local dir latest line temp_dirs=()
  dir="$(gt_screenshot_dir)"
  while IFS= read -r line; do [ -n "$line" ] && temp_dirs+=("$line"); done < <(gt_screenshot_temp_dirs)
  latest="$(gt_latest_screenshot "$dir" "${temp_dirs[@]}")" || {
    "$tmux_cmd" display-message "ghost-tab: no screenshot found in $dir" 2>/dev/null || true
    return 0
  }

  # Copy to a stable, ghost-tab-owned path first. A just-taken screenshot often
  # lives in a screencaptureui temp dir that macOS deletes within moments; if we
  # injected that path the AI tool would read a file that has already vanished
  # and silently attach nothing. The stable copy can't disappear out from under it.
  local stable
  stable="$(gt_stash_screenshot "$latest")" || stable="$latest"

  # Deliver the path to the AI pane as a bracketed paste (-p), regardless of
  # which pane is currently active.
  "$tmux_cmd" set-buffer -b gt-screenshot -- "$stable"
  "$tmux_cmd" paste-buffer -d -p -b gt-screenshot -t "${session}:0.${pane}"
  "$tmux_cmd" select-pane -t "${session}:0.${pane}" 2>/dev/null || true
}

# gt_stable_screenshot_dir — print (creating) a stable, ghost-tab-owned directory
# screenshots are copied into so their paths survive macOS deleting the original
# screencaptureui temp file. Overridable via GT_SCREENSHOT_STASH_DIR for tests.
gt_stable_screenshot_dir() {
  local d="${GT_SCREENSHOT_STASH_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/ghost-tab/screenshots}"
  mkdir -p "$d" 2>/dev/null || true
  printf '%s\n' "$d"
}

# gt_stash_screenshot <src> — copy a screenshot into the stable dir and print the
# stable path. The original is left untouched. Returns non-zero if src is missing.
gt_stash_screenshot() {
  local src="$1"
  [ -f "$src" ] || return 1
  local dir dest
  dir="$(gt_stable_screenshot_dir)"
  dest="$dir/${src##*/}"
  cp -p "$src" "$dest" 2>/dev/null || return 1
  printf '%s\n' "$dest"
}

# gt_ghost_tab_mra_session <tmux_cmd> — print the ghost-tab session the user is
# most likely looking at: among sessions that own an @gt_ai pane, the one whose
# client was most recently active. With several ghost-tab windows open this is
# what keeps an auto-injected screenshot landing in the session in front of you
# instead of every Claude at once. Prints nothing / non-zero when none qualify.
gt_ghost_tab_mra_session() {
  local tmux_cmd="$1" gt_sessions
  gt_sessions="$("$tmux_cmd" list-panes -a -F '#{session_name} #{@gt_ai}' 2>/dev/null \
    | awk '$2=="1"{print $1}' | sort -u | tr '\n' ' ')"
  [ -n "$gt_sessions" ] || return 1
  "$tmux_cmd" list-clients -F '#{client_session} #{client_activity}' 2>/dev/null \
    | awk -v gts="$gt_sessions" '
        BEGIN { n=split(gts, a, " "); for (i=1;i<=n;i++) if (a[i]!="") gt[a[i]]=1 }
        ($1 in gt) && ($2+0 >= best+0) { best=$2+0; sess=$1 }
        END { if (sess!="") print sess }'
}

# gt_autoinject_tick <tmux_cmd> <session> <state_file> <dir>... — one watcher
# iteration: if the newest screenshot across <dir>... is newer than the last one
# handled (recorded in state_file) AND <session> is the most-recently-active
# ghost-tab session, copy it to a stable path and inject it into the AI pane.
# Safe to call repeatedly; only the session the user is in injects.
gt_autoinject_tick() {
  local tmux_cmd="$1" session="$2" state_file="$3"
  shift 3
  local latest mtime last
  latest="$(gt_latest_screenshot "$@")" || return 0
  mtime="$(stat -f '%m' "$latest" 2>/dev/null)" || return 0
  last="$(cat "$state_file" 2>/dev/null || echo 0)"
  [ -n "$last" ] || last=0
  # Only react to a screenshot newer than the last one this session handled.
  [ "$mtime" -gt "$last" ] 2>/dev/null || return 0
  # Record it as handled regardless of whether we inject, so a screenshot meant
  # for another session isn't re-evaluated here forever.
  printf '%s\n' "$mtime" > "$state_file" 2>/dev/null || true

  local mra
  mra="$(gt_ghost_tab_mra_session "$tmux_cmd")" || return 0
  [ "$mra" = "$session" ] || return 0

  local stable pane
  stable="$(gt_stash_screenshot "$latest")" || return 0
  pane="$(gt_ai_pane "$tmux_cmd" "$session")"
  "$tmux_cmd" set-buffer -b gt-screenshot -- "$stable"
  "$tmux_cmd" paste-buffer -d -p -b gt-screenshot -t "${session}:0.${pane}"
  "$tmux_cmd" select-pane -t "${session}:0.${pane}" 2>/dev/null || true
}

# gt_watch_and_inject_screenshots <tmux_cmd> <session> — background loop that
# auto-injects newly-taken screenshots into the AI pane (no drag, no hotkey).
# Starts "caught up" so it ignores screenshots taken before launch, then polls
# the saved location and the screencaptureui temp dirs once a second.
gt_watch_and_inject_screenshots() {
  local tmux_cmd="$1" session="$2"
  local state_file dir line temp_dirs=() initial
  state_file="$(gt_stable_screenshot_dir)/.autoinject-${session}.state"
  dir="$(gt_screenshot_dir)"
  while IFS= read -r line; do [ -n "$line" ] && temp_dirs+=("$line"); done < <(gt_screenshot_temp_dirs)
  if initial="$(gt_latest_screenshot "$dir" "${temp_dirs[@]}" 2>/dev/null)"; then
    stat -f '%m' "$initial" 2>/dev/null > "$state_file" || printf '0\n' > "$state_file"
  else
    printf '0\n' > "$state_file"
  fi
  while true; do
    sleep 1
    temp_dirs=()
    while IFS= read -r line; do [ -n "$line" ] && temp_dirs+=("$line"); done < <(gt_screenshot_temp_dirs)
    gt_autoinject_tick "$tmux_cmd" "$session" "$state_file" "$dir" "${temp_dirs[@]}"
  done
}
