# Tab Title Spinner Animation Design

When Claude Code finishes responding and waits for input, animate the terminal tab title with a braille spinner so the user knows where to look.

## Animation

**Frames:** `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` (cycles continuously)

**Tab title format:** `⠋ project-name` (spinner prefix)

**Speed:** ~100ms per frame

## Mechanism

Uses Claude Code's hook system:

- `idle_prompt` hook → Start spinner animation
- `prompt_submit` hook → Stop animation, restore normal title

## Files Created

### `~/.claude/tab-spinner-start.sh`

Background script that:
1. Checks if animation already running (via PID file)
2. Writes PID to `/tmp/ghost-tab-spinner-${PROJECT}.pid`
3. Loops through frames, updating tab title via `printf '\033]0;%s\007'`

### `~/.claude/tab-spinner-stop.sh`

Script that:
1. Reads PID from temp file
2. Kills the animation process
3. Restores tab title to project name
4. Cleans up PID file

## Hook Configuration

Added to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "idle_prompt": [{ "command": "~/.claude/tab-spinner-start.sh" }],
    "prompt_submit": [{ "command": "~/.claude/tab-spinner-stop.sh" }]
  }
}
```

Extends existing hooks (e.g., sound notification) rather than replacing them.

## Setup Integration

Added to `bin/ghost-tab` after sound notification section:

1. Prompt: "Would you like tab title animation when Claude is ready for input? (y/n)"
2. Create both scripts in `~/.claude/`
3. Use Python to safely merge hooks into settings.json
4. Idempotent: detects existing scripts, asks to reconfigure

## Edge Cases

**Multiple tabs:** PID file includes project name to prevent cross-talk.

**Duplicate prevention:** Start script exits if animation already running.

**Orphan cleanup:** Stop script removes PID file; start script validates PID is alive.

**Plain terminal mode:** Works without tmux; project name falls back to `basename $PWD`.
