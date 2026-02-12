package bash_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// ghosttyConfigSnippet builds a bash snippet that sources tui.sh and ghostty-config.sh,
// then runs the provided bash code.
func ghosttyConfigSnippet(t *testing.T, body string) string {
	t.Helper()
	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	ghosttyConfigPath := filepath.Join(root, "lib", "ghostty-config.sh")
	return fmt.Sprintf("source %q && source %q && %s", tuiPath, ghosttyConfigPath, body)
}

// --- merge_ghostty_config ---

func TestMergeGhosttyConfig_replaces_existing_command_line(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command = /old/path\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "command = /new/path" {
		t.Errorf("expected config to be %q, got %q", "command = /new/path", content)
	}
}

func TestMergeGhosttyConfig_appends_when_no_command_line_exists(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := string(data)
	assertContains(t, content, "font-size = 14")
	assertContains(t, content, "command = /new/path")
}

// --- backup_replace_ghostty_config ---

func TestBackupReplaceGhosttyConfig_creates_backup_and_replaces(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Backed up")
	assertContains(t, out, "Replaced")

	// Verify the config was replaced
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if strings.TrimSpace(string(data)) != "new content" {
		t.Errorf("expected config to be %q, got %q", "new content", strings.TrimSpace(string(data)))
	}

	// Verify a backup file exists
	matches, err := filepath.Glob(filepath.Join(tmpDir, "config.backup.*"))
	if err != nil {
		t.Fatalf("failed to glob backups: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 backup file, found %d", len(matches))
	}
}

// --- Edge Cases: Malformed Config Files ---

func TestMergeGhosttyConfig_handles_empty_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if strings.TrimSpace(string(data)) != "command = /new/path" {
		t.Errorf("expected config to be %q, got %q", "command = /new/path", strings.TrimSpace(string(data)))
	}
}

func TestMergeGhosttyConfig_handles_file_with_only_whitespace(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "   \n\n  \t\t  \n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "command = /new/path")
}

func TestMergeGhosttyConfig_handles_file_with_windows_line_endings(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\r\ncommand = /old/path\r\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "command = /new/path")
}

func TestMergeGhosttyConfig_handles_command_line_with_extra_spaces(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command     =     /old/path\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "command = /new/path" {
		t.Errorf("expected config to be %q, got %q", "command = /new/path", content)
	}
}

func TestMergeGhosttyConfig_handles_command_line_with_tabs(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command\t=\t/old/path\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "command = /new/path" {
		t.Errorf("expected config to be %q, got %q", "command = /new/path", content)
	}
}

func TestMergeGhosttyConfig_handles_command_line_with_no_spaces_around_equals(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command=/old/path\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content != "command = /new/path" {
		t.Errorf("expected config to be %q, got %q", "command = /new/path", content)
	}
}

func TestMergeGhosttyConfig_handles_multiple_command_lines_replaces_all(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command = /first\nfont-size = 14\ncommand = /second\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	// sed should replace ALL matches
	count := strings.Count(string(data), "command = /new/path")
	if count != 2 {
		t.Errorf("expected 2 occurrences of 'command = /new/path', got %d\nfull content: %s", count, string(data))
	}
}

func TestMergeGhosttyConfig_handles_binary_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(configFile, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, 0644); err != nil {
		t.Fatalf("failed to write binary file: %v", err)
	}

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	// sed should append even to binary
	assertContains(t, out, "Appended")
}

func TestMergeGhosttyConfig_handles_very_large_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config")

	var builder strings.Builder
	for i := 1; i <= 1000; i++ {
		fmt.Fprintf(&builder, "# Comment line %d\n", i)
	}
	builder.WriteString("font-size = 14\n")
	if err := os.WriteFile(configFile, []byte(builder.String()), 0644); err != nil {
		t.Fatalf("failed to write large file: %v", err)
	}

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "command = /new/path")
}

func TestMergeGhosttyConfig_handles_command_with_special_characters(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command = /old/path\n")

	// Use single quotes in the bash snippet to avoid shell expansion
	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q 'command = /path/with/$VAR/and/`+"`"+`cmd`+"`"+`'`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "/path/with/$VAR/and/`cmd`")
}

func TestMergeGhosttyConfig_handles_command_with_quotes(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "command = /old/path\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q 'command = "/path/with spaces"'`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Replaced")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), `"/path/with spaces"`)
}

