package bash_test

import (
	"fmt"
	"path/filepath"
	"testing"
)

func terminalAdapterSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")
	registryPath := filepath.Join(root, "lib", "terminals", "registry.sh")
	adapterPath := filepath.Join(root, "lib", "terminals", "adapter.sh")
	return fmt.Sprintf("source %q && source %q && source %q && source %q && %s",
		tuiPath, installPath, registryPath, adapterPath, body)
}

func TestLoadTerminalAdapter_sources_ghostty_adapter(t *testing.T) {
	snippet := terminalAdapterSnippet(t,
		`load_terminal_adapter "ghostty" && type terminal_get_config_path &>/dev/null && echo "loaded"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "loaded")
}

func TestLoadTerminalAdapter_sources_kitty_adapter(t *testing.T) {
	snippet := terminalAdapterSnippet(t,
		`load_terminal_adapter "kitty" && type terminal_get_config_path &>/dev/null && echo "loaded"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "loaded")
}

func TestLoadTerminalAdapter_sources_wezterm_adapter(t *testing.T) {
	snippet := terminalAdapterSnippet(t,
		`load_terminal_adapter "wezterm" && type terminal_get_config_path &>/dev/null && echo "loaded"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "loaded")
}

func TestLoadTerminalAdapter_sources_iterm2_adapter(t *testing.T) {
	snippet := terminalAdapterSnippet(t,
		`load_terminal_adapter "iterm2" && type terminal_get_config_path &>/dev/null && echo "loaded"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "loaded")
}

func TestLoadTerminalAdapter_fails_for_unknown(t *testing.T) {
	snippet := terminalAdapterSnippet(t,
		`load_terminal_adapter "unknown_terminal"`)
	_, code := runBashSnippet(t, snippet, nil)
	if code == 0 {
		t.Error("expected non-zero exit for unknown terminal")
	}
}
