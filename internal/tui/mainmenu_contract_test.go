package tui

import (
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func TestContract_actionStringsUnchanged(t *testing.T) {
	want := []string{"add-project", "delete-project", "open-once", "plain-terminal"}
	if len(actionNames) != len(want) {
		t.Fatalf("actionNames len = %d, want %d", len(actionNames), len(want))
	}
	for i, a := range want {
		if actionNames[i] != a {
			t.Errorf("actionNames[%d] = %q, want %q", i, actionNames[i], a)
		}
	}
}

func TestContract_selectProjectEmitsAction(t *testing.T) {
	projects := []models.Project{{Name: "blok", Path: "/tmp/blok"}}
	m := NewMainMenu(projects, []string{"claude"}, "claude", "none")
	m.selectCurrent()
	r := m.Result()
	if r == nil || r.Action != "select-project" || r.Name != "blok" {
		t.Fatalf("select result = %+v, want action=select-project name=blok", r)
	}
}

func TestContract_plainTerminalAction(t *testing.T) {
	m := NewMainMenu(nil, []string{"claude"}, "claude", "none")
	m.handleRune('p')
	r := m.Result()
	if r == nil || r.Action != "plain-terminal" {
		t.Fatalf("plain result = %+v, want action=plain-terminal", r)
	}
}
