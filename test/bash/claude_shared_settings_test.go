package bash_test

import (
	"os"
	"path/filepath"
	"testing"
)

// A native Claude account is isolated by its own CLAUDE_CONFIG_DIR, so by default
// it sees NONE of the standard ~/.claude login's settings (status line, permission
// mode, skills, hooks, model, …). sync_claude_shared_settings symlinks a curated
// allowlist of *settings* items from the standard login's config dir into the
// per-account dir so all logins share one set of settings, while each keeps its
// own credentials/identity/session state.

// writeFile is a tiny helper that creates a file with content under dir/name,
// making parent dirs as needed.
func writeSharedFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runSync invokes the helper with (source_dir, account_dir).
func runSync(t *testing.T, source, account string) (string, int) {
	t.Helper()
	return runBashFunc(t, "lib/claude-shared-settings.sh", "sync_claude_shared_settings",
		[]string{source, account}, nil)
}

func TestSyncSharedSettings_symlinks_existing_items(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	account := filepath.Join(dir, "account")
	// Settings file, statusline script, and a skills directory all live in source.
	writeSharedFile(t, filepath.Join(source, "settings.json"), `{"model":"opus"}`)
	writeSharedFile(t, filepath.Join(source, "statusline-wrapper.sh"), "echo hi")
	writeSharedFile(t, filepath.Join(source, "skills", "demo", "SKILL.md"), "x")
	if err := os.MkdirAll(account, 0o755); err != nil {
		t.Fatal(err)
	}

	_, code := runSync(t, source, account)
	assertExitCode(t, code, 0)

	for _, item := range []string{"settings.json", "statusline-wrapper.sh", "skills"} {
		dest := filepath.Join(account, item)
		target, err := os.Readlink(dest)
		if err != nil {
			t.Fatalf("%s should be a symlink: %v", item, err)
		}
		if target != filepath.Join(source, item) {
			t.Fatalf("%s links to %q, want %q", item, target, filepath.Join(source, item))
		}
	}
	// The skills dir resolves through the link to the real content.
	if _, err := os.Stat(filepath.Join(account, "skills", "demo", "SKILL.md")); err != nil {
		t.Fatalf("skills content not reachable through link: %v", err)
	}
}

func TestSyncSharedSettings_skips_items_absent_in_source(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	account := filepath.Join(dir, "account")
	writeSharedFile(t, filepath.Join(source, "settings.json"), "{}")
	if err := os.MkdirAll(account, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, code := runSync(t, source, account); code != 0 {
		t.Fatalf("sync failed")
	}
	// commands/ never existed in source, so nothing is created in the account.
	if _, err := os.Lstat(filepath.Join(account, "commands")); !os.IsNotExist(err) {
		t.Fatalf("absent source item should not be linked")
	}
}

func TestSyncSharedSettings_replaces_account_local_copy(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	account := filepath.Join(dir, "account")
	writeSharedFile(t, filepath.Join(source, "settings.json"), `{"shared":true}`)
	// The account starts with its own real settings.json that must be superseded.
	writeSharedFile(t, filepath.Join(account, "settings.json"), `{"theme":"dark"}`)

	if _, code := runSync(t, source, account); code != 0 {
		t.Fatalf("sync failed")
	}
	target, err := os.Readlink(filepath.Join(account, "settings.json"))
	if err != nil {
		t.Fatalf("account settings.json should become a symlink: %v", err)
	}
	if target != filepath.Join(source, "settings.json") {
		t.Fatalf("links to %q, want source", target)
	}
}

func TestSyncSharedSettings_is_idempotent(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	account := filepath.Join(dir, "account")
	writeSharedFile(t, filepath.Join(source, "settings.json"), "{}")
	if err := os.MkdirAll(account, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, code := runSync(t, source, account); code != 0 {
			t.Fatalf("sync run %d failed", i)
		}
	}
	if _, err := os.Readlink(filepath.Join(account, "settings.json")); err != nil {
		t.Fatalf("settings.json should still be a symlink after re-run: %v", err)
	}
}

func TestSyncSharedSettings_preserves_account_credentials_and_state(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	account := filepath.Join(dir, "account")
	writeSharedFile(t, filepath.Join(source, "settings.json"), "{}")
	// Per-account identity / runtime state must never be touched.
	writeSharedFile(t, filepath.Join(account, ".credentials.json"), "secret")
	writeSharedFile(t, filepath.Join(account, ".claude.json"), "identity")
	writeSharedFile(t, filepath.Join(account, "history.jsonl"), "line")
	writeSharedFile(t, filepath.Join(account, "projects", "p", "x"), "y")
	// Source also has its own credentials that must NOT leak into the account.
	writeSharedFile(t, filepath.Join(source, ".credentials.json"), "OTHER")

	if _, code := runSync(t, source, account); code != 0 {
		t.Fatalf("sync failed")
	}
	got, _ := os.ReadFile(filepath.Join(account, ".credentials.json"))
	if string(got) != "secret" {
		t.Fatalf("account credentials clobbered: %q", got)
	}
	if fi, err := os.Lstat(filepath.Join(account, ".credentials.json")); err != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf(".credentials.json must remain a real file, not a link")
	}
	for _, item := range []string{".claude.json", "history.jsonl", "projects"} {
		if _, err := os.Lstat(filepath.Join(account, item)); err != nil {
			t.Fatalf("account state %s should be preserved: %v", item, err)
		}
	}
}

func TestSyncSharedSettings_noop_when_source_equals_account(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "standard")
	writeSharedFile(t, filepath.Join(source, "settings.json"), `{"x":1}`)

	if _, code := runSync(t, source, source); code != 0 {
		t.Fatalf("sync failed")
	}
	// settings.json must remain the original real file, not be rm'd/relinked to itself.
	if fi, err := os.Lstat(filepath.Join(source, "settings.json")); err != nil || fi.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("self-sync must leave the file untouched")
	}
}

func TestSyncSharedSettings_noop_when_dirs_missing(t *testing.T) {
	dir := t.TempDir()
	// Missing source dir.
	if _, code := runSync(t, filepath.Join(dir, "nope"), dir); code != 0 {
		t.Fatalf("missing source should be a clean no-op")
	}
	// Missing account dir.
	source := filepath.Join(dir, "standard")
	writeSharedFile(t, filepath.Join(source, "settings.json"), "{}")
	if _, code := runSync(t, source, filepath.Join(dir, "nope")); code != 0 {
		t.Fatalf("missing account should be a clean no-op")
	}
}
