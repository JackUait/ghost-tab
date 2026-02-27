package npx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// projectRoot returns the absolute path to the ghost-tab repo root.
func projectRoot(t *testing.T) string {
	t.Helper()
	// test/npx/ is two levels below root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "..", "..")
}

// runLauncher executes the npx launcher with the given env overrides.
// Returns stdout, stderr, and exit code.
func runLauncher(t *testing.T, env []string, args ...string) (string, string, int) {
	t.Helper()
	root := projectRoot(t)
	launcher := filepath.Join(root, "bin", "npx-ghost-tab.js")
	cmdArgs := append([]string{launcher}, args...)
	cmd := exec.Command("node", cmdArgs...)
	cmd.Env = env
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run launcher: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}
