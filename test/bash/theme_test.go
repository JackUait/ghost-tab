package bash_test

import "testing"

func TestTheme_resolve_namedPresetWins(t *testing.T) {
	for _, tc := range []struct{ pref, tool, want string }{
		{"purple", "claude", "purple"},
		{"green", "claude", "green"},
		{"orange", "opencode", "orange"},
		{"rose", "opencode", "rose"},
	} {
		out, code := runBashFunc(t, "lib/theme.sh", "gt_resolve_theme", []string{tc.pref, tc.tool}, nil)
		assertExitCode(t, code, 0)
		assertContains(t, out, tc.want)
	}
}

func TestTheme_resolve_autoFollowsTool(t *testing.T) {
	cases := []struct{ pref, tool, want string }{
		{"auto", "opencode", "purple"},
		{"auto", "claude", "orange"},
		{"", "opencode", "purple"},
		{"bogus", "claude", "orange"},
	}
	for _, tc := range cases {
		out, code := runBashFunc(t, "lib/theme.sh", "gt_resolve_theme", []string{tc.pref, tc.tool}, nil)
		assertExitCode(t, code, 0)
		assertContains(t, out, tc.want)
	}
}

func TestTheme_accent_perPreset(t *testing.T) {
	cases := []struct{ key, want string }{
		{"orange", "209"}, {"purple", "141"}, {"green", "78"},
		{"blue", "75"}, {"rose", "211"}, {"cyan", "80"},
	}
	for _, tc := range cases {
		out, code := runBashFunc(t, "lib/theme.sh", "get_theme_accent", []string{tc.key}, nil)
		assertExitCode(t, code, 0)
		assertContains(t, out, tc.want)
	}
}

func TestTheme_palette_perPreset(t *testing.T) {
	cases := []struct{ key, want string }{
		{"orange", "130 166 172 208 209 214 215 220"},
		{"purple", "60 61 62 99 135 141 147 183"},
		{"green", "22 28 34 35 41 77 78 120"},
		{"blue", "17 18 25 26 31 32 75 117"},
		{"rose", "52 89 125 161 168 205 211 218"},
		{"cyan", "23 30 37 43 44 80 116 123"},
	}
	for _, tc := range cases {
		out, code := runBashFunc(t, "lib/theme.sh", "get_theme_palette", []string{tc.key}, nil)
		assertExitCode(t, code, 0)
		assertContains(t, out, tc.want)
	}
}
