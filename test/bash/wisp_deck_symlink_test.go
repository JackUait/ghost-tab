package bash_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstaller_creates_wisp_deck_symlink(t *testing.T) {
	dir := t.TempDir()
	localBin := filepath.Join(dir, ".local", "bin")
	if err := os.MkdirAll(localBin, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a fake wisp-deck-config to link to
	src := filepath.Join(dir, "wisp-deck-config")
	if err := os.WriteFile(src, []byte("#!/bin/bash\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Run the symlink creation snippet
	script := `
		HOME="` + dir + `"
		SCRIPT_PATH="` + src + `"
		LOCAL_BIN="$HOME/.local/bin"
		mkdir -p "$LOCAL_BIN"
		ln -sf "$SCRIPT_PATH" "$LOCAL_BIN/wisp-deck"
	`
	_, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 0)

	// Verify symlink exists and points to the right place
	linkPath := filepath.Join(localBin, "wisp-deck")
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if target != src {
		t.Errorf("symlink target = %q, want %q", target, src)
	}
}
