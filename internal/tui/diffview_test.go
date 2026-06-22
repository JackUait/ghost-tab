package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func sampleDiff(lines int) string {
	var b strings.Builder
	for i := 0; i < lines; i++ {
		b.WriteString("+added line ")
		b.WriteByte(byte('0' + (i % 10)))
		b.WriteByte('\n')
	}
	return b.String()
}

func sizeDiff(m DiffViewModel, w, h int) DiffViewModel {
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return updated.(DiffViewModel)
}

func keyDiff(m DiffViewModel, t tea.KeyType) (DiffViewModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: t})
	return updated.(DiffViewModel), cmd
}

func runeDiff(m DiffViewModel, r rune) (DiffViewModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return updated.(DiffViewModel), cmd
}

func quits(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestNewDiffView_stores_title_and_content(t *testing.T) {
	m := NewDiffView("lib/x.sh", "+hello\n")
	if m.title != "lib/x.sh" {
		t.Errorf("title = %q, want lib/x.sh", m.title)
	}
	if m.content != "+hello\n" {
		t.Errorf("content = %q, want +hello", m.content)
	}
}

func TestDiffView_Escape_quits(t *testing.T) {
	m := sizeDiff(NewDiffView("f", sampleDiff(5)), 80, 24)
	m2, cmd := keyDiff(m, tea.KeyEscape)
	if !m2.quitting {
		t.Error("Escape should set quitting")
	}
	if !quits(cmd) {
		t.Error("Escape should emit tea.Quit")
	}
}

func TestDiffView_q_quits(t *testing.T) {
	m := sizeDiff(NewDiffView("f", sampleDiff(5)), 80, 24)
	m2, cmd := runeDiff(m, 'q')
	if !m2.quitting {
		t.Error("q should set quitting")
	}
	if !quits(cmd) {
		t.Error("q should emit tea.Quit")
	}
}

func TestDiffView_CtrlC_quits(t *testing.T) {
	m := sizeDiff(NewDiffView("f", sampleDiff(5)), 80, 24)
	_, cmd := keyDiff(m, tea.KeyCtrlC)
	if !quits(cmd) {
		t.Error("ctrl+c should emit tea.Quit")
	}
}

func TestDiffView_View_shows_title_controls_and_content(t *testing.T) {
	m := sizeDiff(NewDiffView("lib/x.sh", "+added unique-marker\n"), 80, 24)
	out := m.View()
	if !strings.Contains(out, "lib/x.sh") {
		t.Error("view should show the title (filename)")
	}
	if !strings.Contains(out, "unique-marker") {
		t.Error("view should show the diff content")
	}
	// A control bar advertising how to scroll and close.
	if !strings.Contains(strings.ToLower(out), "scroll") {
		t.Error("view should show a scroll hint")
	}
	if !strings.Contains(strings.ToLower(out), "esc") {
		t.Error("view should advertise Esc to close")
	}
}

func TestDiffView_preserves_ansi_color_in_content(t *testing.T) {
	colored := "\x1b[32m+added\x1b[m\n\x1b[31m-removed\x1b[m\n"
	m := sizeDiff(NewDiffView("f", colored), 80, 24)
	out := m.View()
	if !strings.Contains(out, "\x1b[32m") || !strings.Contains(out, "\x1b[31m") {
		t.Error("view should preserve the diff's ANSI color escapes")
	}
}

func TestDiffView_scrolls_with_keys(t *testing.T) {
	// Content much taller than the viewport so there's room to scroll.
	m := sizeDiff(NewDiffView("f", sampleDiff(100)), 80, 10)
	if !m.viewport.AtTop() {
		t.Fatal("should start at top")
	}
	// Page down moves off the top.
	m, _ = keyDiff(m, tea.KeySpace)
	if m.viewport.AtTop() {
		t.Error("space (page down) should scroll off the top")
	}
	// G jumps to the bottom.
	m, _ = runeDiff(m, 'G')
	if !m.viewport.AtBottom() {
		t.Error("G should jump to the bottom")
	}
	// g jumps back to the top.
	m, _ = runeDiff(m, 'g')
	if !m.viewport.AtTop() {
		t.Error("g should jump back to the top")
	}
}
