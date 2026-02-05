# Tab Spinner Animation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Animate the terminal tab title with a braille spinner when Claude Code is idle, helping users know where to look.

**Architecture:** Two shell scripts (`tab-spinner-start.sh`, `tab-spinner-stop.sh`) triggered by Claude Code hooks. Start script runs a background loop updating tab title; stop script kills it and restores normal title.

**Tech Stack:** Bash, Claude Code hooks system, ANSI escape codes

---

### Task 1: Create the spinner start script

**Files:**
- Create: `~/.claude/tab-spinner-start.sh`

**Step 1: Write the start script**

Create `~/.claude/tab-spinner-start.sh`:

```bash
#!/bin/bash

# Get project name from PWD or tmux session
PROJECT="${PWD##*/}"
if command -v tmux &>/dev/null && tmux list-sessions &>/dev/null; then
  SESSION_NAME=$(tmux display-message -p '#S' 2>/dev/null)
  [ -n "$SESSION_NAME" ] && PROJECT="$SESSION_NAME"
fi

# PID file for this project
PID_FILE="/tmp/ghost-tab-spinner-${PROJECT}.pid"

# Exit if already running
if [ -f "$PID_FILE" ]; then
  OLD_PID=$(cat "$PID_FILE")
  if kill -0 "$OLD_PID" 2>/dev/null; then
    exit 0
  fi
  rm -f "$PID_FILE"
fi

# Spinner frames
FRAMES=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)

# Run animation in background
(
  echo $$ > "$PID_FILE"
  i=0
  while true; do
    printf '\033]0;%s %s\007' "${FRAMES[$i]}" "$PROJECT"
    i=$(( (i + 1) % ${#FRAMES[@]} ))
    sleep 0.1
  done
) &

disown
```

**Step 2: Make it executable and test**

Run:
```bash
chmod +x ~/.claude/tab-spinner-start.sh
~/.claude/tab-spinner-start.sh
```

Expected: Tab title starts spinning with braille animation + project name.

Run:
```bash
cat /tmp/ghost-tab-spinner-*.pid
```

Expected: Shows a PID number.

**Step 3: Kill the test spinner**

Run:
```bash
kill $(cat /tmp/ghost-tab-spinner-*.pid) 2>/dev/null
rm /tmp/ghost-tab-spinner-*.pid
```

---

### Task 2: Create the spinner stop script

**Files:**
- Create: `~/.claude/tab-spinner-stop.sh`

**Step 1: Write the stop script**

Create `~/.claude/tab-spinner-stop.sh`:

```bash
#!/bin/bash

# Get project name from PWD or tmux session
PROJECT="${PWD##*/}"
if command -v tmux &>/dev/null && tmux list-sessions &>/dev/null; then
  SESSION_NAME=$(tmux display-message -p '#S' 2>/dev/null)
  [ -n "$SESSION_NAME" ] && PROJECT="$SESSION_NAME"
fi

# PID file for this project
PID_FILE="/tmp/ghost-tab-spinner-${PROJECT}.pid"

# Kill spinner if running
if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  kill "$PID" 2>/dev/null
  rm -f "$PID_FILE"
fi

# Restore normal tab title
printf '\033]0;%s\007' "$PROJECT"
```

**Step 2: Make it executable and test**

Run:
```bash
chmod +x ~/.claude/tab-spinner-stop.sh
```

First start a spinner:
```bash
~/.claude/tab-spinner-start.sh
```

Then stop it:
```bash
~/.claude/tab-spinner-stop.sh
```

Expected: Animation stops, tab title shows just project name.

---

### Task 3: Add tab spinner setup section to bin/ghost-tab

**Files:**
- Modify: `bin/ghost-tab:391` (insert after sound notification section, before summary)

**Step 1: Add the tab spinner section**

Insert after line 391 (after `fi` closing sound notification) and before line 393 (`# ---------- Summary ----------`):

