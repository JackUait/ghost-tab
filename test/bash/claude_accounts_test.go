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

// add_claude_account reads a label from stdin, registers the account via the
// ghost-tab-tui CLI, runs `claude auth login` under the new account's isolated
// CLAUDE_CONFIG_DIR, and makes it the active account.
func TestAddClaudeAccount_registers_logs_in_and_activates(t *testing.T) {
	dir := t.TempDir()
	cfgRoot := filepath.Join(dir, "ghost-tab")
	if err := os.MkdirAll(cfgRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	// Mock ghost-tab-tui: simulate `claude-account add` by creating the account
	// dir and printing its name. Other invocations are ignored.
	bin := mockCommand(t, dir, "ghost-tab-tui", `
if [ "$1" = "claude-account" ] && [ "$2" = "add" ]; then
  mkdir -p "`+cfgRoot+`/claude-accounts/work"
  echo "work"
fi
`)
	// Mock claude: record argv + the inherited CLAUDE_CONFIG_DIR so the test can
	// assert login ran under the new account.
	mockCommand(t, dir, "claude", `echo "args:$*" > "`+dir+`/claude-call"
echo "cfgdir:$CLAUDE_CONFIG_DIR" >> "`+dir+`/claude-call"`)

	env := buildEnv(t, []string{bin})
	out, code := runBashFuncWithStdin(t, "lib/claude-accounts.sh", "add_claude_account",
		[]string{cfgRoot}, env, "Work\n")
	assertExitCode(t, code, 0)
	_ = out

	// claude auth login ran under the new account's config dir.
	call, _ := os.ReadFile(filepath.Join(dir, "claude-call"))
	if !strings.Contains(string(call), "args:auth login") {
		t.Errorf("expected `claude auth login`, got: %q", string(call))
	}
	if !strings.Contains(string(call), "cfgdir:"+filepath.Join(cfgRoot, "claude-accounts", "work")) {
		t.Errorf("login should run under the new CLAUDE_CONFIG_DIR, got: %q", string(call))
	}
	// The new account is now active.
	ptr := filepath.Join(cfgRoot, "claude-account")
	if b, _ := os.ReadFile(ptr); strings.TrimSpace(string(b)) != "work" {
		t.Errorf("new account should be active, pointer = %q", string(b))
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
