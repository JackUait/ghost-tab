package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffViewModel is a scrollable pager for a (pre-colored) git diff, shown inside
// the click-to-open popup. Unlike less, it closes on a single Esc press because
// Bubbletea's input parser emits a distinct KeyEscape for a lone Esc and parses
// arrow-key escape sequences separately. q and ctrl+c also quit. The viewport
// bubble handles scrolling (↑↓/j/k, space/b page, u/d half-page, mouse wheel);
// g/G jump to the top/bottom. ANSI color in the content is preserved.
type DiffViewModel struct {
	title    string
	content  string
	added    int
	deleted  int
	viewport viewport.Model
	ready    bool
	quitting bool
}

var (
	diffTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("208")). // orange, matching the popup border
			Padding(0, 1)

	diffAddStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	diffDelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red

	diffRuleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

	diffBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 1)
)

var diffAnsiSeq = regexp.MustCompile("\x1b\\[[0-9;]*m")

// countDiffLines tallies the added (+) and deleted (-) lines of the diff body.
// The body is pre-colored (git --color=always) and the +++/--- file markers are
// stripped upstream, so after dropping the ANSI escapes a leading +/- is an
// authoritative add/delete marker; context lines (leading space) are ignored.
func countDiffLines(content string) (added, deleted int) {
	for _, line := range strings.Split(content, "\n") {
		s := diffAnsiSeq.ReplaceAllString(line, "")
		if s == "" {
			continue
		}
		switch s[0] {
		case '+':
			added++
		case '-':
			deleted++
		}
	}
	return added, deleted
}

// NewDiffView builds the pager for the given title (the file path, shown in the
// header) and content (the colored diff body). The added/deleted line counts
// shown in the header are derived from the content.
func NewDiffView(title, content string) DiffViewModel {
	added, deleted := countDiffLines(content)
	return DiffViewModel{title: title, content: content, added: added, deleted: deleted}
}

func (m DiffViewModel) Init() tea.Cmd {
	return nil
}

// headerHeight and footerHeight are the chrome rows reserved above and below the
// scrolling viewport: a title line + a rule, and a single control bar.
const (
	diffHeaderHeight = 2
	diffFooterHeight = 1
)

func (m DiffViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h := msg.Height - diffHeaderHeight - diffFooterHeight
		if h < 1 {
			h = 1
		}
		if !m.ready {
			m.viewport = viewport.New(msg.Width, h)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = h
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'q', 'Q':
					m.quitting = true
					return m, tea.Quit
				case 'g':
					m.viewport.GotoTop()
					return m, nil
				case 'G':
					m.viewport.GotoBottom()
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DiffViewModel) View() string {
	if m.quitting {
		return ""
	}
	if !m.ready {
		return ""
	}

	width := m.viewport.Width
	// Top line: ONLY the file path and the added/deleted line counts.
	title := diffTitleStyle.Render(m.title) +
		diffAddStyle.Render("+"+itoa(m.added)) + " " +
		diffDelStyle.Render("−"+itoa(m.deleted))
	rule := diffRuleStyle.Render(strings.Repeat("─", maxInt(width, 0)))

	pct := int(m.viewport.ScrollPercent() * 100)
	hints := "↑↓/jk scroll · space/b page · g/G top·end · q/Esc close"
	bar := diffBarStyle.Render(hints + "    " + padPercent(pct))

	return strings.Join([]string{title, rule, m.viewport.View(), bar}, "\n")
}

func padPercent(p int) string {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	s := "  "
	switch {
	case p >= 100:
		s = ""
	case p >= 10:
		s = " "
	}
	return s + itoa(p) + "%"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// itoa avoids pulling in strconv for a single small non-negative int.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
