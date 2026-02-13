package bash_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// releaseSnippet builds a bash snippet that sources scripts/release.sh
// (which has a source guard so main doesn't run), then runs the provided bash code.
func releaseSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	releasePath := filepath.Join(root, "scripts", "release.sh")
	return fmt.Sprintf("source %q && %s", releasePath, body)
}

// initGitRepo creates a minimal git repo in dir with one commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git setup failed: %v\n%s", err, out)
		}
	}
	writeTempFile(t, dir, "dummy", "init")
	for _, args := range [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "init"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit failed: %v\n%s", err, out)
		}
	}
}

// ============================================================
// check_clean_tree tests
// ============================================================

func TestCheckCleanTree_passes_on_clean_repo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_clean_tree`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	_ = out
}

func TestCheckCleanTree_fails_on_dirty_repo(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeTempFile(t, dir, "untracked.txt", "dirty")

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_clean_tree`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 1)
	assertContains(t, out, "clean")
}

// ============================================================
// check_main_branch tests
// ============================================================

func TestCheckMainBranch_passes_on_main(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_main_branch`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	_ = out
}

func TestCheckMainBranch_fails_on_other_branch(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.CombinedOutput()

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_main_branch`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 1)
	assertContains(t, out, "main")
}

// ============================================================
// read_version tests
// ============================================================

func TestReadVersion_reads_valid_semver(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "VERSION", "2.0.0\n")

	snippet := releaseSnippet(t, `read_version "`+filepath.Join(dir, "VERSION")+`"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "2.0.0" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "2.0.0")
	}
}

func TestReadVersion_fails_on_missing_file(t *testing.T) {
	snippet := releaseSnippet(t, `read_version "/nonexistent/VERSION"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 1)
	assertContains(t, out, "VERSION")
}

func TestReadVersion_fails_on_invalid_format(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "VERSION", "not-a-version\n")

	snippet := releaseSnippet(t, `read_version "`+filepath.Join(dir, "VERSION")+`"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 1)
	assertContains(t, out, "semver")
}

// ============================================================
// check_tag_not_exists tests
// ============================================================

func TestCheckTagNotExists_passes_when_tag_missing(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_tag_not_exists "v2.0.0"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	_ = out
}

func TestCheckTagNotExists_fails_when_tag_exists(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	cmd := exec.Command("git", "tag", "v1.0.0")
	cmd.Dir = dir
	cmd.CombinedOutput()

	snippet := releaseSnippet(t, `cd "`+dir+`" && check_tag_not_exists "v1.0.0"`)
	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 1)
	assertContains(t, out, "v1.0.0")
}