func TestMergeGhosttyConfig_handles_file_with_no_trailing_newline(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	assertContains(t, string(data), "command = /new/path")
}

func TestMergeGhosttyConfig_handles_commented_out_command_line(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "# command = /commented\nfont-size = 14\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Appended")

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := string(data)
	assertContains(t, content, "# command = /commented")
	assertContains(t, content, "command = /new/path")
}

// --- Edge Cases: backup_replace_ghostty_config ---

func TestBackupReplaceGhosttyConfig_handles_empty_source_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Backed up")
	assertContains(t, out, "Replaced")

	// Verify config is now empty
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(data) != "" {
		t.Errorf("expected config to be empty, got %q", string(data))
	}
}

func TestBackupReplaceGhosttyConfig_handles_very_large_source_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")

	var builder strings.Builder
	for i := 1; i <= 1000; i++ {
		fmt.Fprintf(&builder, "line %d\n", i)
	}
	sourceFile := writeTempFile(t, tmpDir, "source", builder.String())

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Backed up")
	assertContains(t, out, "Replaced")

	// Verify config has 1000 lines
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1000 {
		t.Errorf("expected 1000 lines, got %d", len(lines))
	}
}

func TestBackupReplaceGhosttyConfig_handles_binary_source_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := filepath.Join(tmpDir, "source")
	if err := os.WriteFile(sourceFile, []byte{0x00, 0x01, 0x02, 0x03, 0x04}, 0644); err != nil {
		t.Fatalf("failed to write binary source: %v", err)
	}

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Backed up")
	assertContains(t, out, "Replaced")

	// Verify config now contains binary data by checking file command output
	checkSnippet := fmt.Sprintf(`file %q`, configFile)
	fileOut, _ := runBashSnippet(t, checkSnippet, nil)
	assertContains(t, fileOut, "data")
}

func TestBackupReplaceGhosttyConfig_handles_windows_line_endings_in_source(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "line1\r\nline2\r\nline3\r\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	assertContains(t, out, "Backed up")
	assertContains(t, out, "Replaced")

	// Verify content was copied
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	content := string(data)
	assertContains(t, content, "line1")
	assertContains(t, content, "line3")
}

func TestBackupReplaceGhosttyConfig_creates_multiple_backups_without_collision(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "content1\n")
	source1File := writeTempFile(t, tmpDir, "source1", "source1\n")
	source2File := writeTempFile(t, tmpDir, "source2", "source2\n")

	// Create first backup
	snippet1 := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, source1File))
	_, code1 := runBashSnippet(t, snippet1, nil)
	assertExitCode(t, code1, 0)

	// Sleep to ensure different timestamp
	time.Sleep(1100 * time.Millisecond)

	// Create second backup
	snippet2 := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, source2File))
	_, code2 := runBashSnippet(t, snippet2, nil)
	assertExitCode(t, code2, 0)

	// Verify two backup files exist
	matches, err := filepath.Glob(filepath.Join(tmpDir, "config.backup.*"))
	if err != nil {
		t.Fatalf("failed to glob backups: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("expected 2 backup files, found %d", len(matches))
	}
}

// --- Edge Cases: Permission Denied ---

func TestMergeGhosttyConfig_handles_read_only_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\n")
	if err := os.Chmod(configFile, 0444); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	defer os.Chmod(configFile, 0644) // cleanup

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// sed -i fails but function continues and prints success (no set -e)
	assertContains(t, out, "Appended")
}

func TestBackupReplaceGhosttyConfig_handles_read_only_config(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")
	if err := os.Chmod(configFile, 0444); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	defer os.Chmod(configFile, 0644) // cleanup

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// Functions don't use set -e, so they succeed even with errors
	assertContains(t, out, "Backed up")
}

func TestBackupReplaceGhosttyConfig_handles_unreadable_source_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")
	if err := os.Chmod(sourceFile, 0000); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	defer os.Chmod(sourceFile, 0644) // cleanup

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// Functions don't use set -e, so they print success even with cp errors
	assertContains(t, out, "Backed up")
}

