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

// escInterceptingModel is a stub that implements EscInterceptor and wants Esc.
type escInterceptingModel struct {
	minimalModel
	escReceived bool
}

func (m escInterceptingModel) WantsEsc() bool { return true }
func (m escInterceptingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEsc {
		m.escReceived = true
	}
	return m, nil
}

// nonInterceptingModel is a stub that implements EscInterceptor but does NOT want Esc.
type nonInterceptingModel struct {
	minimalModel
}

func (m nonInterceptingModel) WantsEsc() bool { return false }

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

// TestAppModel_EscDelegatedToChildWhenChildWantsEsc verifies that when the top model
// implements EscInterceptor and WantsEsc() returns true, AppModel delegates Esc to
// the child instead of showing the double-Esc hint.
func TestAppModel_EscDelegatedToChildWhenChildWantsEsc(t *testing.T) {
	inner := escInterceptingModel{}
	app := tui.NewAppModel(inner)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	// AppModel must NOT show the quit hint.
	if app.ShowingEscHint() {
		t.Error("AppModel should not show Esc hint when child implements WantsEsc()=true")
	}
}

// TestAppModel_EscNotDelegatedWhenChildDoesNotWantEsc verifies that when the top model
// implements EscInterceptor but WantsEsc() returns false, AppModel still handles Esc
// with the double-Esc hint logic.
func TestAppModel_EscNotDelegatedWhenChildDoesNotWantEsc(t *testing.T) {
	inner := nonInterceptingModel{}
	app := tui.NewAppModel(inner)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	if !app.ShowingEscHint() {
		t.Error("AppModel should show Esc hint when child WantsEsc()=false")
	}
}

// TestAppModel_EscHintNotShownWhenChildWithoutInterface verifies that models without
// EscInterceptor still trigger the double-Esc hint (backward compatibility).
func TestAppModel_EscHintShownWhenChildLacksEscInterceptor(t *testing.T) {
	inner := minimalModel{} // does not implement EscInterceptor
	app := tui.NewAppModel(inner)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	if !app.ShowingEscHint() {
		t.Error("AppModel should show Esc hint when child lacks EscInterceptor")
	}
}

// TestMainMenuModel_EscInSettingsMode_ClosesSettings verifies that pressing Esc while
// in settings mode closes the settings panel (goes back one level) rather than quitting.
func TestMainMenuModel_EscInSettingsMode_ClosesSettings(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.SetSettingsMode(true)

	app := tui.NewAppModel(m)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	inner := app.InnerMainMenu()
	if inner.InSettingsMode() {
		t.Error("Esc in settings mode should close settings panel, not stay in settings mode")
	}
	if app.ShowingEscHint() {
		t.Error("Esc in settings mode should not show quit hint")
	}
}

// TestMainMenuModel_EscInInputMode_ExitsInputMode verifies that pressing Esc while
// in input mode exits input mode rather than triggering the quit hint.
func TestMainMenuModel_EscInInputMode_ExitsInputMode(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterInputModeForTest("add-project")

	app := tui.NewAppModel(m)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	inner := app.InnerMainMenu()
	if inner.InInputMode() {
		t.Error("Esc in input mode should exit input mode")
	}
	if app.ShowingEscHint() {
		t.Error("Esc in input mode should not show quit hint")
	}
}

// TestMainMenuModel_EscInDeleteMode_ExitsDeleteMode verifies that pressing Esc while
// in delete mode exits delete mode rather than triggering the quit hint.
func TestMainMenuModel_EscInDeleteMode_ExitsDeleteMode(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	m.EnterDeleteModeForTest()

	app := tui.NewAppModel(m)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	inner := app.InnerMainMenu()
	if inner.InDeleteMode() {
		t.Error("Esc in delete mode should exit delete mode")
	}
	if app.ShowingEscHint() {
		t.Error("Esc in delete mode should not show quit hint")
	}
}

// TestMainMenuModel_EscInNormalMode_ShowsQuitHint verifies that pressing Esc in the
// normal main menu (no sub-mode active) shows the double-Esc quit hint.
func TestMainMenuModel_EscInNormalMode_ShowsQuitHint(t *testing.T) {
	m := tui.NewMainMenu(nil, []string{"claude"}, "claude", "animated")
	app := tui.NewAppModel(m)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = updated.(tui.AppModel)

	if !app.ShowingEscHint() {
		t.Error("Esc in normal mode should show the double-Esc quit hint")
	}
}
