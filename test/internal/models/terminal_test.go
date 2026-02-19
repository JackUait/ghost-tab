package models_test

import (
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func TestSupportedTerminals_returns_four_terminals(t *testing.T) {
	terminals := models.SupportedTerminals()
	if len(terminals) != 4 {
		t.Errorf("expected 4 terminals, got %d", len(terminals))
	}
}

func TestSupportedTerminals_contains_ghostty(t *testing.T) {
	terminals := models.SupportedTerminals()
	found := false
	for _, term := range terminals {
		if term.Name == "ghostty" {
			found = true
			if term.DisplayName != "Ghostty" {
				t.Errorf("expected DisplayName 'Ghostty', got %q", term.DisplayName)
			}
			if term.CaskName != "ghostty" {
				t.Errorf("expected CaskName 'ghostty', got %q", term.CaskName)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find ghostty in supported terminals")
	}
}

func TestTerminal_String_shows_installed_status(t *testing.T) {
	term := models.Terminal{
		Name:        "ghostty",
		DisplayName: "Ghostty",
		Installed:   true,
	}
	got := term.String()
	if got != "Ghostty ✓" {
		t.Errorf("expected 'Ghostty ✓', got %q", got)
	}
}

func TestTerminal_String_shows_not_installed(t *testing.T) {
	term := models.Terminal{
		Name:        "ghostty",
		DisplayName: "Ghostty",
		Installed:   false,
	}
	got := term.String()
	if got != "Ghostty (not installed)" {
		t.Errorf("expected 'Ghostty (not installed)', got %q", got)
	}
}

func TestDetectTerminals_marks_installation_status(t *testing.T) {
	terminals := models.DetectTerminals()
	if len(terminals) == 0 {
		t.Error("expected at least one terminal")
	}
}
