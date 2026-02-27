package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func configTuiSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	configPath := filepath.Join(root, "lib", "config-tui.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, configPath, body)
}

func TestConfigMenuInteractive_dispatches_quit(t *testing.T) {
	dir := t.TempDir()
	binDir := mockCommand(t, dir, "ghost-tab-tui", `echo '{"action":"quit"}'`)
	mockCommand(t, dir, "jq", `
		if [ "$1" = "-r" ] && [ "$2" = ".action" ]; then
			read -r input
			# Extract action value from JSON using bash
			action="${input#*\"action\":\"}"
			action="${action%%\"*}"
			echo "$action"
		fi
	`)
	env := buildEnv(t, []string{binDir})

	snippet := configTuiSnippet(t, `config_menu_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
}

func TestConfigMenuInteractive_returns_1_when_binary_missing(t *testing.T) {
	dir := t.TempDir()
	emptyBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(emptyBin, 0755); err != nil {
		t.Fatal(err)
	}
	// Minimal PATH that excludes ~/.local/bin where ghost-tab-tui lives
	env := buildEnv(t, nil, "PATH="+emptyBin+":/usr/bin:/bin:/usr/sbin:/sbin")

	snippet := configTuiSnippet(t, `config_menu_interactive`)
	out, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit when binary missing")
	}
	assertContains(t, out, "ghost-tab-tui")
}

func TestConfigMenuInteractive_dispatches_display_settings(t *testing.T) {
	dir := t.TempDir()
	callCount := filepath.Join(dir, "call_count")
	// First call returns display-settings, second returns quit
	binDir := mockCommand(t, dir, "ghost-tab-tui", fmt.Sprintf(`
count=0
if [ -f %q ]; then count=$(cat %q); fi
count=$((count + 1))
echo "$count" > %q
case "$1" in
  config-menu)
    if [ "$count" -eq 1 ]; then
      echo '{"action":"display-settings"}'
    else
      echo '{"action":"quit"}'
    fi
    ;;
  settings-menu)
    echo '{"action":"quit"}'
    ;;
esac
`, callCount, callCount, callCount))
	mockCommand(t, dir, "jq", `
		if [ "$1" = "-r" ] && [ "$2" = ".action" ]; then
			read -r input
			action="${input#*\"action\":\"}"
			action="${action%%\"*}"
			echo "$action"
		fi
	`)
	env := buildEnv(t, []string{binDir})

	snippet := configTuiSnippet(t, `config_menu_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
}

func TestConfigMenuInteractive_returns_1_on_tui_failure(t *testing.T) {
	dir := t.TempDir()
	binDir := mockCommand(t, dir, "ghost-tab-tui", `exit 1`)
	env := buildEnv(t, []string{binDir})

	snippet := configTuiSnippet(t, `config_menu_interactive`)
	_, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Error("expected non-zero exit on TUI failure")
	}
}
