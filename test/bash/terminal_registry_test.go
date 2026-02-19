package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func terminalRegistrySnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	registryPath := filepath.Join(root, "lib", "terminals", "registry.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, registryPath, body)
}

func TestGetSupportedTerminals_lists_four_terminals(t *testing.T) {
	snippet := terminalRegistrySnippet(t, `get_supported_terminals`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	for _, name := range []string{"ghostty", "iterm2", "wezterm", "kitty"} {
		assertContains(t, out, name)
	}
}

func TestGetTerminalDisplayName_returns_display_names(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		display string
	}{
		{"ghostty", "ghostty", "Ghostty"},
		{"iterm2", "iterm2", "iTerm2"},
		{"wezterm", "wezterm", "WezTerm"},
		{"kitty", "kitty", "kitty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snippet := terminalRegistrySnippet(t,
				fmt.Sprintf(`get_terminal_display_name %q`, tt.input))
			out, code := runBashSnippet(t, snippet, nil)
			assertExitCode(t, code, 0)
			got := strings.TrimSpace(out)
			if got != tt.display {
				t.Errorf("got %q, want %q", got, tt.display)
			}
		})
	}
}

func TestLoadTerminalPreference_reads_saved_file(t *testing.T) {
	tmpDir := t.TempDir()
	writeTempFile(t, tmpDir, "terminal", "wezterm\n")
	snippet := terminalRegistrySnippet(t,
		fmt.Sprintf(`load_terminal_preference %q`, filepath.Join(tmpDir, "terminal")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	if got != "wezterm" {
		t.Errorf("got %q, want %q", got, "wezterm")
	}
}

func TestLoadTerminalPreference_returns_empty_when_missing(t *testing.T) {
	tmpDir := t.TempDir()
	snippet := terminalRegistrySnippet(t,
		fmt.Sprintf(`load_terminal_preference %q`, filepath.Join(tmpDir, "nonexistent")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSaveTerminalPreference_writes_file(t *testing.T) {
	tmpDir := t.TempDir()
	prefFile := filepath.Join(tmpDir, "terminal")
	snippet := terminalRegistrySnippet(t,
		fmt.Sprintf(`save_terminal_preference %q %q`, "kitty", prefFile))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(prefFile)
	if err != nil {
		t.Fatalf("failed to read pref file: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "kitty" {
		t.Errorf("got %q, want %q", got, "kitty")
	}
}
