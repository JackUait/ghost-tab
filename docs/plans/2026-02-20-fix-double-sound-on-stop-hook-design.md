# Fix Double-Sound on Stop Hook

## Problem

When Claude Code's agent stops, two sounds play simultaneously:

1. Ghost-tab's `Stop` hook plays `afplay /System/Library/Sounds/<sound>.aiff &`
2. Claude Code's built-in macOS notification plays the system alert sound and shows a banner

Ghost-tab adds its own sound hook but never disables Claude Code's default notification, so both fire on the same event.

## Root Cause

Claude Code has a built-in notification system (`preferredNotifChannel`) that defaults to macOS notifications. When ghost-tab adds a Stop hook for custom sound, both the hook and the built-in notification trigger independently.

## Solution

When ghost-tab enables its Stop sound hook, also set Claude Code's `preferredNotifChannel` to `terminal_bell`. When it disables the hook, restore the original value.

### Why `terminal_bell`

- In Ghostty (default config), terminal bell produces no audio -- only dock bounce and badge icon
- User retains visual notification feedback (dock bounce signals Claude finished)
- Only ghost-tab's custom sound plays as audio
- Works correctly across all supported terminals

### Why not `notifications_disabled`

- Loses all visual notification (no dock bounce, no badge)
- Less user-friendly for the common case

## Changes

### `lib/settings-json.sh`

- `add_sound_notification_hook()`: After adding the Stop hook, run `claude config set --global preferredNotifChannel terminal_bell` to suppress the built-in audio notification. Store the previous value in `~/.config/ghost-tab/prev-notif-channel` for restoration.
- `remove_sound_notification_hook()`: After removing the Stop hook, restore the previous `preferredNotifChannel` value from the stored file. If no previous value was stored, reset to default by unsetting the config.

### `lib/notification-setup.sh`

- No changes needed. It delegates to `settings-json.sh` functions which handle the channel configuration.

## Edge Cases

- **User already set `preferredNotifChannel` manually**: Store original value before overriding, restore on disable.
- **`claude` CLI not available**: Guard the config command with `command -v claude`.
- **Non-Claude AI tools**: Only apply for Claude Code (existing `case "$tool" in claude)` guard).
- **Nested Claude sessions**: Use `CLAUDECODE="" claude config set` to avoid nested session errors.

## Testing

- Verify only one sound plays when Stop hook fires
- Verify `preferredNotifChannel` is set to `terminal_bell` after enabling sound
- Verify original `preferredNotifChannel` is restored after disabling sound
- Verify behavior when `claude` CLI is not available
- Verify no regression in sound toggle/cycle functionality
