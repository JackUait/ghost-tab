package bash_test

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWrapperRestore_skips_picker_and_resumes runs the real wrapper.sh in
// --restore mode against mocked binaries and verifies it reaches new-session
// directly (no picker), forces the tool, stamps the project path, and applies
// the resume launch flag.
//
// wrapper.sh line 2 resets PATH to start with "$HOME/.local/bin", so mocks
// must live there and HOME must be overridden to our temp dir.
func TestWrapperRestore_skips_picker_and_resumes(t *testing.T) {
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
		"wisp-deck-tui": "#!/bin/bash\nexit 0\n",
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
	assertContains(t, got, "WISP_DECK=1")
	assertContains(t, got, "WISP_DECK_TOOL=claude")
	assertContains(t, got, "WISP_DECK_PATH="+projDir)
	assertContains(t, got, "claude -c")
}
