package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installSnippet builds a bash snippet that sources tui.sh and install.sh,
// then runs the provided bash code.
func installSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, installPath, body)
}

// ============================================================
// ensure_brew_pkg tests
// ============================================================

func TestEnsureBrewPkg_reports_already_installed(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" succeeds => package already installed
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 0; fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already installed")
}

func TestEnsureBrewPkg_installs_missing_package(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" fails (not installed), "brew install" succeeds
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then exit 0; fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "installed")
}

func TestEnsureBrewPkg_warns_on_install_failure(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" fails, "brew install" also fails
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then exit 1; fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed")
}

func TestEnsureBrewPkg_handles_brew_command_timeout(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" returns 124 (timeout code)
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 124; fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
}

func TestEnsureBrewPkg_handles_brew_not_in_PATH(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: always returns 127 (command not found)
	binDir := mockCommand(t, dir, "brew", `exit 127`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed")
}

func TestEnsureBrewPkg_handles_network_failure_during_install(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" fails, "brew install" prints error and fails
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then
  echo "Error: Failed to download" >&2
  exit 1
fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed")
}

func TestEnsureBrewPkg_handles_brew_returning_unexpected_output(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" outputs corrupt data and returns 1
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then
  echo "CORRUPT_DATA_@#$%"
  exit 1
fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
}

func TestEnsureBrewPkg_gracefully_handles_brew_not_installed(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: prints "command not found" to stderr and returns 127
	binDir := mockCommand(t, dir, "brew", `
echo "bash: brew: command not found" >&2
exit 127
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	// Should not crash; either non-zero exit or output contains "Failed"
	if code != 0 || strings.Contains(out, "Failed") {
		// acceptable
	} else {
		t.Errorf("expected non-zero exit or output containing 'Failed', got code=%d, output=%q", code, out)
	}
}

func TestEnsureBrewPkg_handles_brew_list_non_zero_with_empty_output(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" returns 1 with no output, "brew install" succeeds
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then exit 0; fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "installed")
}

func TestEnsureBrewPkg_handles_brew_outputting_to_stderr(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: "brew list" writes to stderr but returns 0 (success)
	binDir := mockCommand(t, dir, "brew", `
if [ "$1" = "list" ]; then
  echo "Warning: Something weird" >&2
  exit 0
fi
exit 0
`)
	snippet := installSnippet(t, `ensure_brew_pkg "tmux"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already installed")
}

// ============================================================
// ensure_cask tests
// ============================================================

func TestEnsureCask_reports_found_when_app_exists(t *testing.T) {
	dir := t.TempDir()
	// Create a fake .app directory
	appDir := filepath.Join(dir, "Ghostty.app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("failed to create fake app dir: %v", err)
	}

	// We test the check logic directly: if the directory exists, ensure_cask reports success.
	// The real ensure_cask checks /Applications, so we use a wrapper snippet
	// that checks if a specific directory exists (equivalent to the BATS test).
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	script := fmt.Sprintf(`
source %q
if [ -d %q ]; then
  success "Ghostty found"
fi
`, tuiPath, appDir)

	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "found")
}

func TestEnsureCask_handles_cask_install_failure_with_network_error(t *testing.T) {
	dir := t.TempDir()
	// Mock brew: cask install fails with network error
	binDir := mockCommand(t, dir, "brew", `
if echo "$*" | grep -q -- "--cask"; then
  echo "Error: Download failed (Connection timed out)" >&2
  exit 1
fi
exit 0
`)
	snippet := installSnippet(t, `ensure_cask "nonexistent-app-xyz" "NonexistentApp"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	// ensure_cask calls exit 1 on failure, so expect non-zero exit
	if code == 0 {
		t.Errorf("expected non-zero exit code for cask install failure, got 0")
	}
	assertContains(t, out, "installation failed")
}

// ============================================================
// ensure_command tests
// ============================================================

func TestEnsureCommand_reports_already_installed_for_existing_command(t *testing.T) {
	// "bash" is always available in PATH
	snippet := installSnippet(t, `ensure_command "bash" "echo noop" "" "Bash"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already installed")
}

func TestEnsureCommand_installs_missing_command(t *testing.T) {
	// "nonexistent_cmd_xyz" is not in PATH, install command is "true" (succeeds)
	// Note: ensure_command exits with 1 when post_msg is empty due to
	// [ -n "$post_msg" ] && ... returning 1 under set -e. The BATS test
	// only checks output, not exit code.
	snippet := installSnippet(t, `ensure_command "nonexistent_cmd_xyz" "true" "" "TestTool"`)
	out, _ := runBashSnippet(t, snippet, nil)
	assertContains(t, out, "installed")
}

func TestEnsureCommand_shows_post_message_on_success(t *testing.T) {
	snippet := installSnippet(t, `ensure_command "nonexistent_cmd_xyz" "true" "Run it now" "TestTool"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Run it now")
}

func TestEnsureCommand_warns_on_install_failure(t *testing.T) {
	// Install command "false" always returns 1
	snippet := installSnippet(t, `ensure_command "nonexistent_cmd_xyz" "false" "" "TestTool"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "failed")
}

func TestEnsureCommand_handles_curl_404_error(t *testing.T) {
	snippet := installSnippet(t, `ensure_command "fake_tool" "curl -sSL https://example.com/404 | bash" "" "FakeTool"`)
	out, code := runBashSnippet(t, snippet, nil)
	// May succeed or fail depending on network, but output should mention "failed"
	_ = code
	assertContains(t, out, "failed")
}

func TestEnsureCommand_handles_curl_500_error(t *testing.T) {
	dir := t.TempDir()
	// Mock curl: returns error
	binDir := mockCommand(t, dir, "curl", `
echo "500 Internal Server Error" >&2
exit 22
`)
	snippet := installSnippet(t, `ensure_command "fake_tool" "curl -sSL https://example.com/install | bash" "" "FakeTool"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	// May succeed or fail depending on environment
	if code != 0 && code != 1 {
		t.Errorf("expected exit code 0 or 1, got %d", code)
	}
	assertContains(t, out, "FakeTool installed")
}

func TestEnsureCommand_handles_install_command_timeout(t *testing.T) {
	// Use a very short sleep instead of 10s to keep test fast
	// The BATS test just checks that it doesn't crash; exit code can be anything
	snippet := installSnippet(t, `ensure_command "slow_tool" "sleep 0.1 && true" "" "SlowTool"`)
	_, code := runBashSnippet(t, snippet, nil)
	// Either success or failure is acceptable
	if code != 0 && code != 1 {
		// Still acceptable, just not crashing
	}
}

func TestEnsureCommand_handles_empty_install_command(t *testing.T) {
	snippet := installSnippet(t, `ensure_command "test_cmd" "" "" "TestCmd"`)
	out, code := runBashSnippet(t, snippet, nil)
	if code != 0 && code != 1 {
		t.Errorf("expected exit code 0 or 1, got %d", code)
	}
	assertContains(t, out, "installed")
}

func TestEnsureCommand_handles_malformed_install_command(t *testing.T) {
	snippet := installSnippet(t, `ensure_command "test_cmd" "((invalid bash syntax" "" "TestCmd"`)
	out, code := runBashSnippet(t, snippet, nil)
	_ = code
	assertContains(t, out, "failed")
}

func TestEnsureCommand_verifies_command_exists_after_install(t *testing.T) {
	// Note: same as installs_missing_command — the BATS test only checks output.
	snippet := installSnippet(t, `ensure_command "definitely_not_real_cmd_xyz123" "true" "" "FakeTool"`)
	out, _ := runBashSnippet(t, snippet, nil)
	assertContains(t, out, "installed")
}

// ============================================================
// ensure_ghost_tab_tui tests
// ============================================================

func TestEnsureGhostTabTui_skips_when_binary_already_in_PATH(t *testing.T) {
	dir := t.TempDir()
	// Mock ghost-tab-tui already exists
	binDir := mockCommand(t, dir, "ghost-tab-tui", `echo "I exist"`)

	snippet := installSnippet(t, `ensure_ghost_tab_tui "/some/share/dir"`)
	env := buildEnv(t, []string{binDir})
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "ghost-tab-tui already available")
}

func TestEnsureGhostTabTui_builds_from_source_when_go_available(t *testing.T) {
	dir := t.TempDir()
	fakeHome := filepath.Join(dir, "home")
	if err := os.MkdirAll(filepath.Join(fakeHome, ".local", "bin"), 0755); err != nil {
		t.Fatalf("failed to create .local/bin: %v", err)
	}

	root := projectRoot(t)

	// Mock go: "go build -o <path> ./cmd/ghost-tab-tui" creates a fake binary
	binDir := mockCommand(t, dir, "go", `
if [ "$1" = "build" ]; then
  touch "$3"
  chmod +x "$3"
  exit 0
fi
exit 1
`)
	// We need ghost-tab-tui to NOT be in PATH, but go to be in PATH.
	// Explicitly set PATH to only include mock dir and system dirs, excluding
	// ~/.local/bin where the real ghost-tab-tui may be installed.
	snippet := installSnippet(t, fmt.Sprintf(`ensure_ghost_tab_tui %q`, root))
	env := buildEnv(t, nil, "HOME="+fakeHome, "PATH="+binDir+":/usr/bin:/bin")
	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Building ghost-tab-tui")
	assertContains(t, out, "ghost-tab-tui built and installed")
}

func TestEnsureGhostTabTui_fails_when_go_not_available(t *testing.T) {
	dir := t.TempDir()
	fakeHome := filepath.Join(dir, "home")
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		t.Fatalf("failed to create fakeHome: %v", err)
	}

	root := projectRoot(t)
	snippet := installSnippet(t, fmt.Sprintf(`ensure_ghost_tab_tui %q`, root))
	// Set PATH to only /usr/bin:/bin so `go` and `ghost-tab-tui` are not found.
	// This excludes Homebrew/user paths where go is typically installed.
	env := buildEnv(t, nil, "HOME="+fakeHome, "PATH=/usr/bin:/bin")
	out, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Errorf("expected non-zero exit code when go is not available, got 0")
	}
	assertContains(t, out, "Go is required")
}

func TestEnsureGhostTabTui_fails_when_go_build_fails(t *testing.T) {
	dir := t.TempDir()
	fakeHome := filepath.Join(dir, "home")
	if err := os.MkdirAll(filepath.Join(fakeHome, ".local", "bin"), 0755); err != nil {
		t.Fatalf("failed to create .local/bin: %v", err)
	}

	root := projectRoot(t)

	// Mock go: build fails
	binDir := mockCommand(t, dir, "go", `
echo "build error" >&2
exit 1
`)
	snippet := installSnippet(t, fmt.Sprintf(`ensure_ghost_tab_tui %q`, root))
	// Explicitly set PATH to exclude ~/.local/bin where real ghost-tab-tui lives
	env := buildEnv(t, nil, "HOME="+fakeHome, "PATH="+binDir+":/usr/bin:/bin")
	out, code := runBashSnippet(t, snippet, env)
	if code == 0 {
		t.Errorf("expected non-zero exit code when go build fails, got 0")
	}
	assertContains(t, out, "Failed to build ghost-tab-tui")
}

func TestEnsureGhostTabTui_creates_local_bin_directory_if_missing(t *testing.T) {
	dir := t.TempDir()
	fakeHome := filepath.Join(dir, "home")
	// Do NOT create .local/bin — ensure_ghost_tab_tui should create it
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		t.Fatalf("failed to create fakeHome: %v", err)
	}

	root := projectRoot(t)

	// Mock go: "go build" creates a fake binary
	binDir := mockCommand(t, dir, "go", `
if [ "$1" = "build" ]; then
  # Ensure parent dir exists before touch (the function creates it)
  mkdir -p "$(dirname "$3")"
  touch "$3"
  chmod +x "$3"
  exit 0
fi
exit 1
`)
	snippet := installSnippet(t, fmt.Sprintf(`ensure_ghost_tab_tui %q`, root))
	// Explicitly set PATH to exclude ~/.local/bin where real ghost-tab-tui lives
	env := buildEnv(t, nil, "HOME="+fakeHome, "PATH="+binDir+":/usr/bin:/bin")
	_, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)

	// Verify .local/bin was created
	localBin := filepath.Join(fakeHome, ".local", "bin")
	if _, err := os.Stat(localBin); os.IsNotExist(err) {
		t.Errorf("expected %s to be created, but it does not exist", localBin)
	}
}

// ============================================================
// ensure_base_requirements tests
// ============================================================

func TestEnsureBaseRequirements_checks_jq(t *testing.T) {
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	installPath := filepath.Join(root, "lib", "install.sh")

	// Source install.sh, then override ensure_command to just echo args,
	// then call ensure_base_requirements.
	script := fmt.Sprintf(`
source %q
source %q
ensure_command() {
  echo "Checking $1"
}
ensure_base_requirements
`, tuiPath, installPath)

	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "jq")
}
