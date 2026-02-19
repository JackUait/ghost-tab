package bash_test

import (
	"fmt"
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
