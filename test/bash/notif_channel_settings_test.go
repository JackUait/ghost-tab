package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests pin the notification-channel silencing to direct settings.json
// writes. Claude Code 2.1.190 removed the `claude config` subcommand, so the
// old `claude config set preferredNotifChannel` approach is dead — ghost-tab
// must write preferredNotifChannel into ~/.claude/settings.json itself.
//
// Why this matters: in a fresh session Claude emits its own audible idle
// notification (preferredNotifChannel unset → audible default). ghost-tab's
// "off" flag only gates ghost-tab's own afplay, so without silencing Claude's
// channel the sound plays even when ghost-tab sound is Off. Forcing
// terminal_bell (silent in Ghostty) leaves ghost-tab's afplay as the single
// audible source, so Off is truly silent.

// settingsNotifChannel reads preferredNotifChannel from a settings.json file,
// returning "__ABSENT__" when the key is missing.
func settingsNotifChannel(t *testing.T, settingsPath string) string {
	t.Helper()
	verify := fmt.Sprintf(
		`python3 -c "import json; d=json.load(open('%s')); print(d.get('preferredNotifChannel','__ABSENT__'))"`,
		settingsPath)
	out, code := runBashSnippet(t, verify, nil)
	assertExitCode(t, code, 0)
	return strings.TrimSpace(out)
}

func TestNotifChannel_set_writes_terminal_bell_when_settings_absent(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	settingsPath := filepath.Join(tmpDir, "claude", "settings.json")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "terminal_bell" {
		t.Errorf("preferredNotifChannel = %q, want terminal_bell", got)
	}
	prev, err := os.ReadFile(filepath.Join(configDir, "prev-notif-channel"))
	if err != nil {
		t.Fatalf("expected prev-notif-channel saved: %v", err)
	}
	if strings.TrimSpace(string(prev)) != "__UNSET__" {
		t.Errorf("prev = %q, want __UNSET__ sentinel for an absent prior value", strings.TrimSpace(string(prev)))
	}
}

func TestNotifChannel_set_preserves_existing_keys_and_saves_prev(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json",
		`{"preferredNotifChannel": "iterm2", "model": "opus"}`)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "terminal_bell" {
		t.Errorf("preferredNotifChannel = %q, want terminal_bell", got)
	}
	verify := fmt.Sprintf(`python3 -c "import json; print(json.load(open('%s'))['model'])"`, settingsPath)
	out, _ := runBashSnippet(t, verify, nil)
	if strings.TrimSpace(out) != "opus" {
		t.Errorf("expected unrelated key 'model' preserved, got %q", strings.TrimSpace(out))
	}
	prev, _ := os.ReadFile(filepath.Join(configDir, "prev-notif-channel"))
	if strings.TrimSpace(string(prev)) != "iterm2" {
		t.Errorf("prev = %q, want iterm2", strings.TrimSpace(string(prev)))
	}
}

func TestNotifChannel_set_is_idempotent_and_keeps_existing_prev(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json", `{"preferredNotifChannel": "terminal_bell"}`)
	// A prior session already saved the real previous value.
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	prev, _ := os.ReadFile(filepath.Join(configDir, "prev-notif-channel"))
	if strings.TrimSpace(string(prev)) != "iterm2" {
		t.Errorf("prev clobbered: got %q, want iterm2 (a second concurrent session must not overwrite the saved prior value)", strings.TrimSpace(string(prev)))
	}
}

// The user's real-world settings.json holds preferredNotifChannel: "" (present
// but empty). This must be distinguished from an absent key: set saves "" (NOT
// the __UNSET__ sentinel reserved for absent), and restore writes "" back rather
// than removing the key. Guards the load-bearing distinction in
// set_claude_notif_channel (`"__UNSET__" if current is None else str(current)`).
func TestNotifChannel_set_and_restore_round_trip_present_empty_string(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json", `{"preferredNotifChannel": "", "model": "opus"}`)

	// set: silences the channel, saving the empty string (not __UNSET__).
	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "terminal_bell" {
		t.Errorf("preferredNotifChannel = %q, want terminal_bell", got)
	}
	prev, err := os.ReadFile(filepath.Join(configDir, "prev-notif-channel"))
	if err != nil {
		t.Fatalf("expected prev-notif-channel saved: %v", err)
	}
	if strings.TrimSpace(string(prev)) != "" {
		t.Errorf("prev = %q, want empty string (present-but-empty must NOT become __UNSET__)", strings.TrimSpace(string(prev)))
	}

	// restore: writes "" back; the key stays present (not removed).
	snippet = notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code = runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "" {
		t.Errorf("preferredNotifChannel = %q, want \"\" (present empty string; key must remain, not be removed)", got)
	}
}

