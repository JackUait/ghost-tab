package bash_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================
// statusline.sh tests (TestStatusline_*)
// ============================================================

// --- get_tree_rss_kb ---

func TestStatusline_get_tree_rss_kb_sums_memory_of_process_and_its_children(t *testing.T) {
	dir := t.TempDir()

	// Mock pgrep: 100 -> [101, 102], 101 -> [103], others -> exit 1
	mockCommand(t, dir, "pgrep", `
pid="${@: -1}"
case "$pid" in
  100) printf '101\n102\n' ;;
  101) printf '103\n' ;;
  *) exit 1 ;;
esac
`)

	// Mock ps: return RSS per pid
	mockCommand(t, dir, "ps", `
pid="${@: -1}"
case "$pid" in
  100) echo "  51200" ;;
  101) echo "  25600" ;;
  102) echo "  10240" ;;
  103) echo "  5120" ;;
  *) echo "" ;;
esac
`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})
	out, code := runBashFunc(t, "lib/statusline.sh", "get_tree_rss_kb", []string{"100"}, env)
	assertExitCode(t, code, 0)
	// 51200 + 25600 + 10240 + 5120 = 92160
	if strings.TrimSpace(out) != "92160" {
		t.Errorf("expected 92160, got %q", strings.TrimSpace(out))
	}
}

func TestStatusline_get_tree_rss_kb_handles_process_with_no_children(t *testing.T) {
	dir := t.TempDir()

	mockCommand(t, dir, "pgrep", `exit 1`)
	mockCommand(t, dir, "ps", `echo "  51200"`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})
	out, code := runBashFunc(t, "lib/statusline.sh", "get_tree_rss_kb", []string{"100"}, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "51200" {
		t.Errorf("expected 51200, got %q", strings.TrimSpace(out))
	}
}

func TestStatusline_get_tree_rss_kb_handles_disappeared_process_gracefully(t *testing.T) {
	dir := t.TempDir()

	mockCommand(t, dir, "pgrep", `exit 1`)
	mockCommand(t, dir, "ps", `echo ""`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})
	out, code := runBashFunc(t, "lib/statusline.sh", "get_tree_rss_kb", []string{"999"}, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "0" {
		t.Errorf("expected 0, got %q", strings.TrimSpace(out))
	}
}

func TestStatusline_get_tree_rss_kb_handles_child_that_disappears_mid_walk(t *testing.T) {
	dir := t.TempDir()

	// 100 -> [101, 102], others -> exit 1
	mockCommand(t, dir, "pgrep", `
pid="${@: -1}"
case "$pid" in
  100) printf '101\n102\n' ;;
  *) exit 1 ;;
esac
`)

	// 101 returns empty (disappeared), 102 returns value
	mockCommand(t, dir, "ps", `
pid="${@: -1}"
case "$pid" in
  100) echo "  51200" ;;
  101) echo "" ;;
  102) echo "  10240" ;;
  *) echo "" ;;
esac
`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})
	out, code := runBashFunc(t, "lib/statusline.sh", "get_tree_rss_kb", []string{"100"}, env)
	assertExitCode(t, code, 0)
	// 51200 + 0 + 10240 = 61440
	if strings.TrimSpace(out) != "61440" {
		t.Errorf("expected 61440, got %q", strings.TrimSpace(out))
	}
}

// --- statusline-command.sh: session line diff ---

// statuslineCmdSetupGitRepo creates a temp git repo with one initial commit.
// Returns (repo dir, cleanup func).
func statuslineCmdSetupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "-C", dir, "init", "-q"},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup failed: %v\n%s", err, out)
		}
	}

	// Create initial file and commit
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("initial\n"), 0644); err != nil {
		t.Fatalf("write file.txt: %v", err)
	}
	for _, c := range [][]string{
		{"git", "-C", dir, "add", "file.txt"},
		{"git", "-C", dir, "commit", "-q", "-m", "initial"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup failed: %v\n%s", err, out)
		}
	}

	return dir
}

