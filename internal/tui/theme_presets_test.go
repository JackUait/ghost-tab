package tui

import "testing"

func TestThemePresets_listStartsWithAuto(t *testing.T) {
	if len(ThemePresets) < 2 {
		t.Fatalf("expected several presets, got %d", len(ThemePresets))
	}
	if ThemePresets[0] != "auto" {
		t.Errorf("first preset should be 'auto' (the default), got %q", ThemePresets[0])
	}
	// The six named presets must all be present and resolvable.
	for _, want := range []string{"orange", "purple", "green", "blue", "rose", "cyan"} {
		found := false
		for _, p := range ThemePresets {
			if p == want {
				found = true
			}
		}
		if !found {
			t.Errorf("preset %q missing from ThemePresets", want)
		}
	}
}

func TestResolveTheme_autoFollowsTool(t *testing.T) {
	if got := ResolveTheme("opencode", "auto"); got.Name != "opencode" {
		t.Errorf("auto on opencode should give opencode theme, got %q", got.Name)
	}
	if got := ResolveTheme("claude", "auto"); got.Name != "claude" {
		t.Errorf("auto on claude should give claude theme, got %q", got.Name)
	}
	// Empty / unknown pref also falls back to the tool.
	if got := ResolveTheme("opencode", ""); got.Name != "opencode" {
		t.Errorf("empty pref should follow tool, got %q", got.Name)
	}
	if got := ResolveTheme("opencode", "bogus"); got.Name != "opencode" {
		t.Errorf("unknown pref should follow tool, got %q", got.Name)
	}
}

func TestResolveTheme_namedPresetOverridesTool(t *testing.T) {
	// Orange and purple reuse the existing tool palettes...
	if got := ResolveTheme("opencode", "orange"); got.Primary != themes["claude"].Primary {
		t.Errorf("orange preset should use claude's orange Primary, got %v", got.Primary)
	}
	if got := ResolveTheme("claude", "purple"); got.Primary != themes["opencode"].Primary {
		t.Errorf("purple preset should use opencode's purple Primary, got %v", got.Primary)
	}
	// ...the new presets resolve regardless of the active tool.
	for _, name := range []string{"green", "blue", "rose", "cyan"} {
		got := ResolveTheme("claude", name)
		if got.Name != name {
			t.Errorf("preset %q should resolve to its own palette, got Name=%q", name, got.Name)
		}
		if string(got.Primary) == "" {
			t.Errorf("preset %q has no Primary color", name)
		}
	}
}