func TestBackupReplaceGhosttyConfig_handles_unwritable_directory_for_backup(t *testing.T) {
	tmpDir := t.TempDir()
	readonlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readonlyDir, 0755); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}
	configFile := writeTempFile(t, readonlyDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")
	if err := os.Chmod(readonlyDir, 0555); err != nil {
		t.Fatalf("failed to chmod: %v", err)
	}
	defer os.Chmod(readonlyDir, 0755) // cleanup

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// cp fails but function continues (no set -e)
	assertContains(t, out, "Backed up")
}

// --- Edge Cases: Missing Files ---

func TestMergeGhosttyConfig_handles_missing_config_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "nonexistent")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`merge_ghostty_config %q "command = /new/path"`, configFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// grep fails but echo >> succeeds, function prints success (no set -e)
	assertContains(t, out, "Appended")
}

func TestBackupReplaceGhosttyConfig_handles_missing_config_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "nonexistent")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// Functions don't check for errors, print success anyway
	assertContains(t, out, "Backed up")
}

func TestBackupReplaceGhosttyConfig_handles_missing_source_file(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := filepath.Join(tmpDir, "nonexistent")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	out, _ := runBashSnippet(t, snippet, nil)
	// cp fails but function continues and prints success
	assertContains(t, out, "Backed up")
}

// --- Edge Cases: Concurrent Operations ---

func TestMergeGhosttyConfig_handles_concurrent_modifications(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "font-size = 14\n")

	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	ghosttyConfigPath := filepath.Join(root, "lib", "ghostty-config.sh")

	// Run two simultaneous merges via a bash snippet that backgrounds both
	snippet := fmt.Sprintf(
		`source %q && source %q && `+
			`merge_ghostty_config %q "command = /path1" > /dev/null 2>&1 & `+
			`pid1=$!; `+
			`merge_ghostty_config %q "command = /path2" > /dev/null 2>&1 & `+
			`pid2=$!; `+
			`wait "$pid1"; wait "$pid2"; `+
			`cat %q`,
		tuiPath, ghosttyConfigPath, configFile, configFile, configFile)

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	// At least one command line should be appended
	assertContains(t, out, "command = ")
}

func TestBackupReplaceGhosttyConfig_backup_filename_contains_numeric_timestamp(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "old content\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")

	snippet := ghosttyConfigSnippet(t,
		fmt.Sprintf(`backup_replace_ghostty_config %q %q`, configFile, sourceFile))

	_, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)

	// Verify backup filename ends with a numeric timestamp (digits only)
	matches, err := filepath.Glob(filepath.Join(tmpDir, "config.backup.*"))
	if err != nil {
		t.Fatalf("failed to glob backups: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one backup file")
	}

	backupFile := matches[0]
	// Extract timestamp from filename: config.backup.<timestamp>
	parts := strings.SplitN(filepath.Base(backupFile), ".backup.", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected backup filename format: %s", backupFile)
	}
	timestamp := parts[1]
	matched, err := regexp.MatchString(`^[0-9]+$`, timestamp)
	if err != nil {
		t.Fatalf("regexp error: %v", err)
	}
	if !matched {
		t.Errorf("expected timestamp to be all digits, got %q", timestamp)
	}
}

func TestBackupReplaceGhosttyConfig_handles_file_modified_during_backup(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := writeTempFile(t, tmpDir, "config", "original\n")
	sourceFile := writeTempFile(t, tmpDir, "source", "new content\n")

	root := projectRoot(t)
	tuiPath := filepath.Join(root, "lib", "tui.sh")
	ghosttyConfigPath := filepath.Join(root, "lib", "ghostty-config.sh")

	// Start backup in background, modify config during backup
	snippet := fmt.Sprintf(
		`source %q && source %q && `+
			`backup_replace_ghostty_config %q %q > /dev/null 2>&1 & `+
			`pid1=$!; `+
			`sleep 0.05; `+
			`echo "modified" >> %q; `+
			`wait "$pid1"; `+
			`cat %q`,
		tuiPath, ghosttyConfigPath, configFile, sourceFile, configFile, configFile)

	out, code := runBashSnippet(t, snippet, nil)
	assertExitCode(t, code, 0)
	// Due to race condition, config might have both or just source content
	// Just verify replacement happened
	assertContains(t, out, "new content")
}
