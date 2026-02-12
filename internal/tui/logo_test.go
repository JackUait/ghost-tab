package tui

import (
	"testing"
	"time"
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
