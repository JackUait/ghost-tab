# Fix Double-Sound on Stop Hook — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Prevent two sounds from playing when Claude Code's Stop hook fires, by setting `preferredNotifChannel` to `terminal_bell` when ghost-tab's sound hook is active.

**Architecture:** Two new helper functions (`set_claude_notif_channel` and `restore_claude_notif_channel`) in `notification-setup.sh` manage Claude Code's `preferredNotifChannel` config. They are called from `setup_sound_notification()` on enable and `remove_sound_notification()` on disable. The previous channel value is stored in `~/.config/ghost-tab/prev-notif-channel` for restoration.

**Tech Stack:** Bash, `claude` CLI (`CLAUDECODE="" claude config set/get`)

---

### Task 1: Add `set_claude_notif_channel` function

**Files:**
- Modify: `lib/notification-setup.sh` (add function after line 19)
- Test: `test/bash/notification_update_test.go` (add tests after line 91)

**Step 1: Write the failing test**

Add to `test/bash/notification_update_test.go`:

```go
func TestNotification_set_claude_notif_channel_saves_previous_and_sets_terminal_bell(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// Mock claude CLI that reports current channel and accepts set
	claudeBody := `
if [[ "$*" == *"config get"*"preferredNotifChannel"* ]]; then
  echo "iterm2"
  exit 0
fi
if [[ "$*" == *"config set"*"preferredNotifChannel"*"terminal_bell"* ]]; then
  exit 0
fi
echo "unexpected: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify previous value was saved
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	data, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("expected prev-notif-channel file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "iterm2" {
		t.Errorf("expected saved channel 'iterm2', got %q", strings.TrimSpace(string(data)))
	}
}

func TestNotification_set_claude_notif_channel_skips_when_claude_not_available(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// No claude in PATH
	env := buildEnv(t, nil, "PATH=/nonexistent", "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// No file should be created
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should not exist when claude is unavailable")
	}
}

func TestNotification_set_claude_notif_channel_handles_empty_current_value(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// claude config get returns empty (unset)
	claudeBody := `
if [[ "$*" == *"config get"* ]]; then
  echo ""
  exit 0
fi
if [[ "$*" == *"config set"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Should save empty string (meaning "unset/default")
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	data, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("expected prev-notif-channel file: %v", err)
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty saved channel, got %q", strings.TrimSpace(string(data)))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./test/bash/... -run "TestNotification_set_claude_notif_channel" -v`
Expected: FAIL — `set_claude_notif_channel` function does not exist.

**Step 3: Write the implementation**

Add to `lib/notification-setup.sh` after the `setup_sound_notification` function (after line 19):

```bash
# Set Claude Code's preferredNotifChannel to terminal_bell to prevent
# double sounds (ghost-tab hook + built-in notification).
# Saves the previous value to <config_dir>/prev-notif-channel for restoration.
# Usage: set_claude_notif_channel <config_dir>
set_claude_notif_channel() {
  local config_dir="$1"
  if ! command -v claude &>/dev/null; then
    return 0
  fi
  mkdir -p "$config_dir"
  local prev
  prev="$(CLAUDECODE="" claude config get preferredNotifChannel 2>/dev/null || true)"
  echo "$prev" > "$config_dir/prev-notif-channel"
  CLAUDECODE="" claude config set --global preferredNotifChannel terminal_bell 2>/dev/null || true
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./test/bash/... -run "TestNotification_set_claude_notif_channel" -v`
Expected: PASS

**Step 5: Run shellcheck**

Run: `shellcheck lib/notification-setup.sh`
Expected: No errors

**Step 6: Commit**

```bash
git add lib/notification-setup.sh test/bash/notification_update_test.go
git commit -m "feat: add set_claude_notif_channel to suppress built-in notification sound"
```

---

### Task 2: Add `restore_claude_notif_channel` function

**Files:**
- Modify: `lib/notification-setup.sh` (add function after `set_claude_notif_channel`)
- Test: `test/bash/notification_update_test.go`

**Step 1: Write the failing test**

```go
func TestNotification_restore_claude_notif_channel_restores_saved_value(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")

	claudeBody := `
if [[ "$*" == *"config set"*"preferredNotifChannel"*"iterm2"* ]]; then
  exit 0
fi
echo "unexpected: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify saved file was cleaned up
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after restore")
	}
}

func TestNotification_restore_claude_notif_channel_unsets_when_prev_was_empty(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "\n")

	claudeBody := `
if [[ "$*" == *"config set"*"preferredNotifChannel"*""* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
}

func TestNotification_restore_claude_notif_channel_noop_when_no_saved_file(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	// No prev-notif-channel file exists

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}

func TestNotification_restore_claude_notif_channel_skips_when_claude_not_available(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")

	env := buildEnv(t, nil, "PATH=/nonexistent", "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// File should still exist since claude wasn't available to restore
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should still exist when claude is unavailable")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./test/bash/... -run "TestNotification_restore_claude_notif_channel" -v`
Expected: FAIL — `restore_claude_notif_channel` function does not exist.

