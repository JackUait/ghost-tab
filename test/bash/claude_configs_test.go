package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadClaudeConfigs_skips_comments_blanks(t *testing.T) {
	dir := t.TempDir()
	writeTempFile(t, dir, "list", "# header\n\nWork:work.json\nPersonal:personal.json\n")
	out, code := runBashFunc(t, "lib/claude-configs.sh", "load_claude_configs",
		[]string{filepath.Join(dir, "list")}, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Work:work.json")
	assertContains(t, out, "Personal:personal.json")
	assertNotContains(t, out, "header")
}

func TestActivePointer_get_set_and_standard_clears(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-config")
	if _, code := runBashFunc(t, "lib/claude-configs.sh", "set_active_claude_config",
		[]string{ptr, "work.json"}, nil); code != 0 {
		t.Fatalf("set failed")
	}
	out, _ := runBashFunc(t, "lib/claude-configs.sh", "get_active_claude_config", []string{ptr}, nil)
	assertContains(t, out, "work.json")
	if _, code := runBashFunc(t, "lib/claude-configs.sh", "set_active_claude_config",
		[]string{ptr, "standard"}, nil); code != 0 {
		t.Fatalf("set standard failed")
	}
	if _, err := os.Stat(ptr); !os.IsNotExist(err) {
		t.Fatalf("pointer should be removed for standard")
	}
}

func TestResolveClaudeConfigPath_existing_vs_missing(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "claude-configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTempFile(t, cfgDir, "work.json", "{}")
	ptr := filepath.Join(dir, "claude-config")
	writeTempFile(t, dir, "claude-config", "work.json")
	out, _ := runBashFunc(t, "lib/claude-configs.sh", "resolve_claude_config_path",
		[]string{cfgDir, ptr}, nil)
	if strings.TrimSpace(out) != filepath.Join(cfgDir, "work.json") {
		t.Fatalf("got %q", out)
	}
	writeTempFile(t, dir, "claude-config", "missing.json")
	out2, _ := runBashFunc(t, "lib/claude-configs.sh", "resolve_claude_config_path",
		[]string{cfgDir, ptr}, nil)
	if strings.TrimSpace(out2) != "" {
		t.Fatalf("expected empty for missing file, got %q", out2)
	}
}

func TestAddClaudeConfig_creates_file_and_list_line(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "claude-configs")
	list := filepath.Join(dir, "list")
	out, code := runBashFunc(t, "lib/claude-configs.sh", "add_claude_config",
		[]string{list, cfgDir, "My Work"}, nil)
	assertExitCode(t, code, 0)
	if strings.TrimSpace(out) != "my-work.json" {
		t.Fatalf("got filename %q", out)
	}
	data, _ := os.ReadFile(filepath.Join(cfgDir, "my-work.json"))
	if strings.TrimSpace(string(data)) != "{}" {
		t.Fatalf("file should contain {}")
	}
	listData, _ := os.ReadFile(list)
	assertContains(t, string(listData), "My Work:my-work.json")
}

func TestDeleteClaudeConfig_active_resets_pointer(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "claude-configs")
	_ = os.MkdirAll(cfgDir, 0o755)
	writeTempFile(t, cfgDir, "work.json", "{}")
	list := filepath.Join(dir, "list")
	writeTempFile(t, dir, "list", "Work:work.json\n")
	ptr := filepath.Join(dir, "claude-config")
	writeTempFile(t, dir, "claude-config", "work.json")
	if _, code := runBashFunc(t, "lib/claude-configs.sh", "delete_claude_config",
		[]string{list, cfgDir, ptr, "work.json"}, nil); code != 0 {
		t.Fatal("delete failed")
	}
	if _, err := os.Stat(filepath.Join(cfgDir, "work.json")); !os.IsNotExist(err) {
		t.Fatal("config file should be gone")
	}
	if _, err := os.Stat(ptr); !os.IsNotExist(err) {
		t.Fatal("pointer should be cleared when active config deleted")
	}
	listData, _ := os.ReadFile(list)
	assertNotContains(t, string(listData), "work.json")
}

func TestRenameClaudeConfig_changes_name_only(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "list")
	writeTempFile(t, dir, "list", "Work:work.json\n")
	if _, code := runBashFunc(t, "lib/claude-configs.sh", "rename_claude_config",
		[]string{list, "work.json", "Day Job"}, nil); code != 0 {
		t.Fatal("rename failed")
	}
	listData, _ := os.ReadFile(list)
	assertContains(t, string(listData), "Day Job:work.json")
	assertNotContains(t, string(listData), "Work:work.json")
}

func TestAddClaudeConfig_resolves_collisions(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "claude-configs")
	list := filepath.Join(dir, "list")

	// First add
	out1, code1 := runBashFunc(t, "lib/claude-configs.sh", "add_claude_config",
		[]string{list, cfgDir, "My Work"}, nil)
	assertExitCode(t, code1, 0)
	if strings.TrimSpace(out1) != "my-work.json" {
		t.Fatalf("first add: got %q, want my-work.json", strings.TrimSpace(out1))
	}

	// Second add — same name, collision
	out2, code2 := runBashFunc(t, "lib/claude-configs.sh", "add_claude_config",
		[]string{list, cfgDir, "My Work"}, nil)
	assertExitCode(t, code2, 0)
	if strings.TrimSpace(out2) != "my-work-2.json" {
		t.Fatalf("second add: got %q, want my-work-2.json", strings.TrimSpace(out2))
	}

	// Both files must exist
	if _, err := os.Stat(filepath.Join(cfgDir, "my-work.json")); err != nil {
		t.Fatal("my-work.json missing")
	}
	if _, err := os.Stat(filepath.Join(cfgDir, "my-work-2.json")); err != nil {
		t.Fatal("my-work-2.json missing")
	}
}

func TestRenameClaudeConfig_missing_file_returns_error(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "list")
	writeTempFile(t, dir, "list", "Work:work.json\n")

	_, code := runBashFunc(t, "lib/claude-configs.sh", "rename_claude_config",
		[]string{list, "nonexistent.json", "New Name"}, nil)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing file, got 0")
	}

	// List unchanged
	listData, _ := os.ReadFile(list)
	assertContains(t, string(listData), "Work:work.json")
	assertNotContains(t, string(listData), "New Name")
}
