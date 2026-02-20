package bash_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// settingsJsonSnippet builds a bash snippet that sources tui.sh and settings-json.sh,
// then runs the provided bash code.
func settingsJsonSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	settingsJsonPath := filepath.Join(root, "lib", "settings-json.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, settingsJsonPath, body)
}

func TestSettingsJson_merge_claude_settings_creates_file_when_missing(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`merge_claude_settings %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Created Claude settings with status line")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("settings.json should have been created: %v", err)
	}
	content := string(data)
	assertContains(t, content, `"statusLine"`)
	assertContains(t, content, "statusline-wrapper.sh")
}

func TestSettingsJson_merge_claude_settings_adds_status_line_to_existing(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {}
}
`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`merge_claude_settings %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Added status line to Claude settings")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	content := string(data)
	assertContains(t, content, `"statusLine"`)
}

// --- add_waiting_indicator_hooks ---

func TestSettingsJson_add_waiting_indicator_hooks_creates_file_with_Stop_PreToolUse_and_UserPromptSubmit(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "added")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("settings.json should have been created: %v", err)
	}
	content := string(data)
	assertContains(t, content, `"Stop"`)
	assertContains(t, content, `"PreToolUse"`)
	assertContains(t, content, `"UserPromptSubmit"`)
	assertContains(t, content, "GHOST_TAB_MARKER_FILE")
}

func TestSettingsJson_add_waiting_indicator_hooks_adds_to_existing_settings(t *testing.T) {
	tmpDir := t.TempDir()
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

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "added")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	content := string(data)
	assertContains(t, content, "afplay")
	assertContains(t, content, "GHOST_TAB_MARKER_FILE")
	assertContains(t, content, `"PreToolUse"`)
}

func TestSettingsJson_add_waiting_indicator_hooks_reports_exists_when_duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && touch \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && rm -f \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ]
  }
}
`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, strings.TrimSpace(out), "exists")
}

// --- add_waiting_indicator_hooks: safe exit code format ---

func TestSettingsJson_add_waiting_indicator_hooks_uses_if_then_fi_not_and_operator(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("settings.json should have been created: %v", err)
	}
	content := string(data)

	// Must use safe if-then-fi format
	assertContains(t, content, `if [ -n`)
	assertContains(t, content, `; then`)
	assertContains(t, content, `; fi`)

	// Must NOT use old && format that returns exit 1 when var is empty
	assertNotContains(t, content, `] && touch`)
	assertNotContains(t, content, `] && rm`)
}

func TestSettingsJson_hook_commands_exit_zero_when_marker_env_var_empty(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// Generate hooks into settings file
	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Parse the generated JSON to extract hook commands
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	type hookEntry struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	type hookGroup struct {
		Hooks []hookEntry `json:"hooks"`
	}
	var settings struct {
		Hooks map[string][]hookGroup `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Collect all hook commands
	var commands []string
	for _, groups := range settings.Hooks {
		for _, group := range groups {
			for _, h := range group.Hooks {
				if strings.Contains(h.Command, "GHOST_TAB_MARKER_FILE") {
					commands = append(commands, h.Command)
				}
			}
		}
	}

	if len(commands) == 0 {
		t.Fatal("no GHOST_TAB_MARKER_FILE hook commands found in generated settings")
	}

	// Run each command with GHOST_TAB_MARKER_FILE="" — must exit 0
	for _, cmd := range commands {
		bashScript := fmt.Sprintf(`GHOST_TAB_MARKER_FILE="" ; %s`, cmd)
		_, exitCode := runBashSnippet(t, bashScript, nil)
		if exitCode != 0 {
			t.Errorf("command should exit 0 when GHOST_TAB_MARKER_FILE is empty, got %d for: %s", exitCode, cmd)
		}
	}
}

func TestSettingsJson_hook_commands_exit_zero_when_marker_env_var_set(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")

	// Generate hooks into settings file
	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`add_waiting_indicator_hooks %q`, settingsFile))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Parse the generated JSON to extract hook commands
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}

	type hookEntry struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	type hookGroup struct {
		Hooks []hookEntry `json:"hooks"`
	}
	var settings struct {
		Hooks map[string][]hookGroup `json:"hooks"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse settings.json: %v", err)
	}

	// Collect all hook commands
	var commands []string
	for _, groups := range settings.Hooks {
		for _, group := range groups {
			for _, h := range group.Hooks {
				if strings.Contains(h.Command, "GHOST_TAB_MARKER_FILE") {
					commands = append(commands, h.Command)
				}
			}
		}
	}

	if len(commands) == 0 {
		t.Fatal("no GHOST_TAB_MARKER_FILE hook commands found in generated settings")
	}

	markerFile := filepath.Join(tmpDir, "test-marker")

	// Test with marker file NOT existing yet — touch should create it, rm -f should succeed
	for _, cmd := range commands {
		// Remove marker between each command so each starts fresh
		os.Remove(markerFile)
		bashScript := fmt.Sprintf(`export GHOST_TAB_MARKER_FILE=%q ; %s`, markerFile, cmd)
		_, exitCode := runBashSnippet(t, bashScript, nil)
		if exitCode != 0 {
			t.Errorf("command should exit 0 when marker file does not exist, got %d for: %s", exitCode, cmd)
		}
	}

	// Test with marker file already existing — touch and rm -f should both succeed
	if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}
	for _, cmd := range commands {
		// Re-create marker before each command so it always exists
		if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
			t.Fatalf("failed to re-create marker file: %v", err)
		}
		bashScript := fmt.Sprintf(`export GHOST_TAB_MARKER_FILE=%q ; %s`, markerFile, cmd)
		_, exitCode := runBashSnippet(t, bashScript, nil)
		if exitCode != 0 {
			t.Errorf("command should exit 0 when marker file exists, got %d for: %s", exitCode, cmd)
		}
	}
}

