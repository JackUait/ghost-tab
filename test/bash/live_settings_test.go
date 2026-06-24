package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests cover live propagation of Settings-menu changes into ALL
// already-running sessions. Each session's tab-title-watcher re-reads the
// settings file every poll tick and re-applies the theme accent + tab-title
// mode, so a toggle in the menu reaches every open window without a relaunch.

// --- read_settings_value: parse a key=value line from the settings file ---

func TestLiveSettings_read_settings_value_returns_value(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "settings", "animation=on\ntab_title=model\ntheme=purple\n")
	settings := filepath.Join(dir, "settings")

	out, code := runBashFunc(t, "lib/tab-title-watcher.sh", "read_settings_value",
		[]string{settings, "theme"}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "purple" {
		t.Errorf("expected 'purple', got %q", strings.TrimSpace(out))
	}
}

func TestLiveSettings_read_settings_value_empty_when_key_absent(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "settings", "animation=on\n")
	settings := filepath.Join(dir, "settings")

	out, code := runBashFunc(t, "lib/tab-title-watcher.sh", "read_settings_value",
		[]string{settings, "theme"}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty for absent key, got %q", strings.TrimSpace(out))
	}
}

func TestLiveSettings_read_settings_value_empty_when_file_missing(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "no-such-settings")

	out, code := runBashFunc(t, "lib/tab-title-watcher.sh", "read_settings_value",
		[]string{settings, "theme"}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty when file missing, got %q", strings.TrimSpace(out))
	}
}

// --- apply_session_theme: re-paint a running session's chrome ---

// recordingTmux records the full arg string of every invocation to $GT_REC.
const recordingTmux = `#!/bin/bash
printf '%s\n' "$*" >> "$GT_REC"
exit 0
`

func TestLiveSettings_apply_session_theme_sets_pane_border(t *testing.T) {
	dir := t.TempDir()
	rec := filepath.Join(dir, "rec")
	binDir := mockCommand(t, dir, "tmux", recordingTmux)
	env := buildEnv(t, []string{binDir}, "GT_REC="+rec)
	tmuxPath := filepath.Join(binDir, "tmux")

	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	watcherPath := filepath.Join(root, "lib", "tab-title-watcher.sh")
	snippet := fmt.Sprintf("source %q && source %q && apply_session_theme %q dev-test-1 141",
		tuiPath, watcherPath, tmuxPath)

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	data, _ := os.ReadFile(rec)
	got := string(data)
	assertContains(t, got, "pane-active-border-style")
	assertContains(t, got, "fg=colour141")
	assertContains(t, got, "dev-test-1")
}

// When lib/spare-tabs.sh is also loaded, apply_session_theme must also repaint
// the nested spare-pane tab bar so the whole window stays one colour.
func TestLiveSettings_apply_session_theme_repaints_spare_chip(t *testing.T) {
	dir := t.TempDir()
	rec := filepath.Join(dir, "rec")
	binDir := mockCommand(t, dir, "tmux", recordingTmux)
	env := buildEnv(t, []string{binDir}, "GT_REC="+rec)
	tmuxPath := filepath.Join(binDir, "tmux")

	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	sparePath := filepath.Join(root, "lib", "spare-tabs.sh")
	watcherPath := filepath.Join(root, "lib", "tab-title-watcher.sh")
	snippet := fmt.Sprintf("source %q && source %q && source %q && apply_session_theme %q dev-test-1 141",
		tuiPath, sparePath, watcherPath, tmuxPath)

	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	data, _ := os.ReadFile(rec)
	got := string(data)
	// Inner spare tmux addressed by its -L socket, status-left repainted purple.
	assertContains(t, got, "gtspare_dev-test-1")
	assertContains(t, got, "status-left")
	assertContains(t, got, "colour141")
}

// --- spare_tabs_status_left: the reusable status-left builder ---

func TestSpareTabs_status_left_uses_accent(t *testing.T) {
	out, code := runBashFunc(t, "lib/spare-tabs.sh", "spare_tabs_status_left",
		[]string{"141"}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "colour141")
	assertContains(t, out, "range=user|new") // + button still present
	assertContains(t, out, "#{window_index}")
}

func TestSpareTabs_status_left_defaults_to_orange(t *testing.T) {
	out, code := runBashFunc(t, "lib/spare-tabs.sh", "spare_tabs_status_left",
		[]string{}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "colour209")
}

// spare_tabs_set_accent repaints a running inner tmux's tab bar with a new accent.
func TestSpareTabs_set_accent_repaints_inner_bar(t *testing.T) {
	dir := t.TempDir()
	rec := filepath.Join(dir, "rec")
	binDir := mockCommand(t, dir, "tmux", recordingTmux)
	env := buildEnv(t, []string{binDir}, "GT_REC="+rec)

	_, code := runBashFunc(t, "lib/spare-tabs.sh", "spare_tabs_set_accent",
		[]string{"gtspare_x", "78"}, env)
	assertExitCode(t, code, 0)

	data, _ := os.ReadFile(rec)
	got := string(data)
	assertContains(t, got, "gtspare_x")
	assertContains(t, got, "status-left")
	assertContains(t, got, "colour78")
}

// --- watcher loop wiring: live re-read each tick ---

func TestLiveSettings_watcher_loop_rereads_settings_live(t *testing.T) {
	root := projectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "lib", "tab-title-watcher.sh"))
	if err != nil {
		t.Fatalf("failed to read tab-title-watcher.sh: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "read_settings_value") {
		t.Error("watcher loop should re-read settings live via read_settings_value")
	}
	if !strings.Contains(content, "apply_session_theme") {
		t.Error("watcher loop should re-apply the theme live via apply_session_theme")
	}
	// The tab-title mode must be read live (not only the frozen launch arg), so
	// a mid-session change reaches the running watcher.
	if !strings.Contains(content, "cur_tab_title") {
		t.Error("watcher loop should track a live tab-title value (cur_tab_title), not only the frozen launch arg")
	}
}

// spare_tabs_config must build its status-left through spare_tabs_status_left so
// the launch-time bar and the live-repaint path share one definition.
func TestSpareTabs_config_uses_status_left_helper(t *testing.T) {
	root := projectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "lib", "spare-tabs.sh"))
	if err != nil {
		t.Fatalf("failed to read spare-tabs.sh: %v", err)
	}
	if !strings.Contains(string(data), "spare_tabs_status_left") {
		t.Error("spare_tabs_config should build status-left via spare_tabs_status_left (shared with live repaint)")
	}
}
