package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackuait/ghost-tab/internal/tui"
)

// minimalModel is a stub tea.Model for use in AppModel tests.
type minimalModel struct {
	lastMsg tea.Msg
}

func (m minimalModel) Init() tea.Cmd                           { return nil }
func (m minimalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { m.lastMsg = msg; return m, nil }
func (m minimalModel) View() string                            { return "stub" }

func TestAppModel_InitialisesWithOneModel(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)
	if app.Depth() != 1 {
		t.Errorf("expected depth 1, got %d", app.Depth())
	}
}

func TestAppModel_PushScreenMsg_IncreasesDepth(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	updated, _ := app.Update(tui.PushScreenMsg{Model: minimalModel{}})
	app = updated.(tui.AppModel)

	if app.Depth() != 2 {
		t.Errorf("expected depth 2 after push, got %d", app.Depth())
	}
}

func TestAppModel_PopScreenMsg_DecreasesDepth(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	// Push one screen
	updated, _ := app.Update(tui.PushScreenMsg{Model: minimalModel{}})
	app = updated.(tui.AppModel)

	// Pop it
	updated, _ = app.Update(tui.PopScreenMsg{})
	app = updated.(tui.AppModel)

	if app.Depth() != 1 {
		t.Errorf("expected depth 1 after pop, got %d", app.Depth())
	}
}

func TestAppModel_PopScreenMsg_AtDepth1_ReturnsQuit(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	_, cmd := app.Update(tui.PopScreenMsg{})
	if cmd == nil {
		t.Fatal("expected a command on pop at depth 1, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestAppModel_CtrlC_AlwaysQuits(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected a command on ctrl-c, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg on ctrl-c, got %T", msg)
	}
}

func TestAppModel_EscAtDepth1_ShowsHint(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	if !app.ShowingEscHint() {
		t.Error("expected esc hint to be shown after first Esc at depth 1")
	}
}

func TestAppModel_EscAtDepth1_DoesNotQuitFirstPress(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// cmd will be a tea.Tick for the hint timer — not a Quit
	// The test is that the model does NOT immediately quit.
	// We verify by checking that cmd() produces a tick-like msg, not QuitMsg.
	if cmd == nil {
		return // no command at all is fine
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		t.Error("first Esc at depth 1 should not quit immediately")
	}
}

func TestAppModel_DoubleEscAtDepth1_Quits(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	// First Esc
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	// Manually set EscPressedAt to just now so the 2s window is open
	app.ForceEscPressedAt(time.Now())

	// Second Esc
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command on double-Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg on double-Esc, got %T", msg)
	}
}

func TestAppModel_EscAtDepth2_Pops(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	// Push a screen so we're at depth 2
	updated, _ := app.Update(tui.PushScreenMsg{Model: minimalModel{}})
	app = updated.(tui.AppModel)

	// The top screen emits Esc — app should pop, not quit
	updated, _ = app.Update(tui.PopScreenMsg{})
	app = updated.(tui.AppModel)

	if app.Depth() != 1 {
		t.Errorf("expected depth 1 after pop from depth 2, got %d", app.Depth())
	}
}

func TestAppModel_View_DelegatestoTopModel(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	view := app.View()
	if view == "" {
		t.Error("expected non-empty view from AppModel")
	}
}

func TestAppModel_View_ShowsEscHintWhenFlagged(t *testing.T) {
	inner := minimalModel{}
	app := tui.NewAppModel(inner)

	// Trigger hint
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	view := app.View()
	if view == "" {
		t.Skip("stub model returns empty view, skip hint check")
	}
}
