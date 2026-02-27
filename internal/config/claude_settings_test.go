package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- MergeStatusLine tests (ported from settings-json.bats merge_claude_settings) ---

func TestMergeStatusLine(t *testing.T) {
	defaultStatusLine := map[string]interface{}{
		"type":    "command",
		"command": "bash ~/.claude/statusline-wrapper.sh",
	}

	t.Run("creates new file with statusLine", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineCreated {
			t.Errorf("expected StatusLineCreated, got %v", result)
		}

		// Verify file exists and contains statusLine
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if !strings.Contains(string(data), "statusLine") {
			t.Error("file should contain 'statusLine'")
		}
		if !strings.Contains(string(data), "statusline-wrapper.sh") {
			t.Error("file should contain 'statusline-wrapper.sh'")
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file is not valid JSON: %v", err)
		}
	})

	t.Run("skips when statusLine already exists", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		existing := `{"statusLine": {"type": "command"}}`
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineExists {
			t.Errorf("expected StatusLineExists, got %v", result)
		}

		// Verify file was not modified
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file should remain valid JSON: %v", err)
		}
	})

	t.Run("handles Windows line endings in existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		existing := "{\"foo\":\"bar\"}\r\n"
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineAdded {
			t.Errorf("expected StatusLineAdded, got %v", result)
		}

		// Verify the result is valid JSON with statusLine
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "statusLine") {
			t.Error("file should contain 'statusLine'")
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file should be valid JSON: %v", err)
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Empty file should be treated as new — add statusLine
		if result != StatusLineAdded {
			t.Errorf("expected StatusLineAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "statusLine") {
			t.Error("file should contain 'statusLine'")
		}
	})

	t.Run("handles read-only file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0444); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(path, 0644)

		_, err := MergeStatusLine(path, defaultStatusLine)
		if err == nil {
			t.Error("expected error for read-only file, got nil")
		}
	})

	t.Run("handles file with 1000+ lines", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		// Build a large JSON object
		var sb strings.Builder
		sb.WriteString("{\n")
		for i := 1; i <= 1000; i++ {
			if i > 1 {
				sb.WriteString(",\n")
			}
			sb.WriteString(fmt.Sprintf("  \"key%d\": \"value%d\"", i, i))
		}
		sb.WriteString("\n}")

		if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineAdded {
			t.Errorf("expected StatusLineAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "statusLine") {
			t.Error("file should contain 'statusLine'")
		}
		if !strings.Contains(string(data), "key1000") {
			t.Error("file should preserve existing key1000")
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file should be valid JSON: %v", err)
		}
	})

	t.Run("creates parent directories if needed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "nested", "settings.json")

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineCreated {
			t.Errorf("expected StatusLineCreated, got %v", result)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("file should have been created")
		}
	})

	t.Run("preserves existing keys when adding statusLine", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		existing := `{"permissions": {"allow": ["Bash"]}, "apiKey": "test123"}`
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := MergeStatusLine(path, defaultStatusLine)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != StatusLineAdded {
			t.Errorf("expected StatusLineAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("file should be valid JSON: %v", err)
		}

		if _, ok := parsed["statusLine"]; !ok {
			t.Error("should have statusLine key")
		}
		if _, ok := parsed["permissions"]; !ok {
			t.Error("should preserve permissions key")
		}
		if _, ok := parsed["apiKey"]; !ok {
			t.Error("should preserve apiKey key")
		}
	})
}