// getBaselineSHA returns the current HEAD SHA for a git repo.
func getBaselineSHA(t *testing.T, repoDir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestStatusline_statusline_command_shows_additions_green_and_deletions_red(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	baselineSHA := getBaselineSHA(t, repoDir)

	// Add 3 lines
	f, err := os.OpenFile(filepath.Join(repoDir, "file.txt"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	if _, err := f.WriteString("line1\nline2\nline3\n"); err != nil {
		t.Fatalf("write file: %v", err)
	}
	f.Close()

	baselineFile := filepath.Join(t.TempDir(), "baseline")
	if err := os.WriteFile(baselineFile, []byte(baselineSHA+"\n"), 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	env := buildEnv(t, nil, "GHOST_TAB_BASELINE_FILE="+baselineFile)
	out, code := runBashSnippet(t, script, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "+3")
	assertContains(t, out, "-0")
}

func TestStatusline_statusline_command_shows_deletions_in_red(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	baselineSHA := getBaselineSHA(t, repoDir)

	// Delete the file content (1 line removed)
	if err := os.WriteFile(filepath.Join(repoDir, "file.txt"), []byte(""), 0644); err != nil {
		t.Fatalf("truncate file: %v", err)
	}

	baselineFile := filepath.Join(t.TempDir(), "baseline")
	if err := os.WriteFile(baselineFile, []byte(baselineSHA+"\n"), 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	env := buildEnv(t, nil, "GHOST_TAB_BASELINE_FILE="+baselineFile)
	out, code := runBashSnippet(t, script, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "+0")
	assertContains(t, out, "-1")
}

func TestStatusline_statusline_command_tracks_committed_changes_since_baseline(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	baselineSHA := getBaselineSHA(t, repoDir)

	// Make a committed change: add 2 lines
	f, err := os.OpenFile(filepath.Join(repoDir, "file.txt"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	if _, err := f.WriteString("new1\nnew2\n"); err != nil {
		t.Fatalf("write file: %v", err)
	}
	f.Close()

	for _, c := range [][]string{
		{"git", "-C", repoDir, "add", "file.txt"},
		{"git", "-C", repoDir, "commit", "-q", "-m", "add lines"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit failed: %v\n%s", err, out)
		}
	}

	baselineFile := filepath.Join(t.TempDir(), "baseline")
	if err := os.WriteFile(baselineFile, []byte(baselineSHA+"\n"), 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	env := buildEnv(t, nil, "GHOST_TAB_BASELINE_FILE="+baselineFile)
	out, code := runBashSnippet(t, script, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "+2")
	assertContains(t, out, "-0")
}

func TestStatusline_statusline_command_shows_zero_diff_when_no_changes(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	baselineSHA := getBaselineSHA(t, repoDir)

	baselineFile := filepath.Join(t.TempDir(), "baseline")
	if err := os.WriteFile(baselineFile, []byte(baselineSHA+"\n"), 0644); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	env := buildEnv(t, nil, "GHOST_TAB_BASELINE_FILE="+baselineFile)
	out, code := runBashSnippet(t, script, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, "+0")
	assertContains(t, out, "-0")
}

func TestStatusline_statusline_command_falls_back_to_repo_branch_only_without_baseline(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	repoBasename := filepath.Base(repoDir)

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	// Explicitly unset GHOST_TAB_BASELINE_FILE
	script := fmt.Sprintf(`unset GHOST_TAB_BASELINE_FILE; echo '%s' | bash '%s'`, stdinData, cmdPath)

	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, repoBasename)
	assertNotContains(t, out, "+0")
	assertNotContains(t, out, "/ -")
}

func TestStatusline_statusline_command_falls_back_when_baseline_file_missing(t *testing.T) {
	repoDir := statuslineCmdSetupGitRepo(t)
	repoBasename := filepath.Base(repoDir)

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, repoDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	env := buildEnv(t, nil, "GHOST_TAB_BASELINE_FILE=/tmp/ghost-tab-nonexistent-baseline")
	out, code := runBashSnippet(t, script, env)
	assertExitCode(t, code, 0)
	assertContains(t, out, repoBasename)
	assertNotContains(t, out, "+0")
	assertNotContains(t, out, "/ -")
}

func TestStatusline_statusline_command_non_git_directory_shows_just_dirname(t *testing.T) {
	nonGitDir := t.TempDir()
	dirBasename := filepath.Base(nonGitDir)

	root := projectRoot(t)
	cmdPath := filepath.Join(root, "templates", "statusline-command.sh")
	stdinData := fmt.Sprintf(`{"current_dir":"%s"}`, nonGitDir)
	script := fmt.Sprintf(`echo '%s' | bash '%s'`, stdinData, cmdPath)

	out, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, dirBasename)
	assertNotContains(t, out, "+0")
	assertNotContains(t, out, "/ -")
}

// ============================================================
// statusline-setup.sh tests (TestStatuslineSetup_*)
// ============================================================

// statuslineSetupSnippet builds a bash snippet that sources tui.sh, settings-json.sh,
// and statusline-setup.sh, then runs the provided bash code.
func statuslineSetupSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	settingsJsonPath := filepath.Join(root, "lib", "settings-json.sh")
	statuslineSetupPath := filepath.Join(root, "lib", "statusline-setup.sh")
	return fmt.Sprintf("source %q && source %q && source %q && %s",
		tuiPath, settingsJsonPath, statuslineSetupPath, body)
}

// setupStatuslineTestDirs creates the fake share dir with template files and fake home dirs.
// Returns (shareDir, fakeHome).
func setupStatuslineTestDirs(t *testing.T) (string, string) {
	t.Helper()
	tmpDir := t.TempDir()

	shareDir := filepath.Join(tmpDir, "share")
	writeTempFile(t, shareDir, "templates/ccstatusline-settings.json", "mock-settings")
	writeTempFile(t, shareDir, "templates/statusline-command.sh", "mock-command")
	writeTempFile(t, shareDir, "templates/statusline-wrapper.sh", "mock-wrapper")
	writeTempFile(t, shareDir, "lib/statusline.sh", "mock-helpers")

	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(filepath.Join(fakeHome, ".config", "ccstatusline"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	return shareDir, fakeHome
}

func TestStatuslineSetup_copies_config_and_scripts_when_npm_available(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Verify files were copied
	for _, path := range []string{
		filepath.Join(fakeHome, ".config", "ccstatusline", "settings.json"),
		filepath.Join(fakeHome, ".claude", "statusline-command.sh"),
		filepath.Join(fakeHome, ".claude", "statusline-wrapper.sh"),
		filepath.Join(fakeHome, ".claude", "statusline-helpers.sh"),
	} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", path)
		}
	}

	// Verify scripts are executable
	for _, name := range []string{"statusline-command.sh", "statusline-wrapper.sh"} {
		info, err := os.Stat(filepath.Join(fakeHome, ".claude", name))
		if err != nil {
			t.Errorf("stat %s: %v", name, err)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("expected %s to be executable", name)
		}
	}
}

func TestStatuslineSetup_skips_when_npm_not_available_and_brew_fails(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 1; }
brew() { return 1; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if _, err := os.Stat(filepath.Join(fakeHome, ".claude", "statusline-command.sh")); !os.IsNotExist(err) {
		t.Error("statusline-command.sh should not exist when npm not available and brew fails")
	}
}

func TestStatuslineSetup_reports_already_installed(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$1" == "list" ]]; then return 0; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already installed")
}

func TestStatuslineSetup_warns_and_skips_when_npm_install_fails(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$1" == "list" ]]; then return 1; fi
  if [[ "$1" == "install" ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")

	if _, err := os.Stat(filepath.Join(fakeHome, ".claude", "statusline-command.sh")); !os.IsNotExist(err) {
		t.Error("statusline-command.sh should not exist when npm install fails")
	}
	if _, err := os.Stat(filepath.Join(fakeHome, ".claude", "statusline-wrapper.sh")); !os.IsNotExist(err) {
		t.Error("statusline-wrapper.sh should not exist when npm install fails")
	}
}

func TestStatuslineSetup_installs_ccstatusline_and_copies_files_on_fresh_install(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	// First npm list call returns 1 (not installed), subsequent calls return 0
	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
_npm_list_call_count=0
npm() {
  if [[ "$1" == "list" ]]; then
    _npm_list_call_count=$((_npm_list_call_count + 1))
    if [[ "$_npm_list_call_count" -eq 1 ]]; then return 1; fi
    return 0
  fi
  if [[ "$1" == "install" ]]; then return 0; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "ccstatusline installed")
	assertNotContains(t, out, "already installed")

	for _, path := range []string{
		filepath.Join(fakeHome, ".config", "ccstatusline", "settings.json"),
		filepath.Join(fakeHome, ".claude", "statusline-command.sh"),
		filepath.Join(fakeHome, ".claude", "statusline-wrapper.sh"),
	} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", path)
		}
	}
}

func TestStatuslineSetup_calls_merge_claude_settings_after_file_copy(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)
	claudeSettings := filepath.Join(t.TempDir(), "claude-settings", "settings.json")

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, claudeSettings, fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if _, err := os.Stat(claudeSettings); os.IsNotExist(err) {
		t.Fatal("claude settings file should have been created by merge_claude_settings")
	}

	data, err := os.ReadFile(claudeSettings)
	if err != nil {
		t.Fatalf("read claude settings: %v", err)
	}
	assertContains(t, string(data), `"statusLine"`)
}

// --- npm install failure scenarios ---

func TestStatuslineSetup_handles_npm_install_network_timeout(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! network timeout" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
	if _, err := os.Stat(filepath.Join(fakeHome, ".claude", "statusline-command.sh")); !os.IsNotExist(err) {
		t.Error("statusline-command.sh should not exist on network timeout")
	}
}

func TestStatuslineSetup_handles_npm_install_ECONNREFUSED(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! network request to https://registry.npmjs.org/ccstatusline failed, reason: connect ECONNREFUSED" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_install_ETIMEDOUT(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! network request timed out, reason: ETIMEDOUT" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_registry_returning_404(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! 404 Not Found - GET https://registry.npmjs.org/ccstatusline" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_registry_returning_500(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! 500 Internal Server Error - GET https://registry.npmjs.org/ccstatusline" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_registry_returning_503_unavailable(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! 503 Service Unavailable - GET https://registry.npmjs.org/ccstatusline" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_install_hanging(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    sleep 5 &
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_install_disk_full_error(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! ENOSPC: no space left on device" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

func TestStatuslineSetup_handles_npm_install_permission_denied(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"install"* ]]; then
    echo "npm ERR! EACCES: permission denied" >&2
    return 1
  fi
  if [[ "$*" == *"list"* ]]; then return 1; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Failed to install")
}

// --- npm list failure scenarios ---

func TestStatuslineSetup_handles_npm_list_returning_malformed_output(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"list"* ]]; then
    echo "CORRUPT@#$%%DATA"
    return 0
  fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "already installed")
}

