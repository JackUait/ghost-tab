package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An "account" is a native Claude login isolated by its own CLAUDE_CONFIG_DIR.
// Storage mirrors claude-configs: <root>/claude-accounts/<dir>/ holds the login,
// named in <root>/claude-accounts.list (label:dir), with the active dir name in
// <root>/claude-account. The Default account (empty/absent pointer) means the
// user's standard ~/.claude login (Keychain), so no CLAUDE_CONFIG_DIR is set.

func TestLoadClaudeAccounts_skips_comments_blanks(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "list", "# header\n\nWork:work\nPersonal:personal\n")
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "load_claude_accounts",
		[]string{filepath.Join(dir, "list")}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Work:work")
	assertContains(t, out, "Personal:personal")
	assertNotContains(t, out, "header")
}

func TestActiveAccountPointer_get_set_and_default_clears(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	if _, code := runBashFunc(t, "lib/claude-accounts.sh", "set_active_claude_account",
		[]string{ptr, "work"}, nil); code != 0 {
		t.Fatalf("set failed")
	}
	out, _ := runBashFunc(t, "lib/claude-accounts.sh", "get_active_claude_account", []string{ptr}, nil)
	assertContains(t, out, "work")
	if _, code := runBashFunc(t, "lib/claude-accounts.sh", "set_active_claude_account",
		[]string{ptr, "default"}, nil); code != 0 {
		t.Fatalf("set default failed")
	}
	if _, err := os.Stat(ptr); !os.IsNotExist(err) {
		t.Fatalf("pointer should be removed for default")
	}
}

func TestGetActiveAccount_default_when_no_pointer(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "get_active_claude_account", []string{ptr}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Fatalf("expected empty (Default) for no pointer, got %q", out)
	}
}

func TestResolveClaudeAccountDir_existing_vs_missing(t *testing.T) {
	dir := t.TempDir()
	acctRoot := filepath.Join(dir, "claude-accounts")
	if err := os.MkdirAll(filepath.Join(acctRoot, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
	ptr := filepath.Join(dir, "claude-account")
	writeTempFile(t, dir, "claude-account", "work")
	out, _ := runBashFunc(t, "lib/claude-accounts.sh", "resolve_claude_account_dir",
		[]string{acctRoot, ptr}, nil)
	if strings.TrimSpace(out) != filepath.Join(acctRoot, "work") {
		t.Fatalf("got %q", out)
	}
	// Missing dir → empty (falls back to Default/Keychain).
	writeTempFile(t, dir, "claude-account", "missing")
	out2, _ := runBashFunc(t, "lib/claude-accounts.sh", "resolve_claude_account_dir",
		[]string{acctRoot, ptr}, nil)
	if strings.TrimSpace(out2) != "" {
		t.Fatalf("expected empty for missing dir, got %q", out2)
	}
}

func TestResolveClaudeAccountDir_default_is_empty(t *testing.T) {
	dir := t.TempDir()
	acctRoot := filepath.Join(dir, "claude-accounts")
	ptr := filepath.Join(dir, "claude-account") // absent → Default
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "resolve_claude_account_dir",
		[]string{acctRoot, ptr}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "" {
		t.Fatalf("Default account must resolve to empty (Keychain), got %q", out)
	}
}

// get_active_claude_account_name maps the active pointer to its display label so
// the compact-view ledger / menu can show which account is in use. Default (no
// pointer) reads as "Default".
func TestActiveAccountName_default_when_no_pointer(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	list := filepath.Join(dir, "claude-accounts.list")
	writeTempFile(t, dir, "claude-accounts.list", "Work:work\n")
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "get_active_claude_account_name",
		[]string{ptr, list}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Default" {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), "Default")
	}
}

func TestActiveAccountName_maps_active_dir_to_list_label(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	list := filepath.Join(dir, "claude-accounts.list")
	writeTempFile(t, dir, "claude-account", "work")
	writeTempFile(t, dir, "claude-accounts.list", "Work Max:work\nPersonal:personal\n")
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "get_active_claude_account_name",
		[]string{ptr, list}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Work Max" {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), "Work Max")
	}
}

func TestActiveAccountName_unknown_dir_falls_back_to_default(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	list := filepath.Join(dir, "claude-accounts.list")
	writeTempFile(t, dir, "claude-account", "ghost")
	writeTempFile(t, dir, "claude-accounts.list", "Work:work\n")
	out, code := runBashFunc(t, "lib/claude-accounts.sh", "get_active_claude_account_name",
		[]string{ptr, list}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "Default" {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), "Default")
	}
}
