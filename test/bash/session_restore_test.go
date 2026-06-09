package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func quote(s string) string { return "\"" + s + "\"" }

func TestCurrentBootId_parses_sec_value(t *testing.T) {
	dir := t.TempDir()
	// macOS sysctl prints: { sec = 1700000000, usec = 123456 } Thu ...
	binDir := mockCommand(t, dir, "sysctl", `echo "{ sec = 1700000000, usec = 123456 } Thu Jan  1 00:00:00 2024"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashFunc(t, "lib/session-restore.sh", "current_boot_id", nil, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "1700000000" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "1700000000")
	}
}

func TestCurrentBootId_empty_when_sysctl_fails(t *testing.T) {
	dir := t.TempDir()
	binDir := mockCommand(t, dir, "sysctl", `exit 1`)
	env := buildEnv(t, []string{binDir})
	out, _ := runBashFunc(t, "lib/session-restore.sh", "current_boot_id", nil, env)
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected empty, got %q", strings.TrimSpace(out))
	}
}

// referenced by later tasks; keep import of filepath/os used
var _ = filepath.Join
var _ = os.Environ
