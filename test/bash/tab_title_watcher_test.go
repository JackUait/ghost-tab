package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tabTitleSnippet sources tui.sh and tab-title-watcher.sh, then runs the provided bash code.
func tabTitleSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	watcherPath := filepath.Join(root, "lib", "tab-title-watcher.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, watcherPath, body)
}

// --- check_ai_tool_state: Claude with marker file ---

func TestTabTitleWatcher_check_ai_tool_state_claude_returns_waiting_when_marker_exists_and_prompt_visible(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker")
	os.WriteFile(markerFile, []byte(""), 0644)
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'Some output\n> \n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "claude" "dev-test-123" %q %q`, tmuxPath, markerFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "waiting" {
		t.Errorf("expected 'waiting', got %q", strings.TrimSpace(out))
	}
}

func TestTabTitleWatcher_check_ai_tool_state_claude_returns_active_when_marker_exists_but_no_prompt(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker")
	os.WriteFile(markerFile, []byte(""), 0644)
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'Processing request...\nGenerating code\n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "claude" "dev-test-123" %q %q`, tmuxPath, markerFile))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "active" {
		t.Errorf("expected 'active', got %q", strings.TrimSpace(out))
	}
}

func TestTabTitleWatcher_check_ai_tool_state_claude_returns_active_when_marker_absent(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "claude" "" "" %q`, markerFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "active" {
		t.Errorf("expected 'active', got %q", strings.TrimSpace(out))
	}
}

// --- check_ai_tool_state: non-Claude with mock tmux ---

func TestTabTitleWatcher_check_ai_tool_state_codex_returns_waiting_when_prompt_detected(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'some output\n❯ \n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "codex" "dev-test-123" %q ""`, tmuxPath))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "waiting" {
		t.Errorf("expected 'waiting', got %q", strings.TrimSpace(out))
	}
}

func TestTabTitleWatcher_check_ai_tool_state_codex_returns_active_when_no_prompt(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'Processing request...\nGenerating code\n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "codex" "dev-test-123" %q ""`, tmuxPath))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "active" {
		t.Errorf("expected 'active', got %q", strings.TrimSpace(out))
	}
}

func TestTabTitleWatcher_check_ai_tool_state_detects_dollar_prompt(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'Welcome to copilot\n$ \n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "copilot" "dev-test-123" %q ""`, tmuxPath))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "waiting" {
		t.Errorf("expected 'waiting', got %q", strings.TrimSpace(out))
	}
}

func TestTabTitleWatcher_check_ai_tool_state_detects_gt_prompt(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := mockCommand(t, tmpDir, "tmux", `
if [ "$1" = "capture-pane" ]; then
  printf 'Ready\n> \n'
  exit 0
fi
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "opencode" "dev-test-123" %q ""`, tmuxPath))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "waiting" {
		t.Errorf("expected 'waiting', got %q", strings.TrimSpace(out))
	}
}

// --- check_ai_tool_state: pane targeting ---

func TestTabTitleWatcher_check_ai_tool_state_targets_correct_pane(t *testing.T) {
	tmpDir := t.TempDir()
	// Mock tmux that only returns a prompt for pane 0.3
	binDir := mockCommand(t, tmpDir, "tmux", `
for arg in "$@"; do
  if [ "$arg" = "dev-test-123:0.3" ]; then
    printf 'Some output\n❯ \n'
    exit 0
  fi
done
printf 'no prompt here\n'
exit 0
`)
	env := buildEnv(t, []string{binDir})
	tmuxPath := filepath.Join(binDir, "tmux")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`check_ai_tool_state "codex" "dev-test-123" %q "" "3"`, tmuxPath))

	out, code := runBashSnippet(t, snippet, env)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "waiting" {
		t.Errorf("expected 'waiting' (pane 0.3 targeted), got %q", strings.TrimSpace(out))
	}
}

// --- stop_tab_title_watcher: cleanup ---

func TestTabTitleWatcher_stop_tab_title_watcher_removes_marker_file(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "marker")
	os.WriteFile(markerFile, []byte(""), 0644)

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`_TAB_TITLE_WATCHER_PID=""; stop_tab_title_watcher %q`, markerFile))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Errorf("expected marker file to be removed")
	}
}

func TestTabTitleWatcher_stop_tab_title_watcher_succeeds_when_marker_absent(t *testing.T) {
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "no-such-marker")

	snippet := tabTitleSnippet(t,
		fmt.Sprintf(`_TAB_TITLE_WATCHER_PID=""; stop_tab_title_watcher %q`, markerFile))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
}
