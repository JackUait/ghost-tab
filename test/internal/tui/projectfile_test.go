package tui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/jackuait/ghost-tab/internal/tui"
)

// --- AppendProject tests ---

func TestAppendProject_creates_file_and_writes_entry(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	err := tui.AppendProject("myapp", "/home/user/myapp", fp)
	if err != nil {
		t.Fatalf("AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "myapp:/home/user/myapp\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestAppendProject_appends_to_existing_file(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	existing := "first:/path/first\n"
	if err := os.WriteFile(fp, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.AppendProject("second", "/path/second", fp)
	if err != nil {
		t.Fatalf("AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "first:/path/first\nsecond:/path/second\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestAppendProject_creates_parent_directories(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "deep", "nested", "dir", "projects")

	err := tui.AppendProject("proj", "/some/path", fp)
	if err != nil {
		t.Fatalf("AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "proj:/some/path\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestAppendProject_handles_paths_with_spaces(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	err := tui.AppendProject("my app", "/home/user/my projects/app", fp)
	if err != nil {
		t.Fatalf("AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "my app:/home/user/my projects/app\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestAppendProject_handles_paths_with_colons(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	err := tui.AppendProject("winapp", "C:/Users/path", fp)
	if err != nil {
		t.Fatalf("AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "winapp:C:/Users/path\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestAppendProject_multiple_appends(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	err := tui.AppendProject("alpha", "/path/alpha", fp)
	if err != nil {
		t.Fatalf("first AppendProject returned error: %v", err)
	}

	err = tui.AppendProject("beta", "/path/beta", fp)
	if err != nil {
		t.Fatalf("second AppendProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "alpha:/path/alpha\nbeta:/path/beta\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

// --- RemoveProject tests ---

func TestRemoveProject_removes_exact_line(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	content := "first:/path/first\nsecond:/path/second\nthird:/path/third\n"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.RemoveProject("second:/path/second", fp)
	if err != nil {
		t.Fatalf("RemoveProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "first:/path/first\nthird:/path/third\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestRemoveProject_file_unchanged_when_line_not_found(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	content := "first:/path/first\nsecond:/path/second\n"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.RemoveProject("nonexistent:/nope", fp)
	if err != nil {
		t.Fatalf("RemoveProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	want := "first:/path/first\nsecond:/path/second\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestRemoveProject_removes_last_entry_leaves_empty_file(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	content := "only:/path/only\n"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.RemoveProject("only:/path/only", fp)
	if err != nil {
		t.Fatalf("RemoveProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)
	if got != "" {
		t.Errorf("file content = %q, want empty string", got)
	}
}

func TestRemoveProject_preserves_other_lines(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	content := "a:/path/a\nb:/path/b\nc:/path/c\nd:/path/d\n"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.RemoveProject("b:/path/b", fp)
	if err != nil {
		t.Fatalf("RemoveProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)

	if strings.Contains(got, "b:/path/b") {
		t.Error("removed line should not be present")
	}
	for _, line := range []string{"a:/path/a", "c:/path/c", "d:/path/d"} {
		if !strings.Contains(got, line) {
			t.Errorf("expected line %q to be preserved", line)
		}
	}
}

func TestRemoveProject_returns_error_for_nonexistent_file(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "does_not_exist")

	err := tui.RemoveProject("line", fp)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestRemoveProject_handles_duplicate_lines(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "projects")

	content := "dup:/path/dup\nother:/path/other\ndup:/path/dup\n"
	if err := os.WriteFile(fp, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	err := tui.RemoveProject("dup:/path/dup", fp)
	if err != nil {
		t.Fatalf("RemoveProject returned error: %v", err)
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	got := string(data)

	if strings.Contains(got, "dup:/path/dup") {
		t.Error("both duplicate lines should have been removed")
	}
	want := "other:/path/other\n"
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

// --- IsDuplicateProject tests ---

func TestIsDuplicateProject_finds_exact_match(t *testing.T) {
	projects := []models.Project{
		{Name: "alpha", Path: "/home/user/alpha"},
		{Name: "beta", Path: "/home/user/beta"},
	}

	if !tui.IsDuplicateProject("/home/user/alpha", projects) {
		t.Error("expected true for exact match, got false")
	}
}

func TestIsDuplicateProject_no_match(t *testing.T) {
	projects := []models.Project{
		{Name: "alpha", Path: "/home/user/alpha"},
		{Name: "beta", Path: "/home/user/beta"},
	}

	if tui.IsDuplicateProject("/home/user/gamma", projects) {
		t.Error("expected false for non-matching path, got true")
	}
}

func TestIsDuplicateProject_ignores_trailing_slash(t *testing.T) {
	projects := []models.Project{
		{Name: "proj", Path: "/home/user/proj"},
	}

	// Input has trailing slash, stored path does not
	if !tui.IsDuplicateProject("/home/user/proj/", projects) {
		t.Error("expected true when input has trailing slash, got false")
	}

	// Stored path has trailing slash, input does not
	projectsWithSlash := []models.Project{
		{Name: "proj", Path: "/home/user/proj/"},
	}
	if !tui.IsDuplicateProject("/home/user/proj", projectsWithSlash) {
		t.Error("expected true when stored path has trailing slash, got false")
	}
}

func TestIsDuplicateProject_empty_list(t *testing.T) {
	var projects []models.Project

	if tui.IsDuplicateProject("/any/path", projects) {
		t.Error("expected false for empty project list, got true")
	}
}
