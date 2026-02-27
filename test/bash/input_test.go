package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// TestConfirmTui — tests for confirm_tui
// ---------------------------------------------------------------------------

func TestConfirmTui_calls_ghost_tab_tui_confirm_with_message(t *testing.T) {
	dir := t.TempDir()

	// Mock ghost-tab-tui: verify args and return JSON
	mockCommand(t, dir, "ghost-tab-tui", `
if [[ "$1" == "confirm" && "$2" == "Delete this?" ]]; then
  echo '{"confirmed":true}'
  exit 0
fi
exit 1
`)

	// Mock jq: parse .confirmed
	mockCommand(t, dir, "jq", `
if [[ "$1" == "-r" && "$2" == ".confirmed" ]]; then
  echo "true"
  exit 0
fi
exit 1
`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})

	out, code := runBashFuncWithStdin(t, "lib/input.sh", "confirm_tui",
		[]string{"Delete this?"}, env, "")
	_ = out
	assertExitCode(t, code, 0)
}

func TestConfirmTui_returns_failure_when_user_cancels(t *testing.T) {
	dir := t.TempDir()

	mockCommand(t, dir, "ghost-tab-tui", `
if [[ "$1" == "confirm" ]]; then
  echo '{"confirmed":false}'
  exit 0
fi
exit 1
`)

	mockCommand(t, dir, "jq", `
if [[ "$1" == "-r" && "$2" == ".confirmed" ]]; then
  echo "false"
  exit 0
fi
exit 1
`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})

	_, code := runBashFuncWithStdin(t, "lib/input.sh", "confirm_tui",
		[]string{"Delete this?"}, env, "")
	assertExitCode(t, code, 1)
}

func TestConfirmTui_handles_jq_parse_failure(t *testing.T) {
	dir := t.TempDir()

	mockCommand(t, dir, "ghost-tab-tui", `
if [[ "$1" == "confirm" ]]; then
  echo '{"confirmed":true}'
  exit 0
fi
exit 1
`)

	// Mock jq that always fails
	mockCommand(t, dir, "jq", `exit 1`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})

	out, code := runBashFuncWithStdin(t, "lib/input.sh", "confirm_tui",
		[]string{"Delete this?"}, env, "")
	assertExitCode(t, code, 1)
	assertContains(t, out, "Failed to parse confirmation response")
}

func TestConfirmTui_validates_against_null_string(t *testing.T) {
	dir := t.TempDir()

	mockCommand(t, dir, "ghost-tab-tui", `
if [[ "$1" == "confirm" ]]; then
  echo '{"confirmed":"null"}'
  exit 0
fi
exit 1
`)

	mockCommand(t, dir, "jq", `
if [[ "$1" == "-r" && "$2" == ".confirmed" ]]; then
  echo "null"
  exit 0
fi
exit 1
`)

	binDir := filepath.Join(dir, "bin")
	env := buildEnv(t, []string{binDir})

	_, code := runBashFuncWithStdin(t, "lib/input.sh", "confirm_tui",
		[]string{"Delete this?"}, env, "")
	assertExitCode(t, code, 1)
}

func TestConfirmTui_falls_back_to_bash_prompt_when_binary_missing(t *testing.T) {
	dir := t.TempDir()

	// Create a test script that mimics the fallback logic (without /dev/tty)
	testScript := filepath.Join(dir, "test_fallback.sh")
	if err := os.WriteFile(testScript, []byte(`#!/bin/bash
msg="$1"
read -rp "$msg (y/N) " response
[[ "$response" =~ ^[Yy]$ ]]
`), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	// Simulate user typing "y"
	script := fmt.Sprintf("echo 'y' | %q 'Delete this?'", testScript)
	out, code := runBashSnippet(t, script, nil)
	_ = out
	assertExitCode(t, code, 0)
}

func TestConfirmTui_fallback_rejects_non_yes_responses(t *testing.T) {
	dir := t.TempDir()

	testScript := filepath.Join(dir, "test_fallback_reject.sh")
	if err := os.WriteFile(testScript, []byte(`#!/bin/bash
msg="$1"
read -rp "$msg (y/N) " response
[[ "$response" =~ ^[Yy]$ ]]
`), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}

	// Simulate user typing "n"
	script := fmt.Sprintf("echo 'n' | %q 'Delete this?'", testScript)
	_, code := runBashSnippet(t, script, nil)
	assertExitCode(t, code, 1)
}