**Step 3: Write the implementation**

Add to `lib/notification-setup.sh` after `set_claude_notif_channel`:

```bash
# Restore Claude Code's preferredNotifChannel from saved value.
# If no saved value exists, does nothing.
# Usage: restore_claude_notif_channel <config_dir>
restore_claude_notif_channel() {
  local config_dir="$1"
  local saved_file="$config_dir/prev-notif-channel"
  if [ ! -f "$saved_file" ]; then
    return 0
  fi
  if ! command -v claude &>/dev/null; then
    return 0
  fi
  local prev
  prev="$(cat "$saved_file")"
  prev="$(echo "$prev" | tr -d '[:space:]')"
  if [[ -n "$prev" ]]; then
    CLAUDECODE="" claude config set --global preferredNotifChannel "$prev" 2>/dev/null || true
  else
    CLAUDECODE="" claude config set --global preferredNotifChannel "" 2>/dev/null || true
  fi
  rm -f "$saved_file"
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./test/bash/... -run "TestNotification_restore_claude_notif_channel" -v`
Expected: PASS

**Step 5: Run shellcheck**

Run: `shellcheck lib/notification-setup.sh`
Expected: No errors

**Step 6: Commit**

```bash
git add lib/notification-setup.sh test/bash/notification_update_test.go
git commit -m "feat: add restore_claude_notif_channel to restore notification setting"
```

---

### Task 3: Wire channel management into enable/disable paths

**Files:**
- Modify: `lib/notification-setup.sh:7-19` (`setup_sound_notification`) and `lib/notification-setup.sh:108-115` (`remove_sound_notification`)
- Test: `test/bash/notification_update_test.go`

**Step 1: Write the failing test**

```go
func TestNotification_setup_sound_notification_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{}`)

	claudeBody := `
if [[ "$*" == *"config get"* ]]; then
  echo ""
  exit 0
fi
if [[ "$*" == *"config set"*"terminal_bell"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`setup_sound_notification %q "afplay /System/Library/Sounds/Bottle.aiff &" %q`, settingsFile, configDir))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "configured")

	// Verify prev-notif-channel was saved
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel to be created")
	}
}

func TestNotification_remove_sound_notification_restores_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Bottle.aiff &"}]
      }
    ]
  }
}
`)

	claudeBody := `
if [[ "$*" == *"config set"*"iterm2"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`remove_sound_notification %q "afplay /System/Library/Sounds/Bottle.aiff &" %q`, settingsFile, configDir))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "removed")

	// Verify prev-notif-channel was cleaned up
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after restore")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./test/bash/... -run "TestNotification_(setup_sound_notification_sets_notif|remove_sound_notification_restores_notif)" -v`
Expected: FAIL — functions don't accept the `config_dir` parameter yet.

**Step 3: Update `setup_sound_notification` and `remove_sound_notification`**

In `lib/notification-setup.sh`, update `setup_sound_notification` to accept an optional `config_dir` parameter:

```bash
# Add sound notification hook to Claude settings.
# Usage: setup_sound_notification <settings_path> <sound_command> [config_dir]
setup_sound_notification() {
  local settings_path="$1" sound_command="$2" config_dir="${3:-}"
  local result
  result="$(add_sound_notification_hook "$settings_path" "$sound_command")"
  if [ "$result" = "added" ]; then
    success "Sound notification configured"
  elif [ "$result" = "exists" ]; then
    success "Sound notification already configured"
  else
    warn "Failed to configure sound notification"
    return 1
  fi
  if [[ -n "$config_dir" ]]; then
    set_claude_notif_channel "$config_dir"
  fi
}
```

Update `remove_sound_notification` similarly:

```bash
# Remove sound notification hook from Claude settings.
# Usage: remove_sound_notification <settings_path> <sound_command> [config_dir]
remove_sound_notification() {
  local settings_path="$1" sound_command="$2" config_dir="${3:-}"
  local result
  result="$(remove_sound_notification_hook "$settings_path" "$sound_command")"
  echo "$result"
  if [[ -n "$config_dir" ]]; then
    restore_claude_notif_channel "$config_dir"
  fi
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./test/bash/... -run "TestNotification_(setup_sound_notification|remove_sound_notification)" -v`
Expected: ALL PASS (new tests and existing tests since `config_dir` is optional)

**Step 5: Run shellcheck**

Run: `shellcheck lib/notification-setup.sh`
Expected: No errors

**Step 6: Commit**

```bash
git add lib/notification-setup.sh test/bash/notification_update_test.go
git commit -m "feat: wire notif channel management into setup/remove sound notification"
```

---

### Task 4: Update callers to pass `config_dir`

