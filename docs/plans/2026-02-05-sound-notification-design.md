# Sound Notification on Idle Prompt

## Summary

During `ghost-tab` setup, prompt the user if they want a sound when Claude Code finishes generating and waits for input. If yes, add a Claude Code hook. If no, skip.

## Implementation

### Setup flow change

New step in `bin/ghost-tab` after status line setup:

- Prompt: `Would you like to hear a sound when Claude Code is done generating? (y/n)`
- If `y`: merge the hook into `~/.claude/settings.json`
- If `n`: skip

### Hook configuration

Added to `~/.claude/settings.json`:

```json
"hooks": {
  "Notification": [
    {
      "matcher": "idle_prompt",
      "hooks": [
        {
          "type": "command",
          "command": "afplay /System/Library/Sounds/Bottle.aiff &"
        }
      ]
    }
  ]
}
```

### Edge cases

- If hooks already exist in settings.json, merge without overwriting other hooks
- If the exact `idle_prompt` notification hook already exists, skip (idempotent)
- `afplay` runs with `&` to avoid blocking Claude Code
