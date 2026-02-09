# Cross-Tool Feature Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend sound notification, tab spinner, and status line features from Claude Code to Codex CLI and OpenCode.

**Architecture:** Each tool uses its native config system — Codex CLI uses `config.toml` with a `notify` wrapper script, OpenCode uses a global TypeScript plugin. The setup script gates features by tool support.

**Tech Stack:** Bash (setup script), TOML (Codex config), TypeScript (OpenCode plugin), Python (TOML merge)

---

### Task 1: Create Codex notify wrapper script template

The setup script needs to write `~/.config/ghost-tab/codex-notify.sh` — a script that handles both sound and spinner via PID file management.

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Write the codex-notify.sh heredoc in the setup script**

Add a new section after the existing Claude Code status line section. When Codex is selected and sound or spinner is enabled, write the notify script:

```bash
# ---------- Codex CLI Features ----------
if [ "$_sel_codex" -eq 1 ] && { [ "$_sound_enabled" -eq 1 ] || [ "$_spinner_enabled" -eq 1 ]; }; then
  header "Setting up Codex CLI features..."
  mkdir -p ~/.config/ghost-tab

  cat > ~/.config/ghost-tab/codex-notify.sh << 'CXEOF'
#!/bin/bash
# Ghost Tab: Codex CLI notify handler (sound + spinner)

PROJECT="${PWD##*/}"
if command -v tmux &>/dev/null && tmux list-sessions &>/dev/null 2>&1; then
  SESSION_NAME=$(tmux display-message -p '#S' 2>/dev/null)
  [ -n "$SESSION_NAME" ] && PROJECT="$SESSION_NAME"
fi

PID_FILE="/tmp/ghost-tab-spinner-${PROJECT}.pid"
FEATURES_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/ghost-tab/codex-features.json"

# Kill any running spinner
if [ -f "$PID_FILE" ]; then
  OLD_PID=$(cat "$PID_FILE")
  kill "$OLD_PID" 2>/dev/null
  rm -f "$PID_FILE"
  printf '\033]0;%s\007' "$PROJECT"
fi

# Read feature flags
_sound=0; _spinner=0
if [ -f "$FEATURES_FILE" ]; then
  _sound=$(python3 -c "import json; print(json.load(open('$FEATURES_FILE')).get('sound',0))" 2>/dev/null || echo 0)
  _spinner=$(python3 -c "import json; print(json.load(open('$FEATURES_FILE')).get('spinner',0))" 2>/dev/null || echo 0)
fi

# Play sound
if [ "$_sound" -eq 1 ]; then
  afplay /System/Library/Sounds/Bottle.aiff &
fi

# Start spinner
if [ "$_spinner" -eq 1 ]; then
  FRAMES=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
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
fi
CXEOF

  chmod +x ~/.config/ghost-tab/codex-notify.sh
  success "Created Codex CLI notify script"
fi
```

**Step 2: Write the codex-features.json file**

```bash
  # Save Codex feature flags
  python3 -c "
import json, os
path = os.path.expanduser('~/.config/ghost-tab/codex-features.json')
json.dump({'sound': ${_sound_enabled}, 'spinner': ${_spinner_enabled}}, open(path, 'w'))
"
```

**Step 3: Verify the script is created**

Run: `ls -la ~/.config/ghost-tab/codex-notify.sh`
Expected: File exists, is executable

**Step 4: Commit**

```bash
git add bin/ghost-tab
git commit -m "feat: add Codex CLI notify wrapper script to setup"
```

---

### Task 2: Create Codex config.toml merge logic

Write or merge Codex CLI configuration into `~/.codex/config.toml` during setup.

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Add Codex config.toml creation/merge using Python**

After writing the notify script (Task 1), add TOML config merge:

```bash
  # Configure Codex config.toml
  mkdir -p ~/.codex

  if python3 - ~/.codex/config.toml "$_sound_enabled" "$_spinner_enabled" << 'PYEOF'
import sys, os, json

config_path = sys.argv[1]
sound_enabled = int(sys.argv[2])
spinner_enabled = int(sys.argv[3])

# Simple TOML read/write (no external deps)
config_lines = []
if os.path.exists(config_path):
    with open(config_path, "r") as f:
        config_lines = f.readlines()

# Check if notify already set
has_notify = any(line.strip().startswith("notify") for line in config_lines)
has_tui_section = any(line.strip() == "[tui]" for line in config_lines)
has_status_line = any(line.strip().startswith("status_line") for line in config_lines)

with open(config_path, "a") as f:
    if not has_notify and (sound_enabled or spinner_enabled):
        if sound_enabled and not spinner_enabled:
            f.write('\nnotify = ["afplay", "/System/Library/Sounds/Bottle.aiff"]\n')
        else:
            f.write('\nnotify = ["bash", "~/.config/ghost-tab/codex-notify.sh"]\n')

    if not has_tui_section:
        f.write("\n[tui]\n")

    if not has_status_line:
        f.write('status_line = ["model-with-reasoning", "git-branch", "context-remaining", "used-tokens"]\n')

print("ok")
PYEOF
  then
    success "Configured Codex CLI (config.toml)"
  else
    warn "Failed to configure Codex CLI"
  fi
```

