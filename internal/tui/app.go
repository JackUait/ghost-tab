package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PushScreenMsg tells AppModel to push a new screen onto the navigation stack.
type PushScreenMsg struct{ Model tea.Model }

// PopScreenMsg tells AppModel to pop the current screen and reveal the previous one.
type PopScreenMsg struct{}

// escHintExpiredMsg is fired when the double-Esc hint timer expires.
type escHintExpiredMsg struct{}

const escHintTimeout = 2 * time.Second

// AppModel is the root Bubbletea model that owns the navigation stack.
// All screens are pushed/popped via PushScreenMsg and PopScreenMsg.
type AppModel struct {
	stack        []tea.Model
	escPressedAt time.Time
	showEscHint  bool
}

// NewAppModel creates an AppModel with the given model as the initial (bottom) screen.
func NewAppModel(initial tea.Model) AppModel {
	return AppModel{
		stack: []tea.Model{initial},
	}
}

// Depth returns the number of screens on the stack.
func (a AppModel) Depth() int {
	return len(a.stack)
}

// ShowingEscHint reports whether the "Press Esc again to quit" hint is visible.
func (a AppModel) ShowingEscHint() bool {
	return a.showEscHint
}

// ForceEscPressedAt sets escPressedAt for testing purposes.
func (a *AppModel) ForceEscPressedAt(t time.Time) {
	a.escPressedAt = t
}

// top returns the active (topmost) model.
func (a AppModel) top() tea.Model {
	return a.stack[len(a.stack)-1]
}

// replaceTop returns a copy of the AppModel with the top model replaced.
func (a AppModel) replaceTop(m tea.Model) AppModel {
	newStack := make([]tea.Model, len(a.stack))
	copy(newStack, a.stack)
	newStack[len(newStack)-1] = m
	a.stack = newStack
	return a
}

// InnerMainMenu returns the MainMenuModel at the bottom of the stack.
// Panics if the bottom model is not a *MainMenuModel.
func (a AppModel) InnerMainMenu() *MainMenuModel {
	return a.stack[0].(*MainMenuModel)
}

func (a AppModel) Init() tea.Cmd {
	return a.top().Init()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ctrl-C always quits, regardless of depth.
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyCtrlC {
		return a, tea.Quit
	}

	// Handle messages emitted by child models.
	switch msg := msg.(type) {
	case PushScreenMsg:
		newStack := make([]tea.Model, len(a.stack)+1)
		copy(newStack, a.stack)
		newStack[len(newStack)-1] = msg.Model
		a.stack = newStack
		// Reset esc hint state when navigating forward.
		a.showEscHint = false
		a.escPressedAt = time.Time{}
		return a, msg.Model.Init()

	case PopScreenMsg:
		if len(a.stack) <= 1 {
			return a, tea.Quit
		}
		// Check if popped model has a result to relay.
		popped := a.top()
		newStack := make([]tea.Model, len(a.stack)-1)
		copy(newStack, a.stack[:len(a.stack)-1])
		a.stack = newStack
		// Relay result to new top if the popped model implements ResultProvider.
		if rp, ok := popped.(ResultProvider); ok {
			if relayMsg := rp.PopResult(); relayMsg != nil {
				updated, cmd := a.top().Update(relayMsg)
				a = a.replaceTop(updated)
				return a, cmd
			}
		}
		return a, nil

	case escHintExpiredMsg:
		// Only clear hint if no second Esc arrived.
		a.showEscHint = false
		a.escPressedAt = time.Time{}
		return a, nil
	}

	// Handle Esc at depth 1 (double-Esc to quit).
	// Only intercept if the top model does not want to handle Esc itself
	// (e.g. to close a sub-panel or exit an input mode).
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEsc && len(a.stack) == 1 {
		// If the child claims Esc, delegate to it.
		if ei, ok := a.top().(EscInterceptor); ok && ei.WantsEsc() {
			// Fall through to the delegate-to-top-model path below.
		} else {
			if a.showEscHint && !a.escPressedAt.IsZero() &&
				time.Since(a.escPressedAt) < escHintTimeout {
				// Second Esc within window — quit.
				return a, tea.Quit
			}
			// First Esc — show hint, start timer.
			a.showEscHint = true
			a.escPressedAt = time.Now()
			return a, tea.Tick(escHintTimeout, func(time.Time) tea.Msg {
				return escHintExpiredMsg{}
			})
		}
	}

	// Any other keypress clears the hint.
	if _, ok := msg.(tea.KeyMsg); ok {
		a.showEscHint = false
		a.escPressedAt = time.Time{}
	}

	// Delegate to top model.
	updated, cmd := a.top().Update(msg)
	a = a.replaceTop(updated)

	// If the top model returned a PopScreenMsg or PushScreenMsg as a command
	// result, handle it in the next Update cycle (Bubbletea dispatches it naturally).
	return a, cmd
}

func (a AppModel) View() string {
	view := a.top().View()
	if a.showEscHint {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		hint := hintStyle.Render("  Press Esc again to quit")
		if view != "" {
			view = view + "\n" + hint
		} else {
			view = hint
		}
	}
	return view
}

// ResultProvider is implemented by models that want to pass a result back to
// the model below them on the stack when they are popped.
type ResultProvider interface {
	PopResult() tea.Msg
}

// EscInterceptor is implemented by models that want to handle Esc internally
// (e.g. to close a sub-panel or exit a sub-mode) rather than having AppModel
// apply the double-Esc quit logic. AppModel checks WantsEsc() before deciding
// whether to intercept or delegate.
type EscInterceptor interface {
	// WantsEsc returns true when the model is in a state where Esc should be
	// consumed by the model itself (e.g. settings panel open, input mode active).
	// Return false when the model is in its normal top-level state so that
	// AppModel can apply the double-Esc-to-quit protection.
	WantsEsc() bool
}
