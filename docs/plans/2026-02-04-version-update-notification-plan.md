# Version Update Notification Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Show a notification in the interactive menu when a newer version of ghost-tab is available via Homebrew.

**Architecture:** When the menu loads, check a cache file (`~/.config/ghost-tab/.update-check`) for a cached update result. If the cache is missing or older than 24 hours, spawn a background `brew outdated` check that writes results to the cache. The `draw_menu` function reads the cache and conditionally renders an update notification line between the title and separator. The notification only appears for Homebrew installs.

**Tech Stack:** Bash

---

### Task 1: Add background version check and cache logic

**Files:**
- Modify: `ghostty/claude-wrapper.sh:10` (after PROJECTS_FILE, before the `if` block at line 12)

**Step 1: Add the update check function and trigger**

Insert after line 10 (`PROJECTS_FILE=...`) and before line 12 (`if [ -n "$1" ]...`):

```bash
# Version update check (Homebrew only)
UPDATE_CACHE="${XDG_CONFIG_HOME:-$HOME/.config}/ghost-tab/.update-check"
_update_version=""

check_for_update() {
  local cache_ts now age latest
  # Only check if brew is available (Homebrew install)
  command -v brew &>/dev/null || return

  # Read cache if it exists
  if [ -f "$UPDATE_CACHE" ]; then
    latest="$(sed -n '1p' "$UPDATE_CACHE")"
    cache_ts="$(sed -n '2p' "$UPDATE_CACHE")"
    now="$(date +%s)"
    age=$(( now - ${cache_ts:-0} ))
    # Use cached result if less than 24 hours old
    if [ "$age" -lt 86400 ]; then
      _update_version="$latest"
      return
    fi
  fi

  # Spawn background check (non-blocking)
  (
    result="$(brew outdated --verbose --formula ghost-tab 2>/dev/null)"
    mkdir -p "$(dirname "$UPDATE_CACHE")"
    if [ -n "$result" ]; then
      # Extract new version: "ghost-tab (1.0.0) < 1.1.0" -> "1.1.0"
      new_ver="$(echo "$result" | sed -n 's/.*< *//p')"
      printf '%s\n%s\n' "$new_ver" "$(date +%s)" > "$UPDATE_CACHE"
    else
      printf '\n%s\n' "$(date +%s)" > "$UPDATE_CACHE"
    fi
  ) &
  disown
}

check_for_update
```

**Step 2: Verify the function doesn't break startup**

```bash
# From the repo root
bash ghostty/claude-wrapper.sh
```

Expected: The interactive menu loads normally. No errors. No visible delay.

**Step 3: Verify cache file creation**

After running the menu once (then quitting), check:

```bash
cat ~/.config/ghost-tab/.update-check
```

Expected: Two lines — empty first line (no update available for current version) and a Unix timestamp.

**Step 4: Commit**

```bash
git add ghostty/claude-wrapper.sh
git commit -m "Add background version update check with daily cache"
```

---

### Task 2: Display update notification in draw_menu

**Files:**
- Modify: `ghostty/claude-wrapper.sh` — the `draw_menu` function (line 278) and the `_menu_h` calculation (line 290)

**Step 1: Add notification line to _menu_h calculation**

In `draw_menu()`, after line 289 (`[ "${#projects[@]}" -gt 0 ] && _sep_count=1`), add a variable to track the notification height:

```bash
      _update_line=0
      [ -n "$_update_version" ] && _update_line=1
```

Then modify the `_menu_h` calculation on line 290 from:

```bash
      _menu_h=$(( 3 + total * 2 + _sep_count + 2 ))
```

To:

```bash
      _menu_h=$(( 3 + _update_line + total * 2 + _sep_count + 2 ))
```

**Step 2: Render the notification line between title and separator**

In `draw_menu()`, after line 301 (the title line):

```bash
      moveto "$r" "$c"; printf "${_BOLD}${_CYAN}⬡  Ghost Tab${_NC}\033[K"; r=$((r+1))
```

Add the notification line before the separator:

```bash
      if [ -n "$_update_version" ]; then
        moveto "$r" "$c"; printf "  ${_YELLOW}Update available: v${_update_version}${_NC} ${_DIM}(brew upgrade ghost-tab)${_NC}\033[K"; r=$((r+1))
      fi
```

The existing separator line follows immediately after.

**Step 3: Test with a fake cached version**

```bash
mkdir -p ~/.config/ghost-tab
printf '2.0.0\n%s\n' "$(date +%s)" > ~/.config/ghost-tab/.update-check
bash ghostty/claude-wrapper.sh
```

Expected: The menu shows:
```
⬡  Ghost Tab
  Update available: v2.0.0 (brew upgrade ghost-tab)
──────────────────────────────────────
```

**Step 4: Test with no update available**

```bash
printf '\n%s\n' "$(date +%s)" > ~/.config/ghost-tab/.update-check
bash ghostty/claude-wrapper.sh
```

Expected: The menu renders normally with no notification line — identical to the current behavior.

**Step 5: Clean up test cache and commit**

```bash
rm ~/.config/ghost-tab/.update-check
git add ghostty/claude-wrapper.sh
git commit -m "Show update notification in menu when new version available"
```