// --- remove_waiting_indicator_hooks ---

func TestSettingsJson_remove_waiting_indicator_hooks_removes_all_three_hooks(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "if [ -n \"$GHOST_TAB_MARKER_FILE\" ]; then touch \"$GHOST_TAB_MARKER_FILE\"; fi"}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "if [ -n \"$GHOST_TAB_MARKER_FILE\" ]; then rm -f \"$GHOST_TAB_MARKER_FILE\"; fi"}]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [{"type": "command", "command": "if [ -n \"$GHOST_TAB_MARKER_FILE\" ]; then rm -f \"$GHOST_TAB_MARKER_FILE\"; fi"}]
      }
    ]
  }
}
`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`remove_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "removed")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	assertNotContains(t, string(data), "GHOST_TAB_MARKER_FILE")
}

func TestSettingsJson_remove_waiting_indicator_hooks_preserves_other_hooks(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Bottle.aiff &"}]
      },
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && touch \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && rm -f \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ]
  }
}
`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`remove_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "removed")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	content := string(data)
	assertContains(t, content, "afplay")
	assertNotContains(t, content, "GHOST_TAB_MARKER_FILE")
}

func TestSettingsJson_remove_waiting_indicator_hooks_returns_not_found_when_absent(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{}`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`remove_waiting_indicator_hooks %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, strings.TrimSpace(out), "not_found")
}

func TestSettingsJson_merge_claude_settings_skips_when_already_configured(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "statusLine": {
    "type": "command",
    "command": "bash ~/.claude/statusline-wrapper.sh"
  }
}
`)

	snippet := settingsJsonSnippet(t,
		fmt.Sprintf(`merge_claude_settings %q`, settingsFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already configured")
}

// --- cleanup_waiting_indicator_hooks (wrapper.sh cleanup logic) ---

// cleanupHooksSnippet builds a bash snippet that sources the required libraries
// and runs the cleanup-time hook removal logic extracted from wrapper.sh.
// It simulates the conditional: if claude + no other markers, remove hooks.
// markerDir controls where the snippet looks for marker files (for test isolation).
func cleanupHooksSnippet(t *testing.T, aiTool, settingsFile, markerDir string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	settingsJsonPath := filepath.Join(root, "lib", "settings-json.sh")
	return fmt.Sprintf(`source %q && source %q
SELECTED_AI_TOOL=%q
if [ "$SELECTED_AI_TOOL" = "claude" ]; then
  if ! ls %s/ghost-tab-waiting-* &>/dev/null; then
    remove_waiting_indicator_hooks %q
  fi
fi
`, tuiPath, settingsJsonPath, aiTool, markerDir, settingsFile)
}

func TestCleanupHooksRemoval_removes_hooks_when_claude_and_no_markers(t *testing.T) {
	tmpDir := t.TempDir()
	markerDir := filepath.Join(tmpDir, "markers")
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		t.Fatalf("failed to create marker dir: %v", err)
	}
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && touch \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && rm -f \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ]
  }
}
`)

	// No marker files exist in markerDir — hooks should be removed
	snippet := cleanupHooksSnippet(t, "claude", settingsFile, markerDir)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "removed")

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	assertNotContains(t, string(data), "GHOST_TAB_MARKER_FILE")
}

func TestCleanupHooksRemoval_skips_when_other_markers_exist(t *testing.T) {
	tmpDir := t.TempDir()
	markerDir := filepath.Join(tmpDir, "markers")
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		t.Fatalf("failed to create marker dir: %v", err)
	}
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && touch \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && rm -f \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ]
  }
}
`)

	// Create a marker file in the isolated marker dir to simulate another session
	markerFile := filepath.Join(markerDir, "ghost-tab-waiting-99999")
	if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}

	snippet := cleanupHooksSnippet(t, "claude", settingsFile, markerDir)
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Hooks should still be present because another marker exists
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	assertContains(t, string(data), "GHOST_TAB_MARKER_FILE")
}

func TestCleanupHooksRemoval_skips_when_not_claude(t *testing.T) {
	tmpDir := t.TempDir()
	markerDir := filepath.Join(tmpDir, "markers")
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		t.Fatalf("failed to create marker dir: %v", err)
	}
	settingsFile := writeTempFile(t, tmpDir, "settings.json", `{
  "hooks": {
    "Stop": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && touch \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ],
    "PreToolUse": [
      {
        "hooks": [{"type": "command", "command": "[ -n \"$GHOST_TAB_MARKER_FILE\" ] && rm -f \"$GHOST_TAB_MARKER_FILE\""}]
      }
    ]
  }
}
`)

	// No marker files, but AI tool is codex — hooks should NOT be removed
	snippet := cleanupHooksSnippet(t, "codex", settingsFile, markerDir)
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("failed to read settings.json: %v", err)
	}
	assertContains(t, string(data), "GHOST_TAB_MARKER_FILE")
}
