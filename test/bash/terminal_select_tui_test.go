package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// Successful install should auto-select the terminal (no second TUI call)
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
`, callCount, callCount, callCount))
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
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"kitty","cask":"kitty","selected":false}'
`, callCount, callCount, callCount))
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
	// Binary exits non-zero but outputs valid install JSON.
	// Brew succeeds — should auto-select without re-launching TUI.
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
exit 1
`, callCount, callCount, callCount))
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
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"myterm","cask":"my-term-cask","selected":false}'
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

func TestSelectTerminalInteractive_autoselects_on_successful_install_no_relaunch(t *testing.T) {
	// Successful brew install should auto-select without re-launching TUI.
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `echo "installed"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")
}

func TestSelectTerminalInteractive_cancel_after_failed_install_returns_1(t *testing.T) {
	// When brew install fails, TUI relaunches. If user then cancels, return 1.
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
	mockCommand(t, tmpDir, "brew", `exit 1`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit on cancel after failed install")
	}
}

func TestSelectTerminalInteractive_autoselects_after_successful_install(t *testing.T) {
	// After successful brew install, the terminal should be auto-selected
	// WITHOUT relaunching the TUI (only 1 TUI call, not 2).
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	// Mock TUI: outputs install action every time but exits after 2 calls
	// to prevent infinite loops if auto-select isn't working.
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"wezterm","cask":"wezterm","selected":false}'
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `echo "installed"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=wezterm")

	// Verify TUI was only called once (auto-selected, no re-launch)
	data, err := os.ReadFile(callCount)
	if err != nil {
		t.Fatalf("failed to read call count: %v", err)
	}
	if strings.TrimSpace(string(data)) != "1" {
		t.Errorf("expected TUI to be called 1 time, got %s", strings.TrimSpace(string(data)))
	}
}

func TestSelectTerminalInteractive_autoselects_uses_terminal_field(t *testing.T) {
	// When cask field differs from terminal field, auto-select should
	// use the terminal name (not the cask name).
	tmpDir := t.TempDir()
	callCount := filepath.Join(tmpDir, "call_count")
	binDir := mockCommand(t, tmpDir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
if [ "$count" -gt 2 ]; then echo '{"selected":false}'; exit 0; fi
echo '{"action":"install","terminal":"myterm","cask":"my-term-cask","selected":false}'
`, callCount, callCount, callCount))
	mockCommand(t, tmpDir, "brew", `echo "installed"`)
	env := buildEnv(t, []string{binDir})

	snippet := terminalSelectTuiSnippet(t,
		`select_terminal_interactive && echo "SELECTED=$_selected_terminal"`)
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "SELECTED=myterm")
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
