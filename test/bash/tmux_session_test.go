package bash_test

import (
	"strings"
	"testing"
)

func aiCmd(t *testing.T, tool string, resume bool) string {
	t.Helper()
	var env []string
	// WISP_DECK_RESUME_SESSION is force-cleared to stay hermetic when the test
	// itself runs inside a restored Wisp Deck session (which exports both).
	if resume {
		env = buildEnv(t, nil, "WISP_DECK_RESUME=1", "WISP_DECK_RESUME_SESSION=")
	} else {
		env = buildEnv(t, nil, "WISP_DECK_RESUME=0", "WISP_DECK_RESUME_SESSION=")
	}
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{tool, "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	return strings.TrimSpace(out)
}

func TestBuildAiLaunchCmd_resume_flags(t *testing.T) {
	cases := []struct {
		tool string
		want string
	}{
		{"claude", "claude -c"},
		{"opencode", "npx opencode-ai@latest --continue"},
	}
	for _, c := range cases {
		if got := aiCmd(t, c.tool, true); got != c.want {
			t.Errorf("resume %s: got %q, want %q", c.tool, got, c.want)
		}
	}
}

func TestBuildAiLaunchCmd_resumes_specific_claude_session(t *testing.T) {
	// Each restored tab must reopen ITS conversation: two tabs of the same
	// project resumed with plain `claude -c` would both open the project's
	// most recent conversation.
	env := buildEnv(t, nil, "WISP_DECK_RESUME=1", "WISP_DECK_RESUME_SESSION=sid-42")
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"claude", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	if got := strings.TrimSpace(out); got != "claude --resume sid-42" {
		t.Errorf("got %q, want %q", got, "claude --resume sid-42")
	}

	// OpenCode has its own continue semantics; the Claude session id must
	// not leak into its command.
	out, code = runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"opencode", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	if got := strings.TrimSpace(out); got != "npx opencode-ai@latest --continue" {
		t.Errorf("opencode got %q, want %q", got, "npx opencode-ai@latest --continue")
	}
}

func TestBuildAiLaunchCmd_normal_unaffected(t *testing.T) {
	if got := aiCmd(t, "opencode", false); got != `npx opencode-ai@latest "/p/app"` {
		t.Errorf("normal opencode: got %q", got)
	}
}

// When WISP_DECK_CLAUDE_FILTER is set (wrapper sets it after confirming the TUI
// binary supports the screenshot-drag filter), the Claude launch is prefixed
// with it so a dropped screenshot's temp path is rewritten to a stable copy
// before Claude reads it. OpenCode is never wrapped.
func TestBuildAiLaunchCmd_wraps_claude_with_filter(t *testing.T) {
	env := buildEnv(t, nil, "WISP_DECK_RESUME=0",
		"WISP_DECK_CLAUDE_FILTER=wisp-deck-tui screenshot-filter -- ")
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"claude", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	if got := strings.TrimSpace(out); got != `wisp-deck-tui screenshot-filter -- claude /p/app` {
		t.Errorf("claude wrap: got %q", got)
	}
	out, _ = runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"opencode", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	if strings.Contains(out, "screenshot-filter") {
		t.Errorf("opencode must not be wrapped: %q", strings.TrimSpace(out))
	}
}

func TestBuildAiLaunchCmd_wraps_claude_resume_with_filter(t *testing.T) {
	env := buildEnv(t, nil, "WISP_DECK_RESUME=1",
		"WISP_DECK_CLAUDE_FILTER=wisp-deck-tui screenshot-filter -- ")
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"claude", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	if got := strings.TrimSpace(out); got != `wisp-deck-tui screenshot-filter -- claude -c` {
		t.Errorf("claude resume wrap: got %q", got)
	}
}

// When a non-Default native account is active, wrapper.sh exports
// WISP_DECK_CLAUDE_ACCOUNT_DIR and the Claude launch is prefixed with
// CLAUDE_CONFIG_DIR=<dir> so `claude` runs under that account's isolated login.
// The Default account leaves the env var unset (Keychain login, unchanged).
func TestBuildAiLaunchCmd_prefixes_claude_config_dir(t *testing.T) {
	env := buildEnv(t, nil, "WISP_DECK_RESUME=0",
		"WISP_DECK_CLAUDE_ACCOUNT_DIR=/cfg/claude-accounts/work")
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"claude", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	if got := strings.TrimSpace(out); got != `CLAUDE_CONFIG_DIR="/cfg/claude-accounts/work" claude /p/app` {
		t.Errorf("claude account prefix: got %q", got)
	}
}

func TestBuildAiLaunchCmd_account_dir_not_applied_to_opencode(t *testing.T) {
	env := buildEnv(t, nil, "WISP_DECK_RESUME=0",
		"WISP_DECK_CLAUDE_ACCOUNT_DIR=/cfg/claude-accounts/work")
	out, _ := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"opencode", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	if strings.Contains(out, "CLAUDE_CONFIG_DIR") {
		t.Errorf("opencode must not get CLAUDE_CONFIG_DIR: %q", strings.TrimSpace(out))
	}
}

// The account prefix composes ahead of the screenshot filter (env is inherited
// by the child claude) and survives resume mode.
func TestBuildAiLaunchCmd_account_dir_composes_with_filter_and_resume(t *testing.T) {
	env := buildEnv(t, nil, "WISP_DECK_RESUME=1",
		"WISP_DECK_CLAUDE_ACCOUNT_DIR=/cfg/claude-accounts/work",
		"WISP_DECK_CLAUDE_FILTER=wisp-deck-tui screenshot-filter -- ")
	out, code := runBashFunc(t, "lib/tmux-session.sh", "build_ai_launch_cmd",
		[]string{"claude", "claude", "npx opencode-ai@latest", "/p/app"}, env)
	assertExitCode(t, code, 0)
	want := `CLAUDE_CONFIG_DIR="/cfg/claude-accounts/work" wisp-deck-tui screenshot-filter -- claude -c`
	if got := strings.TrimSpace(out); got != want {
		t.Errorf("claude account+filter+resume: got %q, want %q", got, want)
	}
}