```bash

# ---------- Tab Title Animation ----------
header "Tab title animation..."
echo ""
echo -e "  Show a spinning animation in the tab title when Claude"
echo -e "  is waiting for your input."
echo ""
read -rn1 -p "$(echo -e "${BLUE}Enable tab animation? (y/n):${NC} ")" enable_spinner </dev/tty
echo ""

if [[ "$enable_spinner" =~ ^[yY]$ ]]; then
  mkdir -p ~/.claude

  # Create start script
  cat > ~/.claude/tab-spinner-start.sh << 'SPINSTART'
#!/bin/bash

# Get project name from PWD or tmux session
PROJECT="${PWD##*/}"
if command -v tmux &>/dev/null && tmux list-sessions &>/dev/null; then
  SESSION_NAME=$(tmux display-message -p '#S' 2>/dev/null)
  [ -n "$SESSION_NAME" ] && PROJECT="$SESSION_NAME"
fi

# PID file for this project
PID_FILE="/tmp/ghost-tab-spinner-${PROJECT}.pid"

# Exit if already running
if [ -f "$PID_FILE" ]; then
  OLD_PID=$(cat "$PID_FILE")
  if kill -0 "$OLD_PID" 2>/dev/null; then
    exit 0
  fi
  rm -f "$PID_FILE"
fi

# Spinner frames
FRAMES=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)

# Run animation in background
(
  echo $$ > "$PID_FILE"
  i=0
  while true; do
    printf '\033]0;%s %s\007' "${FRAMES[$i]}" "$PROJECT"
    i=$(( (i + 1) % ${#FRAMES[@]} ))
    sleep 0.1
  done
) &

disown
SPINSTART

  # Create stop script
  cat > ~/.claude/tab-spinner-stop.sh << 'SPINSTOP'
#!/bin/bash

# Get project name from PWD or tmux session
PROJECT="${PWD##*/}"
if command -v tmux &>/dev/null && tmux list-sessions &>/dev/null; then
  SESSION_NAME=$(tmux display-message -p '#S' 2>/dev/null)
  [ -n "$SESSION_NAME" ] && PROJECT="$SESSION_NAME"
fi

# PID file for this project
PID_FILE="/tmp/ghost-tab-spinner-${PROJECT}.pid"

# Kill spinner if running
if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  kill "$PID" 2>/dev/null
  rm -f "$PID_FILE"
fi

# Restore normal tab title
printf '\033]0;%s\007' "$PROJECT"
SPINSTOP

  chmod +x ~/.claude/tab-spinner-start.sh
  chmod +x ~/.claude/tab-spinner-stop.sh
  success "Created spinner scripts"

  # Add hooks to settings.json using Python
  CLAUDE_SETTINGS="$HOME/.claude/settings.json"
  if python3 - "$CLAUDE_SETTINGS" << 'PYEOF'
import json, sys, os

settings_path = sys.argv[1]

# Load existing settings or start fresh
if os.path.exists(settings_path):
    try:
        with open(settings_path, "r") as f:
            settings = json.load(f)
    except (json.JSONDecodeError, ValueError):
        settings = {}
else:
    settings = {}

hooks = settings.setdefault("hooks", {})

# Add idle_prompt hook for start script
idle_list = hooks.setdefault("idle_prompt", [])
start_cmd = "bash ~/.claude/tab-spinner-start.sh &"
if not any(h.get("command") == start_cmd for h in idle_list):
    idle_list.append({"type": "command", "command": start_cmd})

# Add prompt_submit hook for stop script
submit_list = hooks.setdefault("prompt_submit", [])
stop_cmd = "bash ~/.claude/tab-spinner-stop.sh"
if not any(h.get("command") == stop_cmd for h in submit_list):
    submit_list.append({"type": "command", "command": stop_cmd})

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")
PYEOF
  then
    success "Tab animation hooks configured"
  else
    warn "Failed to configure tab animation hooks"
  fi
else
  info "Skipping tab animation"
fi
```

**Step 2: Update summary section to show tab animation status**

After line 404 (after the sound check), add:

```bash
if [ -f ~/.claude/tab-spinner-start.sh ]; then
  success "Tab animation:   Spinner on idle"
fi
```

---

### Task 4: Test the full setup flow

**Step 1: Run setup script**

Run: `bash bin/ghost-tab`

Expected:
- See "Tab title animation..." section after sound notification
- Press `y` → "✓ Created spinner scripts" and "✓ Tab animation hooks configured"
- Summary shows "✓ Tab animation: Spinner on idle"

**Step 2: Verify files created**

Run:
```bash
ls -la ~/.claude/tab-spinner-*.sh
cat ~/.claude/settings.json | grep -A2 idle_prompt
cat ~/.claude/settings.json | grep -A2 prompt_submit
```

Expected: Both scripts exist, hooks are in settings.json.

**Step 3: Test idempotency**

Run `bash bin/ghost-tab` again, press `y` for tab animation.

Expected: No duplicate hooks in settings.json.

---

### Task 5: Commit all changes

**Step 1: Stage and commit**

```bash
git add bin/ghost-tab
git commit -m "Add tab spinner animation prompt during setup

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

**Step 2: Bump version**

Edit `VERSION` to `1.4.0`, then:

```bash
git add VERSION
git commit -m "Bump version to 1.4.0

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```
