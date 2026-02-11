package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackuait/ghost-tab/internal/util"
)

// SuggestionProvider is a function that returns suggestions for a given input.
type SuggestionProvider func(input string) []string

// PathSuggestionProvider returns a SuggestionProvider that suggests directory paths.
// Empty input defaults to ~/. Results are sorted alphabetically and capped at maxResults.
// Matching is case-insensitive and supports substring (glob-style) matching.
func PathSuggestionProvider(maxResults int) SuggestionProvider {
	return func(input string) []string {
		if input == "" {
			input = "~/"
		}

		expanded := util.ExpandPath(input)

		var dir string
		var prefix string

		if strings.HasSuffix(input, "/") {
			dir = expanded
			prefix = ""
		} else {
			dir = filepath.Dir(expanded)
			prefix = filepath.Base(expanded)
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		lowerPrefix := strings.ToLower(prefix)
		var suggestions []string

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasPrefix(name, ".") {
				continue
			}
			lowerName := strings.ToLower(name)
			// Glob-style: match if prefix appears anywhere in name
			if prefix == "" || strings.Contains(lowerName, lowerPrefix) {
				var suggestion string
				if strings.HasSuffix(input, "/") {
					suggestion = input + name + "/"
				} else {
					parentInput := input[:len(input)-len(filepath.Base(input))]
					suggestion = parentInput + name + "/"
				}
				suggestions = append(suggestions, suggestion)
			}
		}

		sort.Strings(suggestions)

		if len(suggestions) > maxResults {
			suggestions = suggestions[:maxResults]
		}

		return suggestions
	}
}

// AutocompleteModel is a reusable autocomplete component.
// It manages suggestions, navigation, and selection state.
// Embed it in another model and call its methods from Update().
type AutocompleteModel struct {
	provider        SuggestionProvider
	suggestions     []string
	selected        int
	showSuggestions bool
	maxResults      int
	lastInput       string
}

// NewAutocomplete creates a new autocomplete model with the given provider.
func NewAutocomplete(provider SuggestionProvider, maxResults int) AutocompleteModel {
	if maxResults <= 0 {
		maxResults = 8
	}
	return AutocompleteModel{
		provider:   provider,
		maxResults: maxResults,
	}
}

// SetInput updates the current input. Call RefreshSuggestions() after to update suggestions.
func (m *AutocompleteModel) SetInput(input string) {
	m.lastInput = input
}

// RefreshSuggestions calls the provider with the current input.
func (m *AutocompleteModel) RefreshSuggestions() {
	if m.provider == nil {
		return
	}
	m.suggestions = m.provider(m.lastInput)
	m.selected = 0
	m.showSuggestions = len(m.suggestions) > 0
}

// Suggestions returns the current suggestion list.
func (m *AutocompleteModel) Suggestions() []string {
	return m.suggestions
}

// Selected returns the index of the currently highlighted suggestion.
func (m *AutocompleteModel) Selected() int {
	return m.selected
}

// ShowSuggestions returns whether the suggestion dropdown is visible.
func (m *AutocompleteModel) ShowSuggestions() bool {
	return m.showSuggestions
}

// MoveDown moves the selection down, wrapping to top.
func (m *AutocompleteModel) MoveDown() {
	if len(m.suggestions) == 0 {
		return
	}
	m.selected = (m.selected + 1) % len(m.suggestions)
}

// MoveUp moves the selection up, wrapping to bottom.
func (m *AutocompleteModel) MoveUp() {
	if len(m.suggestions) == 0 {
		return
	}
	m.selected = (m.selected - 1 + len(m.suggestions)) % len(m.suggestions)
}

// AcceptSelected returns the currently selected suggestion.
func (m *AutocompleteModel) AcceptSelected() string {
	if len(m.suggestions) == 0 || m.selected >= len(m.suggestions) {
		return ""
	}
	return m.suggestions[m.selected]
}

// Dismiss hides the suggestion dropdown.
func (m *AutocompleteModel) Dismiss() {
	m.showSuggestions = false
	m.suggestions = nil
}
