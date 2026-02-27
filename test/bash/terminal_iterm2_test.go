package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func iterm2AdapterSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")
	adapterPath := filepath.Join(root, "lib", "terminals", "iterm2.sh")
	return fmt.Sprintf("source %q && source %q && source %q && %s",
		tuiPath, installPath, adapterPath, body)
}

func TestIterm2Adapter_get_config_path(t *testing.T) {
	snippet := iterm2AdapterSnippet(t, `terminal_get_config_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/Library/Preferences/com.googlecode.iterm2.plist"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIterm2Adapter_get_wrapper_path(t *testing.T) {
	snippet := iterm2AdapterSnippet(t, `terminal_get_wrapper_path`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	got := strings.TrimSpace(out)
	home := os.Getenv("HOME")
	expected := home + "/.config/ghost-tab/wrapper.sh"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestIterm2Adapter_install_calls_ensure_cask(t *testing.T) {
	tmpDir := t.TempDir()
	appDir := filepath.Join(tmpDir, "Applications", "iTerm.app")
	os.MkdirAll(appDir, 0755)

	snippet := iterm2AdapterSnippet(t, fmt.Sprintf(
		`APPLICATIONS_DIR=%q terminal_install`, filepath.Join(tmpDir, "Applications")))
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "iTerm found")
}

func TestIterm2Adapter_setup_config_calls_plistbuddy(t *testing.T) {
	tmpDir := t.TempDir()
	wrapperPath := filepath.Join(tmpDir, "wrapper.sh")

	binDir := mockCommand(t, tmpDir, "PlistBuddy", `echo "PlistBuddy called: $*"`)
	env := buildEnv(t, []string{binDir})

	plistFile := filepath.Join(tmpDir, "com.googlecode.iterm2.plist")
	writeTempFile(t, tmpDir, "com.googlecode.iterm2.plist", "")

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_setup_config %q %q`, plistFile, wrapperPath))
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "PlistBuddy called")
}

func TestIterm2Adapter_cleanup_config_calls_plistbuddy_delete(t *testing.T) {
	tmpDir := t.TempDir()

	binDir := mockCommand(t, tmpDir, "PlistBuddy", `echo "PlistBuddy called: $*"`)
	env := buildEnv(t, []string{binDir})

	plistFile := filepath.Join(tmpDir, "com.googlecode.iterm2.plist")
	writeTempFile(t, tmpDir, "com.googlecode.iterm2.plist", "")

	snippet := iterm2AdapterSnippet(t,
		fmt.Sprintf(`terminal_cleanup_config %q`, plistFile))
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Delete")
}
