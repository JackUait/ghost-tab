package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ghosttyAdapterSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")
	adapterPath := filepath.Join(root, "lib", "terminals", "ghostty.sh")
	return fmt.Sprintf("source %q && source %q && source %q && %s",
		tuiPath, installPath, adapterPath, body)
}

func TestGhosttyAdapter_get_config_path(t *testing.T) {
	snippet := ghosttyAdapterSnippet(t, `terminal_get_config_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/.config/ghostty/config"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestGhosttyAdapter_get_wrapper_path(t *testing.T) {
	snippet := ghosttyAdapterSnippet(t, `terminal_get_wrapper_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/.config/ghost-tab/wrapper.sh"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestGhosttyAdapter_setup_config_creates_new(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")
	wrapperPath := filepath.Join(tmpDir, "wrapper.sh")

	snippet := ghosttyAdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, configFile, wrapperPath))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "command = "+wrapperPath)
}

func TestGhosttyAdapter_setup_config_replaces_existing(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command = /old/path\n")
	wrapperPath := filepath.Join(tmpDir, "wrapper.sh")

	snippet := ghosttyAdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, configFile, wrapperPath))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := strings.TrimSpace(string(data))
	expected := "command = " + wrapperPath
	if content != expected {
		t.Errorf("got %q, want %q", content, expected)
	}
}

func TestGhosttyAdapter_setup_config_preserves_other_settings(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\ntheme = dark\n")
	wrapperPath := filepath.Join(tmpDir, "wrapper.sh")

	snippet := ghosttyAdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, configFile, wrapperPath))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := string(data)
	assertContains(t, content, "font-size = 14")
	assertContains(t, content, "theme = dark")
	assertContains(t, content, "command = "+wrapperPath)
}

func TestGhosttyAdapter_cleanup_config_removes_command_line(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\ncommand = /some/path\ntheme = dark\n")

	snippet := ghosttyAdapterSnippet(t,
		fmt.Sprintf(`terminal_cleanup_config %q`, configFile))
	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := string(data)
	assertContains(t, content, "font-size = 14")
	assertContains(t, content, "theme = dark")
	assertNotContains(t, content, "command =")
}
