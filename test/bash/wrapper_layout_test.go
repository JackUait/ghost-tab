package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWrapper_terminal_pane_is_45_percent verifies the left column's vertical
// split gives the bottom terminal pane 45% of the height. The whole
// "new-session ... \; split-window ..." chain is one tmux invocation, so the
// mock records all of it via $* and we can assert the split percentage.
func TestWrapper_terminal_pane_is_45_percent(t *testing.T) {
	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}

	recPath := filepath.Join(home, "rec")
	mocks := map[string]string{
		"tmux":          "#!/bin/bash\nif [ \"$1\" = \"new-session\" ]; then printf '%s\\n' \"$*\" > \"$GT_REC\"; exit 0; fi\nexit 0\n",
		"claude":        "#!/bin/bash\nexit 0\n",
		"lazygit":       "#!/bin/bash\nexit 0\n",
		"ghost-tab-tui": "#!/bin/bash\nexit 0\n",
	}
	for name, body := range mocks {
		p := filepath.Join(binDir, name)
		if err := os.WriteFile(p, []byte(body), 0755); err != nil {
			t.Fatalf("write mock %s: %v", name, err)
		}
	}

	projDir := filepath.Join(home, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("mkdir proj: %v", err)
	}

	env := buildEnv(t, nil, "HOME="+home, "GT_REC="+recPath)
	_, code := runBashScript(t, "wrapper.sh", []string{"--restore", projDir, "claude"}, env)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(recPath)
	if err != nil {
		t.Fatalf("new-session was never invoked (no record): %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "split-window -v -p 45") {
		t.Errorf("terminal pane should be split at 45%%; got tmux args:\n%s", got)
	}
}

// recordWrapperNewSession runs wrapper.sh with a tmux mock that records the
// whole "new-session ... \; ..." chain (one invocation, captured via $*) and
// returns that recorded argument string.
func recordWrapperNewSession(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	recPath := filepath.Join(home, "rec")
	mocks := map[string]string{
		"tmux":          "#!/bin/bash\nif [ \"$1\" = \"new-session\" ]; then printf '%s\\n' \"$*\" > \"$GT_REC\"; exit 0; fi\nexit 0\n",
		"claude":        "#!/bin/bash\nexit 0\n",
		"lazygit":       "#!/bin/bash\nexit 0\n",
		"ghost-tab-tui": "#!/bin/bash\nexit 0\n",
	}
	for name, body := range mocks {
		p := filepath.Join(binDir, name)
		if err := os.WriteFile(p, []byte(body), 0755); err != nil {
			t.Fatalf("write mock %s: %v", name, err)
		}
	}
	projDir := filepath.Join(home, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("mkdir proj: %v", err)
	}
	env := buildEnv(t, nil, "HOME="+home, "GT_REC="+recPath)
	_, code := runBashScript(t, "wrapper.sh", []string{"--restore", projDir, "claude"}, env)
	assertExitCode(t, code, 0)
	data, err := os.ReadFile(recPath)
	if err != nil {
		t.Fatalf("new-session was never invoked (no record): %v", err)
	}
	return string(data)
}

// TestWrapper_selects_ai_pane_geometrically verifies the wrapper focuses panes
// by direction (-L / -R) instead of fixed indices. tmux routes external
// drag-drops (e.g. a screenshot) to the ACTIVE pane, so the AI pane must end up
// active for a dropped screenshot to land in the AI tool. Fixed indices
// (select-pane -t 0 / -t 2) silently target the wrong pane under a non-zero
// pane-base-index; directional selection is robust to any base-index.
func TestWrapper_selects_ai_pane_geometrically(t *testing.T) {
	got := recordWrapperNewSession(t)
	if !strings.Contains(got, "select-pane -L") {
		t.Errorf("expected directional 'select-pane -L' to focus the left column; got:\n%s", got)
	}
	if !strings.Contains(got, "select-pane -R") {
		t.Errorf("expected directional 'select-pane -R' to leave the AI (right) pane active; got:\n%s", got)
	}
	if strings.Contains(got, "select-pane -t 0") || strings.Contains(got, "select-pane -t 2") {
		t.Errorf("fixed-index select-pane breaks under non-zero pane-base-index; use directional selection. got:\n%s", got)
	}
}

// TestWrapper_spare_pane_runs_tabbed_tmux verifies the spare bottom-left pane
// launches a nested tmux (the tab bar) instead of a bare shell, and that the
// tab keybindings (add/close) are wired on the outer session.
func TestWrapper_spare_pane_runs_tabbed_tmux(t *testing.T) {
	got := recordWrapperNewSession(t)
	if !strings.Contains(got, "split-window -v -p 45") {
		t.Fatalf("expected the spare pane split; got:\n%s", got)
	}
	for _, want := range []string{
		"env -u TMUX -u TMUX_PANE tmux -L gtspare_", // nested server, $TMUX shed
		"new-session",              // the inner session that hosts the tabs
		"|| exec bash",             // graceful fallback if tmux is unavailable
		"bind-key t ",              // keyboard: add a tab
		"bind-key w ",              // keyboard: close current tab
		"spare_tabs_close_current", // close routes through the guarded helper
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected new-session chain to contain %q; got:\n%s", want, got)
		}
	}
}

// TestWrapper_marks_ai_pane locks in the @gt_ai marker on the AI pane, which
// lib/screenshot.sh uses to resolve the AI pane for prefix+i injection.
func TestWrapper_marks_ai_pane(t *testing.T) {
	got := recordWrapperNewSession(t)
	if !strings.Contains(got, "set-option -p @gt_ai 1") {
		t.Errorf("expected the AI pane to be marked with '@gt_ai 1'; got:\n%s", got)
	}
}

// TestWrapper_active_pane_border_is_visible verifies the active pane has a
// distinct border. Without this, the active and inactive borders look
// identical, so a user can't tell which pane is focused -- and a screenshot
// dropped onto a non-AI active pane silently fails to reach the AI tool.
func TestWrapper_active_pane_border_is_visible(t *testing.T) {
	got := recordWrapperNewSession(t)
	if !strings.Contains(got, "pane-active-border-style") {
		t.Errorf("expected new-session to set a distinct pane-active-border-style; got:\n%s", got)
	}
}

// TestWrapper_pane_dividers_match_tab_bar verifies the pane-border background is
// tinted to the spare tab bar's charcoal (colour235). The divider above the
// spare pane is its own row; tinting its background makes the tab bar appear
// flush against the divider instead of separated by a dark gap.
func TestWrapper_pane_dividers_match_tab_bar(t *testing.T) {
	got := recordWrapperNewSession(t)
	for _, want := range []string{
		"pane-border-style fg=colour238,bg=colour235",
		"pane-active-border-style fg=colour209,bg=colour235",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected pane border tinted to the tab-bar charcoal: %q; got:\n%s", want, got)
		}
	}
}
