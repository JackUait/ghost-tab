package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogo_TickAdvancesFrame(t *testing.T) {
	m := NewLogo("claude")
	// Send a logoTickMsg
	updated, cmd := m.Update(logoTickMsg(time.Now()))
	result := updated.(LogoModel)
	if result.frame != 1 {
		t.Errorf("After tick: expected frame 1, got %d", result.frame)
	}
	// Should return a new tick command since not quitting
	if cmd == nil {
		t.Error("Tick should return a new tick command when not quitting")
	}
}

func TestLogo_TickWhileQuitting(t *testing.T) {
	m := NewLogo("claude")
	m.quitting = true
	updated, cmd := m.Update(logoTickMsg(time.Now()))
	result := updated.(LogoModel)
	_ = result
	if cmd != nil {
		t.Error("Tick should return nil command when quitting")
	}
}

func TestLogo_QuitMsg(t *testing.T) {
	m := NewLogo("claude")
	updated, cmd := m.Update(quitMsg{})
	result := updated.(LogoModel)
	if !result.quitting {
		t.Error("quitMsg should set quitting to true")
	}
	if cmd == nil {
		t.Error("quitMsg should return tea.Quit")
	}
}

func TestLogo_ViewEmptyWhenQuitting(t *testing.T) {
	m := NewLogo("claude")
	m.quitting = true
	if m.View() != "" {
		t.Error("View should be empty when quitting")
	}
}

func TestLogoModel_Update_stores_window_size(t *testing.T) {
	m := NewLogo("claude")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	result := updated.(LogoModel)
	if result.width != 120 {
		t.Errorf("expected width 120, got %d", result.width)
	}
	if result.height != 50 {
		t.Errorf("expected height 50, got %d", result.height)
	}
}

func TestLogoModel_View_uncentered_when_dimensions_unknown(t *testing.T) {
	m := NewLogo("")
	got := m.View()
	want := RenderGhost(GhostForTool("", false))
	if got != want {
		t.Errorf("expected uncentered ghost output when dimensions unknown\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLogoModel_View_centers_when_dimensions_known(t *testing.T) {
	m := NewLogo("")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	m = updated.(LogoModel)

	got := m.View()
	raw := RenderGhost(GhostForTool("", false))

	// lipgloss.Place pads to fill the canvas — the output should be larger than raw
	if len(got) <= len(raw) {
		t.Errorf("expected centered output to be larger than raw ghost (padded by lipgloss.Place)\ngot len=%d, raw len=%d", len(got), len(raw))
	}

	// The output should start with spaces (top/left padding from lipgloss.Place)
	if !strings.HasPrefix(got, " ") && !strings.HasPrefix(got, "\n") {
		t.Errorf("expected centered output to start with padding spaces or newlines, got: %q", got[:min(20, len(got))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
