package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// --- AddSoundHook tests (ported from settings-json.bats add_sound_notification_hook) ---

func TestAddSoundHook(t *testing.T) {
	defaultCmd := "afplay /System/Library/Sounds/Bottle.aiff &"

	t.Run("adds hook to empty settings", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
		if !strings.Contains(content, "Bottle.aiff") {
			t.Error("file should contain 'Bottle.aiff'")
		}

		// Verify valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file should be valid JSON: %v", err)
		}
	})

	t.Run("skips when already exists", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		existing := `{
  "hooks": {
    "Notification": [
      {
        "matcher": "idle_prompt",
        "hooks": [{"type": "command", "command": "afplay /System/Library/Sounds/Bottle.aiff &"}]
      }
    ]
  }
}`
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookExists {
			t.Errorf("expected HookExists, got %v", result)
		}
	})

	t.Run("creates file when missing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "new-settings.json")

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("file should have been created")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	// --- Malformed JSON tests ---

	t.Run("handles malformed JSON - missing closing brace", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(`{"foo": "bar"`), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		// Verify resulting file is valid JSON
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("resulting file should be valid JSON: %v", err)
		}
	})

	t.Run("handles malformed JSON - missing quotes", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(`{foo: bar}`), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles malformed JSON - trailing comma", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(`{"foo": "bar",}`), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles malformed JSON - missing commas", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(`{"foo": "bar" "baz": "qux"}`), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles completely corrupted JSON", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("not even json at all!!!"), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles binary file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles Windows line endings", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{\r\n  \"foo\": \"bar\"\r\n}\r\n"), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	// --- Empty and whitespace files ---

	t.Run("handles empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	t.Run("handles file with only whitespace", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("   \n\n  \t\t  \n"), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
	})

	// --- Special characters ---

	t.Run("handles command with special shell characters", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		specialCmd := `echo "test $VAR & && || ; | > < ( )"`
		result, err := AddSoundHook(path, specialCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, `"command":`) {
			t.Error("file should contain '\"command\":'")
		}
		if !strings.Contains(content, "echo") {
			t.Error("file should contain 'echo'")
		}

		// Verify valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("file should be valid JSON: %v", err)
		}
	})

	t.Run("handles command with quotes and newlines", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := `echo 'single' "double"`
		result, err := AddSoundHook(path, cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// JSON will have the quotes escaped
		if !strings.Contains(string(data), "echo 'single'") {
			t.Error("file should contain the single-quoted part of the command")
		}
	})

	// --- Permission denied ---

	t.Run("handles read-only file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0444); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(path, 0644)

		_, err := AddSoundHook(path, defaultCmd)
		if err == nil {
			t.Error("expected error for read-only file, got nil")
		}
	})

	// --- Large files ---

	t.Run("handles large JSON file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		// Build a JSON file with 1000+ entries
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

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, "idle_prompt") {
			t.Error("file should contain 'idle_prompt'")
		}
		if !strings.Contains(content, "key1000") {
			t.Error("file should preserve existing key1000")
		}
	})

	// --- Concurrent writes ---

	t.Run("handles concurrent writes to same file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		errs := make([]error, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			_, errs[0] = AddSoundHook(path, defaultCmd)
		}()
		go func() {
			defer wg.Done()
			_, errs[1] = AddSoundHook(path, defaultCmd)
		}()
		wg.Wait()

		// At least one should succeed
		anySuccess := errs[0] == nil || errs[1] == nil
		if !anySuccess {
			t.Errorf("at least one concurrent write should succeed, got errors: %v, %v", errs[0], errs[1])
		}

		// File should be valid JSON
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Errorf("resulting file should be valid JSON: %v", err)
		}
	})

	// --- Hook structure verification ---

	t.Run("creates correct hook structure", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("file should be valid JSON: %v", err)
		}

		// Verify structure: hooks -> Notification -> [0] -> matcher == "idle_prompt"
		hooks, ok := parsed["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("expected 'hooks' to be an object")
		}
		notifList, ok := hooks["Notification"].([]interface{})
		if !ok {
			t.Fatal("expected 'hooks.Notification' to be an array")
		}
		if len(notifList) != 1 {
			t.Fatalf("expected 1 notification entry, got %d", len(notifList))
		}
		entry, ok := notifList[0].(map[string]interface{})
		if !ok {
			t.Fatal("expected notification entry to be an object")
		}
		if entry["matcher"] != "idle_prompt" {
			t.Errorf("expected matcher 'idle_prompt', got %v", entry["matcher"])
		}
		hooksList, ok := entry["hooks"].([]interface{})
		if !ok {
			t.Fatal("expected entry.hooks to be an array")
		}
		if len(hooksList) != 1 {
			t.Fatalf("expected 1 hook, got %d", len(hooksList))
		}
		hook, ok := hooksList[0].(map[string]interface{})
		if !ok {
			t.Fatal("expected hook to be an object")
		}
		if hook["type"] != "command" {
			t.Errorf("expected hook type 'command', got %v", hook["type"])
		}
		if hook["command"] != defaultCmd {
			t.Errorf("expected hook command %q, got %v", defaultCmd, hook["command"])
		}
	})

	t.Run("creates parent directories if needed", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "nested", "settings.json")

		result, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != HookAdded {
			t.Errorf("expected HookAdded, got %v", result)
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("file should have been created")
		}
	})

	t.Run("preserves existing keys when adding hook", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		existing := `{"permissions": {"allow": ["Bash"]}, "apiKey": "test123"}`
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("file should be valid JSON: %v", err)
		}

		if _, ok := parsed["hooks"]; !ok {
			t.Error("should have hooks key")
		}
		if _, ok := parsed["permissions"]; !ok {
			t.Error("should preserve permissions key")
		}
		if _, ok := parsed["apiKey"]; !ok {
			t.Error("should preserve apiKey key")
		}
	})

	t.Run("file ends with newline", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "settings.json")

		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := AddSoundHook(path, defaultCmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(string(data), "\n") {
			t.Error("file should end with a newline")
		}
	})
}
