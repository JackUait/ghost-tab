package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// notificationSnippet builds a bash snippet that sources tui.sh, settings-json.sh,
// and notification-setup.sh, then runs the provided bash code.
func notificationSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	settingsJsonPath := filepath.Join(root, "lib", "settings-json.sh")
	notifPath := filepath.Join(root, "lib", "notification-setup.sh")
	return fmt.Sprintf("source %q && source %q && source %q && %s",
		tuiPath, settingsJsonPath, notifPath, body)
}

// updateSnippet builds a bash snippet that sources update.sh then runs the provided bash code.
func updateSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	updatePath := filepath.Join(root, "lib", "update.sh")
	return fmt.Sprintf("source %q && %s", updatePath, body)
}

// ==================== notification-setup.sh tests ====================

// --- setup_sound_notification ---

// --- is_sound_enabled ---

func TestNotification_is_sound_enabled_returns_true_when_features_file_missing(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistentDir := filepath.Join(tmpDir, "nonexistent")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`is_sound_enabled "claude" %q`, nonexistentDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "true" {
		t.Errorf("expected 'true', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_is_sound_enabled_returns_true_when_sound_key_missing(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`is_sound_enabled "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "true" {
		t.Errorf("expected 'true', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_is_sound_enabled_returns_true_when_sound_is_true(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`is_sound_enabled "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "true" {
		t.Errorf("expected 'true', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_is_sound_enabled_returns_false_when_sound_is_false(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": false}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`is_sound_enabled "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "false" {
		t.Errorf("expected 'false', got %q", strings.TrimSpace(out))
	}
}

// --- remove_sound_notification ---

// --- set_sound_feature_flag ---

func TestNotification_set_sound_feature_flag_creates_file_with_sound_true(t *testing.T) {
	tmpDir := t.TempDir()

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_feature_flag "claude" %q true`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	if _, err := os.Stat(featuresFile); os.IsNotExist(err) {
		t.Fatalf("expected features file to be created at %s", featuresFile)
	}

	// Verify sound is true using python3
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; print(json.load(open('%s'))['sound'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "True" {
		t.Errorf("expected 'True', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_set_sound_feature_flag_sets_sound_false_in_existing_file(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_feature_flag "claude" %q false`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; print(json.load(open('%s'))['sound'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "False" {
		t.Errorf("expected 'False', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_set_sound_feature_flag_preserves_other_keys(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": false, "other": 42}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_feature_flag "claude" %q true`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; d=json.load(open('%s')); print(d['sound'], d['other'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "True 42" {
		t.Errorf("expected 'True 42', got %q", strings.TrimSpace(out))
	}
}

// --- get_sound_name ---

func TestNotification_get_sound_name_returns_Bottle_when_features_file_missing(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistentDir := filepath.Join(tmpDir, "nonexistent")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`get_sound_name "claude" %q`, nonexistentDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Bottle" {
		t.Errorf("expected 'Bottle', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_get_sound_name_returns_Bottle_when_sound_name_key_missing(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`get_sound_name "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Bottle" {
		t.Errorf("expected 'Bottle', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_get_sound_name_returns_stored_name(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true, "sound_name": "Glass"}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`get_sound_name "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Glass" {
		t.Errorf("expected 'Glass', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_get_sound_name_returns_empty_when_sound_disabled(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": false}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`get_sound_name "claude" %q`, tmpDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty string, got %q", strings.TrimSpace(out))
	}
}

// --- set_sound_name ---

func TestNotification_set_sound_name_writes_name_to_features_file(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_name "claude" %q "Glass"`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; print(json.load(open('%s'))['sound_name'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Glass" {
		t.Errorf("expected 'Glass', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_set_sound_name_creates_file_when_missing(t *testing.T) {
	tmpDir := t.TempDir()

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_name "claude" %q "Ping"`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; d=json.load(open('%s')); print(d['sound_name'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Ping" {
		t.Errorf("expected 'Ping', got %q", strings.TrimSpace(out))
	}
}

func TestNotification_set_sound_name_preserves_other_keys(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "claude-features.json", `{"sound": true, "other": 42}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_sound_name "claude" %q "Hero"`, tmpDir))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	featuresFile := filepath.Join(tmpDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; d=json.load(open('%s')); print(d['sound_name'], d['other'])"`, featuresFile)
	out, code := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Hero 42" {
		t.Errorf("expected 'Hero 42', got %q", strings.TrimSpace(out))
	}
}

// --- toggle_sound_notification ---

func TestNotification_toggle_sound_notification_enables_for_claude(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": false}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q`, configDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "enabled")

	// Verify feature flag was set
	featuresFile := filepath.Join(configDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; print(json.load(open('%s'))['sound'])"`, featuresFile)
	flagOut, flagCode := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, flagCode, 0)
	if strings.TrimSpace(flagOut) != "True" {
		t.Errorf("expected feature flag 'True', got %q", strings.TrimSpace(flagOut))
	}
}

func TestNotification_toggle_sound_notification_disables_for_claude(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": true}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q`, configDir))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "disabled")

	// Verify feature flag was set to false
	featuresFile := filepath.Join(configDir, "claude-features.json")
	verifySnippet := fmt.Sprintf(
		`python3 -c "import json; print(json.load(open('%s'))['sound'])"`, featuresFile)
	flagOut, flagCode := runBashSnippet(t, verifySnippet, nil)
	assertExitCode(t, flagCode, 0)
	if strings.TrimSpace(flagOut) != "False" {
		t.Errorf("expected feature flag 'False', got %q", strings.TrimSpace(flagOut))
	}
}

// --- apply_sound_notification ---

// --- toggle_sound_notification + config_dir passthrough ---

func TestNotification_toggle_sound_notification_enables_and_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": false}`)

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "get" ]; then
  echo ""
  exit 0
fi
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "terminal_bell" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q`, configDir))

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

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "iterm2" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`toggle_sound_notification "claude" %q`, configDir))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "disabled")

	// Verify prev-notif-channel was cleaned up
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after disable")
	}
}

// --- apply_sound_notification + config_dir passthrough ---

func TestNotification_apply_sound_notification_enables_and_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "claude-features.json", `{"sound": false}`)

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "get" ]; then
  echo ""
  exit 0
fi
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "terminal_bell" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`apply_sound_notification "claude" %q "Glass"`, configDir))

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

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "iterm2" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`apply_sound_notification "claude" %q ""`, configDir))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "disabled")

	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after disable")
	}
}

// --- setup_sound_notification + config_dir wiring ---

func TestNotification_setup_sound_notification_sets_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "get" ]; then
  echo ""
  exit 0
fi
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "terminal_bell" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`setup_sound_notification %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify prev-notif-channel was saved
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel to be created")
	}
}

// --- remove_sound_notification + config_dir wiring ---

func TestNotification_remove_sound_notification_restores_notif_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")

	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "iterm2" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`remove_sound_notification %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify prev-notif-channel was cleaned up
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after restore")
	}
}

// --- set_claude_notif_channel ---

func TestNotification_set_claude_notif_channel_saves_previous_and_sets_terminal_bell(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// Mock claude CLI: "config get" returns "iterm2", "config set" verifies terminal_bell
	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "get" ]; then
  echo "iterm2"
  exit 0
fi
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "terminal_bell" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify prev-notif-channel file was saved with "iterm2"
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	data, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("expected prev-notif-channel file to exist: %v", err)
	}
	if strings.TrimSpace(string(data)) != "iterm2" {
		t.Errorf("expected saved value 'iterm2', got %q", strings.TrimSpace(string(data)))
	}
}

func TestNotification_set_claude_notif_channel_skips_when_claude_not_available(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// PATH=/nonexistent so claude is not found
	env := buildEnv(t, nil, "PATH=/nonexistent", "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify no prev-notif-channel file was created
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel file NOT to exist, but it does")
	}
}

func TestNotification_set_claude_notif_channel_handles_empty_current_value(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)

	// Mock claude CLI: "config get" returns empty, "config set" succeeds
	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "get" ]; then
  echo ""
  exit 0
fi
if [ "$1" = "config" ] && [ "$2" = "set" ]; then
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

	// Verify prev-notif-channel file exists with empty content
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	data, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("expected prev-notif-channel file to exist: %v", err)
	}
	// File should exist, content should be empty (just a newline from echo)
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty saved value, got %q", strings.TrimSpace(string(data)))
	}
}

// --- restore_claude_notif_channel ---

func TestNotification_restore_claude_notif_channel_restores_saved_value(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2\n")

	// Mock claude CLI: "config set" verifies restoring iterm2
	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "set" ] && [ "$4" = "iterm2" ]; then
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`
	binDir := mockCommand(t, tmpDir, "claude", claudeBody)
	env := buildEnv(t, []string{binDir}, "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify prev-notif-channel file was removed after restore
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel file to be removed after restore")
	}
}

func TestNotification_restore_claude_notif_channel_unsets_when_prev_was_empty(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "prev-notif-channel", "\n")

	// Mock claude CLI: "config set" succeeds
	claudeBody := `
if [ "$1" = "config" ] && [ "$2" = "set" ]; then
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

	// Verify prev-notif-channel file was removed
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel file to be removed after restore")
	}
}

func TestNotification_restore_claude_notif_channel_noop_when_no_saved_file(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	// No prev-notif-channel file created

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

	// PATH=/nonexistent so claude is not found
	env := buildEnv(t, nil, "PATH=/nonexistent", "CLAUDECODE=")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q`, configDir))

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify file still exists (not removed since restore couldn't happen)
	savedFile := filepath.Join(configDir, "prev-notif-channel")
	if _, err := os.Stat(savedFile); os.IsNotExist(err) {
		t.Errorf("expected prev-notif-channel file to still exist when claude not available")
	}
}

// ==================== lib/update.sh tests (git-based) ====================

func TestUpdate_notify_if_updated_does_nothing_when_no_flag(t *testing.T) {
	dir := t.TempDir()

	snippet := updateSnippet(t, fmt.Sprintf(`
XDG_CONFIG_HOME=%q
notify_if_updated
`, filepath.Join(dir, "config")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected no output when no flag file, got %q", out)
	}
}

func TestUpdate_notify_if_updated_shows_version_and_deletes_flag(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "ghost-tab")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "updated", "2.3.0")

	snippet := updateSnippet(t, fmt.Sprintf(`
XDG_CONFIG_HOME=%q
notify_if_updated
`, filepath.Join(dir, "config")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "2.3.0")
	flagFile := filepath.Join(configDir, "updated")
	if _, err := os.Stat(flagFile); !os.IsNotExist(err) {
		t.Errorf("expected flag file to be deleted after notify_if_updated")
	}
}

func TestUpdate_notify_if_updated_shows_update_message(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config", "ghost-tab")
	os.MkdirAll(configDir, 0755)
	writeTempFile(t, configDir, "updated", "2.5.0")

	snippet := updateSnippet(t, fmt.Sprintf(`
XDG_CONFIG_HOME=%q
notify_if_updated
`, filepath.Join(dir, "config")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Updated")
}

func TestUpdate_check_for_update_does_nothing_when_not_git_repo(t *testing.T) {
	dir := t.TempDir()
	shareDir := t.TempDir() // not a git repo — no .git directory
	writeTempFile(t, shareDir, "VERSION", "2.2.0")

	snippet := updateSnippet(t, fmt.Sprintf(`
check_for_update %q
sleep 0.2
`, shareDir))
	env := buildEnv(t, nil, "XDG_CONFIG_HOME="+filepath.Join(dir, "config"))
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	flagFile := filepath.Join(dir, "config", "ghost-tab", "updated")
	if _, err := os.Stat(flagFile); !os.IsNotExist(err) {
		t.Errorf("expected no flag file when share_dir is not a git repo")
	}
}
