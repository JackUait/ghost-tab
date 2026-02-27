package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func terminalSelectTuiSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	selectPath := filepath.Join(root, "lib", "terminal-select-tui.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, selectPath, body)
}

func TestSelectTerminalInteractive_parses_selected_json(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", `echo '{"terminal":"wezterm","selected":true}'`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")
}

func TestSelectTerminalInteractive_returns_1_on_cancel(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", `echo '{"selected":false}'`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit on cancel")
	}
}

func TestSelectTerminalInteractive_returns_1_when_binary_missing(t *testing.T) {
	tmpDir := t.TempDir()
	env := buildEnv(t, []string{filepath.Join(tmpDir, "empty")})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit when binary missing")
	}
}

func TestSelectTerminalInteractive_handles_install_action(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", `echo '{"action":"install","terminal":"wezterm","selected":false}'`)
	mockCommand(t, tmpDir, "brew", `echo "==> Installing wezterm"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")
}

func TestSelectTerminalInteractive_install_action_calls_brew(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "brew.log")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", `echo '{"action":"install","terminal":"kitty","selected":false}'`)
	mockCommand(t, tmpDir, "brew", fmt.Sprintf(`echo "$@" >> %q`, logFile))
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive`)
	_, _ = runBashSnippet(t, snippet, env)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read brew log: %v", err)
	}
	assertContains(t, string(data), "install --cask kitty")
}

func TestSelectTerminalInteractive_passes_current_flag(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "tui.log")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui",
		fmt.Sprintf(`echo "$@" >> %q; echo '{"terminal":"ghostty","selected":true}'`, logFile))
	env := buildEnv(t, []string{binDir})

	// Write a current terminal preference file
	writeTempFile(t, tmpDir, "terminal", "ghostty")

	snippet := terminalSelectTuiSnippet(t,
		fmt.Sprintf(`GHOST_TAB_TERMINAL_PREF=%q; select_terminal_interactive`, filepath.Join(tmpDir, "terminal")))
	_, _ = runBashSnippet(t, snippet, env)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read tui log: %v", err)
	}
	assertContains(t, string(data), "--current")
	assertContains(t, string(data), "ghostty")
}
