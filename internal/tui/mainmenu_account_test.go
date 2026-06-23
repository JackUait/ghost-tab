package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func acctTestMenu(tool string) *MainMenuModel {
	projects := []models.Project{
		{Name: "alpha", Path: "/tmp/alpha"},
		{Name: "beta", Path: "/tmp/beta"},
	}
	m := NewMainMenu(projects, []string{"claude", "opencode"}, tool, "animated")
	m.SetSize(100, 40)
	return m
}

// The LOGIN row is always shown as a peer of AGENT/PLAN. With no managed
// accounts it shows "Default" and renders no cycle chevrons (nothing to switch
// to), mirroring the PLAN row's behaviour when only Standard exists.
func TestAccountRow_alwaysShown_noChevronsWhenOnlyDefault(t *testing.T) {
	m := acctTestMenu("claude")
	if got := m.accountRowCount(); got != 1 {
		t.Fatalf("accountRowCount: got %d, want 1", got)
	}
	if m.accountHasChoices() {
		t.Errorf("no managed accounts should mean no choices/chevrons")
	}
	row := stripAnsi(m.renderAccountRow("│", "│"))
	if !strings.Contains(row, "LOGIN") || !strings.Contains(row, "Default") {
		t.Errorf("LOGIN row should show Default: %q", row)
	}
	if strings.Contains(row, "◂") || strings.Contains(row, "▸") {
		t.Errorf("no chevrons expected with only Default: %q", row)
	}
}

// The LOGIN row renders at the very top — above the AGENT row — and shows
// chevrons once a managed account exists.
func TestAccountRow_atTop_withChevronsWhenAccountsExist(t *testing.T) {
	m := acctTestMenu("claude")
	m.SetClaudeAccounts([]ClaudeAccount{{Label: "Work", Dir: "work"}})
	if !m.accountHasChoices() {
		t.Fatalf("one managed account should offer choices")
	}
	out := stripAnsi(m.renderMenuBox())
	loginIdx := strings.Index(out, "LOGIN")
	agentIdx := strings.Index(out, "AGENT")
	if loginIdx < 0 || agentIdx < 0 {
		t.Fatalf("LOGIN/AGENT rows missing:\n%s", out)
	}
	if !(loginIdx < agentIdx) {
		t.Errorf("LOGIN row must be above AGENT row (login=%d agent=%d)", loginIdx, agentIdx)
	}
	row := stripAnsi(m.renderAccountRow("│", "│"))
	if !strings.Contains(row, "◂") || !strings.Contains(row, "▸") {
		t.Errorf("chevrons expected with a managed account: %q", row)
	}
}

// Enter on the focused LOGIN row exits the menu with the add-account action so
// wrapper.sh can run the interactive `claude auth login`.
func TestAccountRow_enterTriggersAddAccount(t *testing.T) {
	m := acctTestMenu("claude")
	m.focus = FocusAccount
	m.focusEnter()
	r := m.Result()
	if r == nil || r.Action != "add-account" {
		t.Fatalf("Enter on LOGIN row should set action add-account, got %+v", r)
	}
}

func TestAccount_setActiveByDir_andLabels(t *testing.T) {
	m := acctTestMenu("claude")
	m.SetClaudeAccounts([]ClaudeAccount{
		{Label: "Work", Dir: "work"},
		{Label: "Personal", Dir: "personal"},
	})
	if m.CurrentClaudeAccountLabel() != "Default" || m.CurrentClaudeAccountDir() != "" {
		t.Fatalf("initial should be Default, got %q/%q", m.CurrentClaudeAccountLabel(), m.CurrentClaudeAccountDir())
	}
	m.SetActiveClaudeAccount("personal")
	if m.CurrentClaudeAccountLabel() != "Personal" || m.CurrentClaudeAccountDir() != "personal" {
		t.Errorf("got %q/%q", m.CurrentClaudeAccountLabel(), m.CurrentClaudeAccountDir())
	}
	// Unknown dir falls back to Default.
	m.SetActiveClaudeAccount("ghost")
	if m.CurrentClaudeAccountDir() != "" {
		t.Errorf("unknown dir should reset to Default, got %q", m.CurrentClaudeAccountDir())
	}
}

// The LOGIN row is always a focus stop (it hosts Enter-to-add even with no
// managed accounts), but only offers choices/chevrons once an account exists.
func TestAccount_alwaysFocusable_choicesGated(t *testing.T) {
	m := acctTestMenu("claude")
	if !m.accountFocusable() {
		t.Errorf("LOGIN row should always be focusable for the add affordance")
	}
	if m.accountHasChoices() {
		t.Errorf("no choices expected with no managed accounts")
	}
	m.SetClaudeAccounts([]ClaudeAccount{{Label: "Work", Dir: "work"}})
	if !m.accountHasChoices() {
		t.Errorf("choices expected with one account (Default + Work)")
	}
}

// Cycling walks Default → managed accounts → Default and persists the active
// dir to the pointer file (Default removes the pointer).
func TestAccount_cyclePersistsPointer(t *testing.T) {
	dir := t.TempDir()
	ptr := filepath.Join(dir, "claude-account")
	m := acctTestMenu("claude")
	m.SetClaudeAccounts([]ClaudeAccount{
		{Label: "Work", Dir: "work"},
		{Label: "Personal", Dir: "personal"},
	})
	m.SetClaudeAccountFile(ptr)

	m.CycleAccount("next") // Default -> Work
	if got := m.CurrentClaudeAccountDir(); got != "work" {
		t.Fatalf("after next: got %q, want work", got)
	}
	if b, _ := os.ReadFile(ptr); strings.TrimSpace(string(b)) != "work" {
		t.Errorf("pointer should be 'work', got %q", string(b))
	}

	m.CycleAccount("next") // Work -> Personal
	m.CycleAccount("next") // Personal -> Default (wrap)
	if got := m.CurrentClaudeAccountDir(); got != "" {
		t.Fatalf("after wrap: got %q, want Default", got)
	}
	if _, err := os.Stat(ptr); !os.IsNotExist(err) {
		t.Errorf("Default should remove the pointer file")
	}

	m.CycleAccount("prev") // Default -> Personal (wrap back)
	if got := m.CurrentClaudeAccountDir(); got != "personal" {
		t.Errorf("after prev wrap: got %q, want personal", got)
	}
}

// The always-present LOGIN row + subscription row push the first project to row
// 8: top, LOGIN, title, subscription, switcher-gap, tab bar, separator, leading
// blank (8) → first project at row 8. Click mapping must follow the same offset.
func TestMapRowToItem_accountsForAccountRow(t *testing.T) {
	m := acctTestMenu("claude")
	if got := m.MapRowToItem(7); got != -1 {
		t.Errorf("row 7 should be the leading blank (-1) with the LOGIN row present, got %d", got)
	}
	if got := m.MapRowToItem(8); got != 0 {
		t.Errorf("first project should be at row 8 with the LOGIN row, got %d", got)
	}
}