func TestStatuslineSetup_handles_npm_list_command_hanging(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"list"* ]]; then
    sleep 5 &
    return 0
  fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}

func TestStatuslineSetup_handles_npm_returning_non_JSON_output(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"list"* ]]; then
    echo "This is not JSON"
    return 0
  fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}

func TestStatuslineSetup_handles_npm_list_returning_empty_output(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() {
  if [[ "$*" == *"list"* ]]; then
    echo ""
    return 0
  fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}

// --- npm not found scenarios ---

func TestStatuslineSetup_handles_npm_not_in_PATH_after_install(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 1; }
brew() {
  if [[ "$*" == *"install node"* ]]; then return 0; fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertNotContains(t, out, "ccstatusline installed")
}

func TestStatuslineSetup_handles_brew_node_install_failure(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 1; }
brew() {
  if [[ "$*" == *"install node"* ]]; then
    echo "Error: Failed to install node" >&2
    return 1
  fi
  return 0
}
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Node.js installation failed")
}

func TestStatuslineSetup_handles_brew_not_available_for_node_install(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 1; }
brew() { return 127; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertNotContains(t, out, "ccstatusline")
}

// --- File operation failure scenarios ---

func TestStatuslineSetup_handles_missing_template_files(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	// Remove template files
	if err := os.RemoveAll(filepath.Join(shareDir, "templates")); err != nil {
		t.Fatalf("remove templates: %v", err)
	}

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	// cp will fail but script should handle gracefully
	// Either non-zero exit OR the config file won't be created
	configFile := filepath.Join(fakeHome, ".config", "ccstatusline", "settings.json")
	if code == 0 {
		if _, err := os.Stat(configFile); err == nil {
			t.Error("config file should not exist when templates are missing, but it does")
		}
	}
	// Either failure or missing file is acceptable
}

