package bash_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWispDeckConfig_runs_config_menu(t *testing.T) {
	dir := t.TempDir()
	// Mock wisp-deck-tui to return quit immediately
	binDir := mockCommand(t, dir, "wisp-deck-tui", `echo '{"action":"quit"}'`)
	// Mock jq for JSON parsing
	mockCommand(t, dir, "jq", `
		if [ "$1" = "-r" ] && [ "$2" = ".action" ]; then
			read -r input
			action="${input#*\"action\":\"}"
			action="${action%%\"*}"
			echo "$action"
		fi
	`)
	// Set HOME to temp dir so script's PATH="$HOME/.local/bin:..." doesn't find real binary
	env := buildEnv(t, []string{binDir}, "HOME="+dir)
	_, code := runBashScript(t, "bin/wisp-deck-config", nil, env)
	assertExitCode(t, code, 0)
}

func TestWispDeckConfig_exits_nonzero_when_tui_missing(t *testing.T) {
	dir := t.TempDir()
	emptyBin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(emptyBin, 0755); err != nil {
		t.Fatal(err)
	}
	env := buildEnv(t, nil, "PATH="+emptyBin+":/usr/bin:/bin:/usr/sbin:/sbin")
	_, code := runBashScript(t, "bin/wisp-deck-config", nil, env)
	if code == 0 {
		t.Error("expected non-zero exit when wisp-deck-tui is missing")
	}
}
