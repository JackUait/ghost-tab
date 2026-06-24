package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeAccountCmd_Registered(t *testing.T) {
	for _, path := range [][]string{
		{"claude-account"},
		{"claude-account", "add"},
		{"claude-account", "remove"},
	} {
		cmd, _, err := rootCmd.Find(path)
		if err != nil {
			t.Fatalf("Find(%v): %v", path, err)
		}
		if cmd.Name() != path[len(path)-1] {
			t.Errorf("Find(%v) resolved to %q", path, cmd.Name())
		}
	}
}

func TestClaudeAccountCmd_AddRemove(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "claude-accounts.list")
	acctDir := filepath.Join(dir, "claude-accounts")
	ptr := filepath.Join(dir, "claude-account")

	out := execRoot(t, "claude-account", "add", "--list", list, "--accounts-dir", acctDir, "--label", "Work Max")
	if strings.TrimSpace(out) != "work-max" {
		t.Fatalf("add printed %q, want work-max", strings.TrimSpace(out))
	}
	if info, err := os.Stat(filepath.Join(acctDir, "work-max")); err != nil || !info.IsDir() {
		t.Fatal("account dir not created")
	}
	data, _ := os.ReadFile(list)
	if !strings.Contains(string(data), "Work Max:work-max") {
		t.Fatalf("list entry not written: %q", data)
	}

	os.WriteFile(ptr, []byte("work-max\n"), 0644)
	execRoot(t, "claude-account", "remove", "--list", list, "--accounts-dir", acctDir, "--pointer", ptr, "--dir", "work-max")
	if _, err := os.Stat(filepath.Join(acctDir, "work-max")); !os.IsNotExist(err) {
		t.Fatal("account dir not removed")
	}
	if _, err := os.Stat(ptr); !os.IsNotExist(err) {
		t.Fatal("pointer not cleared after removing active account")
	}
}