**Step 2: Verify the config**

Run: `cat ~/.codex/config.toml`
Expected: Contains `notify` and `[tui]` section with `status_line`

**Step 3: Commit**

```bash
git add bin/ghost-tab
git commit -m "feat: add Codex CLI config.toml merge to setup"
```

---

### Task 3: Create OpenCode global plugin

Write `~/.config/opencode/plugins/ghost-tab.ts` during setup when OpenCode is selected.

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Add OpenCode plugin creation to setup script**

Add a new section for OpenCode features:

```bash
# ---------- OpenCode Features ----------
if [ "$_sel_opencode" -eq 1 ] && { [ "$_sound_enabled" -eq 1 ] || [ "$_spinner_enabled" -eq 1 ]; }; then
  header "Setting up OpenCode features..."
  mkdir -p ~/.config/opencode/plugins
  mkdir -p ~/.config/ghost-tab

  # Save feature flags
  python3 -c "
import json, os
path = os.path.expanduser('~/.config/ghost-tab/opencode-features.json')
json.dump({'sound': ${_sound_enabled}, 'spinner': ${_spinner_enabled}}, open(path, 'w'))
"

  cat > ~/.config/opencode/plugins/ghost-tab.ts << 'OCEOF'
import { spawn, execSync } from "child_process"
import { readFileSync, writeFileSync, unlinkSync, existsSync } from "fs"
import { join } from "path"
import { tmpdir, homedir } from "os"

// Read feature flags
const configPath = join(
  process.env.XDG_CONFIG_HOME || join(homedir(), ".config"),
  "ghost-tab",
  "opencode-features.json"
)
let features = { sound: false, spinner: false }
try {
  features = JSON.parse(readFileSync(configPath, "utf-8"))
} catch {}

// Spinner state
function getProject(): string {
  try {
    const session = execSync("tmux display-message -p '#S'", { stdio: ["pipe", "pipe", "ignore"] })
      .toString().trim()
    if (session) return session
  } catch {}
  return process.cwd().split("/").pop() || "opencode"
}

function pidFile(): string {
  return join(tmpdir(), `ghost-tab-spinner-${getProject()}.pid`)
}

function killSpinner(): void {
  const pf = pidFile()
  if (existsSync(pf)) {
    try {
      const pid = parseInt(readFileSync(pf, "utf-8").trim())
      process.kill(pid)
    } catch {}
    try { unlinkSync(pf) } catch {}
    const project = getProject()
    process.stdout.write(`\x1b]0;${project}\x07`)
  }
}

function startSpinner(): void {
  const project = getProject()
  const frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
  const child = spawn("bash", ["-c", `
    echo $$ > "${pidFile()}"
    i=0
    while true; do
      printf '\\033]0;%s %s\\007' "${frames.join(" ").split(" ")[0]}" "${project}"
      sleep 0.1
    done
  `.replace(/\$\{frames.*\}/, "")], { stdio: "ignore", detached: true })

  // Write PID and run spinner in bash for portability
  const script = `
    echo $$ > "${pidFile()}"
    FRAMES=(${frames.join(" ")})
    i=0
    while true; do
      printf '\\033]0;%s %s\\007' "\${FRAMES[\$i]}" "${project}"
      i=$(( (i + 1) % \${#FRAMES[@]} ))
      sleep 0.1
    done
  `
  const proc = spawn("bash", ["-c", script], { stdio: "ignore", detached: true })
  proc.unref()
}

export const GhostTab = async () => {
  return {
    event: async ({ event }: { event: { type: string; properties?: any } }) => {
      if (event.type === "session.idle") {
        if (features.sound) {
          spawn("afplay", ["/System/Library/Sounds/Bottle.aiff"], { stdio: "ignore" })
        }
        if (features.spinner) {
          killSpinner()
          startSpinner()
        }
      }

      if (event.type === "session.status") {
        const status = event.properties?.status
        if (status?.type === "busy" && features.spinner) {
          killSpinner()
        }
      }
    },
  }
}
OCEOF

  success "Created OpenCode plugin"
fi
```

**Step 2: Verify the plugin is created**

Run: `ls -la ~/.config/opencode/plugins/ghost-tab.ts`
Expected: File exists

**Step 3: Commit**

```bash
git add bin/ghost-tab
git commit -m "feat: add OpenCode global plugin to setup"
```

---

### Task 4: Refactor sound/spinner prompts to be tool-aware

Move the sound notification and tab spinner prompts outside the Claude-only gate. Show them when any supporting tool is selected.

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Add tool-support detection variables**

Before the sound prompt section, add:

```bash
# Check if any selected tool supports sound/spinner
_supports_sound=0
_supports_spinner=0
[ "$_sel_claude" -eq 1 ] && { _supports_sound=1; _supports_spinner=1; }
[ "$_sel_codex" -eq 1 ] && { _supports_sound=1; _supports_spinner=1; }
[ "$_sel_opencode" -eq 1 ] && { _supports_sound=1; _supports_spinner=1; }
```

**Step 2: Gate the sound prompt on `_supports_sound`**

Change the existing sound notification section from:
```bash
# (inside Claude-only block)
echo -e "  Claude Code can play a sound when it finishes generating"
```

To a standalone section gated on `_supports_sound`:
```bash
if [ "$_supports_sound" -eq 1 ]; then
  header "Sound notification..."
  echo ""
  echo -e "  Play a sound when the AI finishes generating"
  echo -e "  and is waiting for your input."
  ...
fi
```

Set `_sound_enabled=1` or `_sound_enabled=0` based on user choice.

**Step 3: Gate the spinner prompt on `_supports_spinner`**

Same pattern — move the tab animation section outside Claude-only gate, change the description to be tool-agnostic:

```bash
if [ "$_supports_spinner" -eq 1 ]; then
  header "Tab title animation..."
  echo ""
  echo -e "  Show a spinning animation in the tab title when the AI"
  echo -e "  is waiting for your input."
  ...
fi
```

Set `_spinner_enabled=1` or `_spinner_enabled=0` based on user choice.

**Step 4: Keep Claude-specific hook writing inside Claude gate**

The existing Python code that writes to `~/.claude/settings.json` stays inside `if [ "$_sel_claude" -eq 1 ]` but now reads the shared `_sound_enabled` and `_spinner_enabled` variables.

**Step 5: Commit**

```bash
git add bin/ghost-tab
git commit -m "refactor: make sound and spinner prompts tool-agnostic"
```

---

### Task 5: Add Codex status line section to setup

Add a dedicated section that configures Codex CLI's built-in status line (no npm dependencies needed).

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Add Codex status line section**

After the Claude Code status line section, add:

```bash
# ---------- Codex CLI Status Line ----------
if [ "$_sel_codex" -eq 1 ]; then
  header "Setting up Codex CLI status line..."
  # Status line is configured as part of config.toml in the Codex features section
  # (handled by Task 2's TOML merge)
  success "Codex CLI status line configured (model, branch, context, tokens)"
fi
```

This is mostly informational since Task 2 already writes the `status_line` key to `config.toml`. This section just provides user feedback.

**Step 2: Commit**

```bash
git add bin/ghost-tab
git commit -m "feat: add Codex CLI status line setup feedback"
```

---

### Task 6: Update summary section

Update the setup summary to show features for each configured tool.

**Files:**
- Modify: `bin/ghost-tab`

**Step 1: Add Codex and OpenCode lines to summary**

After the existing summary lines, add:

```bash
if [ -f ~/.codex/config.toml ]; then
  success "Codex config:    ~/.codex/config.toml"
fi
if [ -f ~/.config/opencode/plugins/ghost-tab.ts ]; then
  success "OpenCode plugin: ~/.config/opencode/plugins/ghost-tab.ts"
fi
```

**Step 2: Commit**

```bash
git add bin/ghost-tab
git commit -m "feat: show Codex and OpenCode config in setup summary"
```

---

### Task 7: Test the full setup flow

Verify the complete setup flow works for each tool combination.

**Step 1: Run the setup script**

```bash
./bin/ghost-tab
```

Test with different tool selections:
- Claude Code only (existing behavior, should not regress)
- Codex CLI only (should create config.toml and notify script)
- OpenCode only (should create plugin and features.json)
- All tools selected (should configure everything)

**Step 2: Verify created files**

Check each output file exists and has correct content:
- `~/.codex/config.toml`
- `~/.config/ghost-tab/codex-notify.sh`
- `~/.config/ghost-tab/codex-features.json`
- `~/.config/opencode/plugins/ghost-tab.ts`
- `~/.config/ghost-tab/opencode-features.json`

**Step 3: Verify Codex notify script runs**

```bash
bash ~/.config/ghost-tab/codex-notify.sh
```
Expected: Sound plays (if enabled), spinner starts in tab title

**Step 4: Commit any fixes**

```bash
git add bin/ghost-tab
git commit -m "fix: setup flow adjustments from testing"
```

---

### Task 8: Version bump

**Files:**
- Modify: `VERSION`

**Step 1: Bump version**

```bash
echo "1.7.0" > VERSION
```

**Step 2: Commit**

```bash
git add VERSION
git commit -m "bump version to 1.7.0"
```
