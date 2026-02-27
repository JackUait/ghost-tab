package npx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLauncher_copies_bash_files_to_install_dir(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "share", "ghost-tab")

	env := append(os.Environ(),
		"HOME="+home,
		"GHOST_TAB_INSTALL_DIR="+installDir,
		"GHOST_TAB_SKIP_TUI_DOWNLOAD=1",
		"GHOST_TAB_SKIP_EXEC=1",
	)

	_, stderr, code := runLauncher(t, env)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stderr: %s", code, stderr)
	}

	// Verify key files were copied
	for _, rel := range []string{"bin/ghost-tab", "lib/tui.sh", "wrapper.sh", "VERSION"} {
		path := filepath.Join(installDir, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist in install dir", rel)
		}
	}
}

func TestLauncher_writes_version_marker(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "share", "ghost-tab")

	env := append(os.Environ(),
		"HOME="+home,
		"GHOST_TAB_INSTALL_DIR="+installDir,
		"GHOST_TAB_SKIP_TUI_DOWNLOAD=1",
		"GHOST_TAB_SKIP_EXEC=1",
	)

	_, stderr, code := runLauncher(t, env)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d. stderr: %s", code, stderr)
	}

	// Check .version marker
	marker := filepath.Join(installDir, ".version")
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("expected .version marker: %v", err)
	}

	// Read expected version from repo VERSION file
	root := projectRoot(t)
	expected, _ := os.ReadFile(filepath.Join(root, "VERSION"))
	if strings.TrimSpace(string(data)) != strings.TrimSpace(string(expected)) {
		t.Errorf("version marker = %q, want %q", strings.TrimSpace(string(data)), strings.TrimSpace(string(expected)))
	}
}

func TestLauncher_rejects_unsupported_platform(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "share", "ghost-tab")

	env := append(os.Environ(),
		"HOME="+home,
		"GHOST_TAB_INSTALL_DIR="+installDir,
		"GHOST_TAB_SKIP_TUI_DOWNLOAD=1",
		"GHOST_TAB_SKIP_EXEC=1",
		"GHOST_TAB_MOCK_PLATFORM=linux",
	)

	_, stderr, code := runLauncher(t, env)
	if code == 0 {
		t.Fatal("expected non-zero exit for unsupported platform")
	}
	if !strings.Contains(stderr, "macOS") {
		t.Errorf("expected macOS error message, got: %s", stderr)
	}
}

func TestLauncher_skips_copy_when_version_matches(t *testing.T) {
	home := t.TempDir()
	installDir := filepath.Join(home, ".local", "share", "ghost-tab")

	// First run: copy files
	env := append(os.Environ(),
		"HOME="+home,
		"GHOST_TAB_INSTALL_DIR="+installDir,
		"GHOST_TAB_SKIP_TUI_DOWNLOAD=1",
		"GHOST_TAB_SKIP_EXEC=1",
	)

	_, _, code := runLauncher(t, env)
	if code != 0 {
		t.Fatalf("first run failed with code %d", code)
	}

	// Second run: should skip copy (check output message)
	stdout, _, code := runLauncher(t, env)
	if code != 0 {
		t.Fatalf("second run failed with code %d", code)
	}
	if !strings.Contains(stdout, "up to date") {
		t.Errorf("expected 'up to date' message, got: %s", stdout)
	}
}