**Files:**
- Modify: `lib/notification-setup.sh:120-145` (`toggle_sound_notification`) and `lib/notification-setup.sh:150-177` (`apply_sound_notification`)
- Test: `test/bash/notification_update_test.go`

**Step 1: Write the failing test**

```go
func TestNotification_toggle_sound_notification_enables_and_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": false}`)
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{}`)

	claudeBody := `
if [[ "$*" == *"config get"* ]]; then
  echo ""
  exit 0
fi
if [[ "$*" == *"config set"*"terminal_bell"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q %q`, configDir, settingsFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "enabled")

	// Verify prev-notif-channel was saved
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel to be created on enable")
	}
}

func TestNotification_toggle_sound_notification_disables_and_restores_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": true}`)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Bottle.aiff &"}]
      }
    ]
  }
}
`)

	claudeBody := `
if [[ "$*" == *"config set"*"iterm2"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q %q`, configDir, settingsFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "disabled")

	// Verify prev-notif-channel was cleaned up
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after disable")
	}
}

func TestNotification_apply_sound_notification_enables_and_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": false}`)
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{}`)

	claudeBody := `
if [[ "$*" == *"config get"* ]]; then
  echo ""
  exit 0
fi
if [[ "$*" == *"config set"*"terminal_bell"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`apply_sound_notification "claude" %q %q "Glass"`, configDir, settingsFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "enabled")

	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel to be created on enable")
	}
}

func TestNotification_apply_sound_notification_disables_and_restores_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": true, "sound_name": "Glass"}`)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Glass.aiff &"}]
      }
    ]
  }
}
`)

	claudeBody := `
if [[ "$*" == *"config set"*"iterm2"* ]]; then
  exit 0
fi
exit 0
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`apply_sound_notification "claude" %q %q ""`, configDir, settingsFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "disabled")

	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after disable")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./test/bash/... -run "TestNotification_(toggle_sound_notification_enables_and_sets|toggle_sound_notification_disables_and_restores|apply_sound_notification_enables_and_sets|apply_sound_notification_disables_and_restores)" -v`
Expected: FAIL — callers don't pass `config_dir` to `setup_sound_notification`/`remove_sound_notification`.

**Step 3: Update callers**

In `toggle_sound_notification`, pass `config_dir` as third arg:

```bash
toggle_sound_notification() {
  local tool="$1" config_dir="$2" settings_path="$3"
  local current
  current="$(is_sound_enabled "$tool" "$config_dir")"
  local sound_command="afplay /System/Library/Sounds/Bottle.aiff &"

  if [[ "$current" == "true" ]]; then
    # Disable
    set_sound_feature_flag "$tool" "$config_dir" false
    case "$tool" in
      claude)
        remove_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications disabled"
  else
    # Enable
    set_sound_feature_flag "$tool" "$config_dir" true
    case "$tool" in
      claude)
        setup_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications enabled"
  fi
}
```

In `apply_sound_notification`, pass `config_dir` as third arg:

```bash
apply_sound_notification() {
  local tool="$1" config_dir="$2" settings_path="$3" sound_name="$4"

  if [[ -z "$sound_name" ]]; then
    # Disable sound
    set_sound_feature_flag "$tool" "$config_dir" false
    case "$tool" in
      claude)
        remove_sound_notification "$settings_path" "afplay /System/Library/Sounds/" "$config_dir"
        ;;
    esac
    success "Sound notifications disabled"
  else
    # Enable sound with specific name
    set_sound_feature_flag "$tool" "$config_dir" true
    set_sound_name "$tool" "$config_dir" "$sound_name"
    local sound_command="afplay /System/Library/Sounds/${sound_name}.aiff &"
    case "$tool" in
      claude)
        # Remove old hook first (any afplay sound), then add new one
        remove_sound_notification "$settings_path" "afplay /System/Library/Sounds/" "$config_dir"
        setup_sound_notification "$settings_path" "$sound_command" "$config_dir"
        ;;
    esac
    success "Sound notifications enabled"
  fi
}
```

**Step 4: Run all notification tests to verify they pass**

Run: `go test ./test/bash/... -run "TestNotification_" -v`
Expected: ALL PASS (new and existing tests)

**Step 5: Run shellcheck**

Run: `shellcheck lib/notification-setup.sh`
Expected: No errors

**Step 6: Commit**

```bash
git add lib/notification-setup.sh test/bash/notification_update_test.go
git commit -m "feat: pass config_dir through callers for notif channel management"
```

---

### Task 5: Final verification and push

**Step 1: Run shellcheck on all modified scripts**

Run: `shellcheck lib/notification-setup.sh lib/settings-json.sh`
Expected: No errors

**Step 2: Run full test suite**

Run: `./run-tests.sh`
Expected: ALL PASS

**Step 3: Push**

```bash
git pull --rebase && git push
git status
```
Expected: "up to date with origin"