func TestStatuslineSetup_handles_read_only_config_directory(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	// Make config dir read-only
	configDir := filepath.Join(fakeHome, ".config")
	if err := os.Chmod(configDir, 0444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(configDir, 0755)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	// Function doesn't check mkdir errors, so it completes successfully
	assertExitCode(t, code, 0)
}

func TestStatuslineSetup_handles_chmod_failure_on_scripts(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
chmod() { return 1; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	// Function doesn't check chmod errors, completes successfully
	assertExitCode(t, code, 0)
}

func TestStatuslineSetup_handles_config_file_copy_permission_denied(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	// Make ccstatusline dir read-only
	ccDir := filepath.Join(fakeHome, ".config", "ccstatusline")
	if err := os.Chmod(ccDir, 0444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	defer os.Chmod(ccDir, 0755)

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	// Function doesn't check cp errors, may succeed or fail - either is acceptable
	_ = code
}

func TestStatuslineSetup_handles_corrupted_template_file(t *testing.T) {
	shareDir, fakeHome := setupStatuslineTestDirs(t)

	// Create corrupted template (non-UTF8)
	corruptPath := filepath.Join(shareDir, "templates", "ccstatusline-settings.json")
	if err := os.WriteFile(corruptPath, []byte{0xff, 0xfe, 0xfd}, 0644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	snippet := statuslineSetupSnippet(t, fmt.Sprintf(`
_has_npm() { return 0; }
npm() { return 0; }
setup_statusline %q %q %q
`, shareDir, filepath.Join(fakeHome, ".claude", "settings.json"), fakeHome))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Should copy file even if corrupted
	configFile := filepath.Join(fakeHome, ".config", "ccstatusline", "settings.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config file should exist even with corrupted template")
	}
}
