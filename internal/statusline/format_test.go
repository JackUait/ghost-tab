package statusline

import (
	"fmt"
	"strings"
	"testing"
)

// --- FormatMemory tests (ported from statusline.bats) ---

func TestFormatMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts KB to MB",
			input:    "512000",
			expected: "500M",
		},
		{
			name:     "small MB value",
			input:    "102400",
			expected: "100M",
		},
		{
			name:     "converts to GB with decimal",
			input:    "1572864",
			expected: "1.5G",
		},
		{
			name:     "exactly 1 GB",
			input:    "1048576",
			expected: "1.0G",
		},
		{
			name:     "zero returns 0M",
			input:    "0",
			expected: "0M",
		},
		{
			name:     "handles negative values",
			input:    "-1024",
			expected: "0M",
		},
		{
			name:     "handles very large values (10TB in KB)",
			input:    "10737418240",
			expected: "10240.0G",
		},
		{
			name:     "handles non-numeric input gracefully",
			input:    "not_a_number",
			expected: "0M",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "0M",
		},
		{
			name:     "handles floating point input",
			input:    "512000.5",
			expected: "0M",
		},
		{
			name:     "handles exactly 1024 MB boundary",
			input:    "1048576",
			expected: "1.0G",
		},
		{
			name:     "handles just below GB boundary (1023 MB)",
			input:    "1047552",
			expected: "1023M",
		},
		{
			name:     "handles just above GB boundary (1025 MB)",
			input:    "1049600",
			expected: "1.0G",
		},
		{
			name:     "handles KB value with remainder",
			input:    "512500",
			expected: "500M",
		},
		{
			name:     "handles very small non-zero values",
			input:    "1",
			expected: "0M",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMemory(tt.input)
			if result != tt.expected {
				t.Errorf("FormatMemory(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- ParseCWDFromJSON tests (ported from statusline.bats) ---

func TestParseCWDFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "extracts current_dir from JSON",
			input:    `{"current_dir":"/Users/me/project"}`,
			expected: "/Users/me/project",
		},
		{
			name:     "handles nested JSON",
			input:    `{"foo":"bar","current_dir":"/tmp/test","baz":1}`,
			expected: "/tmp/test",
		},
		{
			name:     "returns empty for missing key",
			input:    `{"foo":"bar"}`,
			expected: "",
		},
		{
			name:     "handles malformed JSON - missing quotes",
			input:    `{current_dir:/tmp/test}`,
			expected: "",
		},
		{
			name:     "handles malformed JSON - missing braces",
			input:    `"current_dir":"/tmp/test"`,
			expected: "/tmp/test",
		},
		{
			name:     "handles malformed JSON - trailing comma",
			input:    `{"current_dir":"/tmp/test",}`,
			expected: "/tmp/test",
		},
		{
			name:     "handles empty JSON object",
			input:    `{}`,
			expected: "",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles whitespace-only string",
			input:    "   ",
			expected: "",
		},
		{
			name:     "handles binary data",
			input:    "\x00\x01\x02\x03",
			expected: "",
		},
		{
			name:     "handles path with escaped characters",
			input:    `{"current_dir":"/tmp/test\\ndir"}`,
			expected: `/tmp/test\\ndir`,
		},
		{
			name:     "handles path with special JSON chars (stops at first unescaped quote)",
			input:    `{"current_dir":"/tmp/test\"quoted"}`,
			expected: `/tmp/test\`,
		},
		{
			name:     "handles very long path",
			input:    `{"current_dir":"` + buildLongPath(50) + `"}`,
			expected: buildLongPath(50),
		},
		{
			name:     "handles path with Unicode characters",
			input:    "{\"current_dir\":\"/tmp/t\u00ebst/\u65e5\u672c\u8a9e/\u00e9moji\U0001F389\"}",
			expected: "/tmp/t\u00ebst/\u65e5\u672c\u8a9e/\u00e9moji\U0001F389",
		},
		{
			name:     "handles multiple current_dir keys (takes last)",
			input:    `{"current_dir":"/first","other":"stuff","current_dir":"/second"}`,
			expected: "/second",
		},
		{
			name:     "handles Windows-style line endings",
			input:    "{\r\n\"current_dir\":\"/tmp/test\"\r\n}\r\n",
			expected: "/tmp/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCWDFromJSON(tt.input)
			if result != tt.expected {
				t.Errorf("ParseCWDFromJSON(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// buildLongPath creates a path like /very/long/path/with/many/segments/segment1/segment2/...
// Matches the bash test: long_path="/very/long/path/with/many/segments"; for i in {1..50}; do long_path+="/segment${i}"; done
func buildLongPath(segments int) string {
	var sb strings.Builder
	sb.WriteString("/very/long/path/with/many/segments")
	for i := 1; i <= segments; i++ {
		sb.WriteString(fmt.Sprintf("/segment%d", i))
	}
	return sb.String()
}