// A corrupt (unparseable) settings.json must never be clobbered. Claude's own
// settings.json holds the user's entire Claude Code configuration; silently
// replacing it with {"preferredNotifChannel":"terminal_bell"} on a JSON parse
// error would destroy all of it. set must fail safe: leave the file untouched
// and save no prev value (so restore has nothing to wrongly write back).
func TestNotifChannel_set_does_not_clobber_corrupt_settings(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	corrupt := `{"preferredNotifChannel": "iterm2", "model": "opus",}` // trailing comma → invalid JSON
	writeTempFile(t, claudeDir, "settings.json", corrupt)

	snippet := notificationSnippet(t,
		fmt.Sprintf(`set_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0) // must not break the wrapper

	got, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	if string(got) != corrupt {
		t.Errorf("corrupt settings.json was modified — data loss.\n got: %q\nwant: %q", string(got), corrupt)
	}
	if _, err := os.Stat(filepath.Join(configDir, "prev-notif-channel")); !os.IsNotExist(err) {
		t.Error("prev-notif-channel must not be saved when settings.json could not be parsed")
	}
}

// restore must likewise never clobber a corrupt settings.json (else the user's
// whole config is replaced with just the restored key).
func TestNotifChannel_restore_does_not_clobber_corrupt_settings(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	corrupt := `{"preferredNotifChannel": "terminal_bell", "model": "opus",}` // invalid JSON
	writeTempFile(t, claudeDir, "settings.json", corrupt)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	got, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	if string(got) != corrupt {
		t.Errorf("corrupt settings.json was modified on restore — data loss.\n got: %q\nwant: %q", string(got), corrupt)
	}
}

func TestNotifChannel_restore_restores_saved_value(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json", `{"preferredNotifChannel": "terminal_bell", "model": "opus"}`)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "iterm2" {
		t.Errorf("preferredNotifChannel = %q, want iterm2", got)
	}
	if _, err := os.Stat(filepath.Join(configDir, "prev-notif-channel")); !os.IsNotExist(err) {
		t.Errorf("prev-notif-channel should be removed after restore")
	}
}

func TestNotifChannel_restore_removes_key_when_prev_unset(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json", `{"preferredNotifChannel": "terminal_bell", "model": "opus"}`)
	writeTempFile(t, configDir, "prev-notif-channel", "__UNSET__")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "__ABSENT__" {
		t.Errorf("preferredNotifChannel = %q, want the key removed", got)
	}
	verify := fmt.Sprintf(`python3 -c "import json; print(json.load(open('%s'))['model'])"`, settingsPath)
	out, _ := runBashSnippet(t, verify, nil)
	if strings.TrimSpace(out) != "opus" {
		t.Errorf("expected unrelated key 'model' preserved, got %q", strings.TrimSpace(out))
	}
}

func TestNotifChannel_restore_noop_when_no_prev_file(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	settingsPath := filepath.Join(tmpDir, "claude", "settings.json")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`restore_claude_notif_channel %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Errorf("restore should not create settings.json when there is no saved prev")
	}
}

func TestNotifChannel_setup_sound_notification_silences_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	settingsPath := filepath.Join(tmpDir, "claude", "settings.json")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`setup_sound_notification %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "terminal_bell" {
		t.Errorf("preferredNotifChannel = %q, want terminal_bell", got)
	}
}

func TestNotifChannel_remove_sound_notification_restores_channel(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	os.MkdirAll(configDir, 0755)
	claudeDir := filepath.Join(tmpDir, "claude")
	os.MkdirAll(claudeDir, 0755)
	settingsPath := filepath.Join(claudeDir, "settings.json")
	writeTempFile(t, claudeDir, "settings.json", `{"preferredNotifChannel": "terminal_bell"}`)
	writeTempFile(t, configDir, "prev-notif-channel", "iterm2")

	snippet := notificationSnippet(t,
		fmt.Sprintf(`remove_sound_notification %q %q`, configDir, settingsPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if got := settingsNotifChannel(t, settingsPath); got != "iterm2" {
		t.Errorf("preferredNotifChannel = %q, want iterm2", got)
	}
}

func TestNotifChannel_wrapper_silences_claude_channel(t *testing.T) {
	root := projectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "wrapper.sh"))
	if err != nil {
		t.Fatalf("read wrapper.sh: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "setup_sound_notification") {
		t.Error("wrapper.sh should call setup_sound_notification so Claude's own idle notification is silenced (otherwise the Off flag never silences it)")
	}
	if !strings.Contains(content, "remove_sound_notification") {
		t.Error("wrapper.sh should call remove_sound_notification on last-session cleanup to restore the channel")
	}
}
