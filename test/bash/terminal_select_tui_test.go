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

func TestSelectTerminalInteractive_handles_install_action_with_nonzero_exit(t *testing.T) {
	tmpDir := t.TempDir()
	// Binary exits non-zero but still outputs valid install JSON
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui",
		`echo '{"action":"install","terminal":"wezterm","selected":false}'; exit 1`)
	mockCommand(t, tmpDir, "brew", `echo "==> Installing wezterm"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")
}

func TestSelectTerminalInteractive_uses_cask_field_for_brew(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "brew.log")
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -eq 1 ]; then
  echo '{"action":"install","terminal":"myterm","cask":"my-term-cask","selected":false}'
else
  echo '{"terminal":"myterm","selected":true}'
fi
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", fmt.Sprintf(`echo "$@" >> %q`, logFile))
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=myterm")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read brew log: %v", err)
	}
	assertContains(t, string(data), "install --cask my-term-cask")
}

func TestSelectTerminalInteractive_loops_after_failed_install(t *testing.T) {
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -eq 1 ]; then
  echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
else
  echo '{"terminal":"ghostty","selected":true}'
fi
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `exit 1`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=ghostty")
}

func TestSelectTerminalInteractive_loops_after_successful_install(t *testing.T) {
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -eq 1 ]; then
  echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
else
  echo '{"terminal":"wezterm","selected":true}'
fi
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `echo "installed"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")
}

func TestSelectTerminalInteractive_cancel_after_install_returns_1(t *testing.T) {
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -eq 1 ]; then
  echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
else
  echo '{"selected":false}'
fi
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `echo "installed"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit on cancel after install")
	}
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
