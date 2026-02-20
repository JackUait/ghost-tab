package bash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================
// add_project_to_file tests
// ============================================================

func TestAddProjectToFile_appends_entry_to_existing_file(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")
	writeTempFile(t, dir, "projects", "existing-project:/path\n")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"new-app", "/home/user/new-app", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	assertContains(t, content, "existing-project:/path")
	assertContains(t, content, "new-app:/home/user/new-app")
}

func TestAddProjectToFile_creates_file_when_not_exists(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"my-app", "/home/user/my-app", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	assertContains(t, content, "my-app:/home/user/my-app")
}

func TestAddProjectToFile_creates_parent_directories(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "deep", "nested", "dir", "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"app", "/path/to/app", projectsFile}, nil)
	assertExitCode(t, code, 0)

	// Verify parent directories were created
	parentDir := filepath.Join(dir, "deep", "nested", "dir")
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("parent directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", parentDir)
	}

	// Verify file content
	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	assertContains(t, string(data), "app:/path/to/app")
}

func TestAddProjectToFile_writes_name_colon_path_format(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"ghost-tab", "/Users/dev/ghost-tab", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	// The exact line should be "name:path" followed by a newline
	expected := "ghost-tab:/Users/dev/ghost-tab\n"
	if content != expected {
		t.Errorf("expected exact content %q, got %q", expected, content)
	}
}

func TestAddProjectToFile_handles_paths_with_spaces(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"my app", "/home/user/my projects/app dir", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	assertContains(t, content, "my app:/home/user/my projects/app dir")
}

func TestAddProjectToFile_appends_multiple_entries(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"first", "/path/to/first", projectsFile}, nil)
	assertExitCode(t, code, 0)

	_, code = runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"second", "/path/to/second", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	assertContains(t, content, "first:/path/to/first")
	assertContains(t, content, "second:/path/to/second")

	// Verify both entries are separate lines
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), content)
	}
}

func TestAddProjectToFile_file_ends_with_newline(t *testing.T) {
	dir := t.TempDir()
	projectsFile := filepath.Join(dir, "projects")

	_, code := runBashFunc(t, "lib/project-actions.sh", "add_project_to_file",
		[]string{"app", "/path/to/app", projectsFile}, nil)
	assertExitCode(t, code, 0)

	data, err := os.ReadFile(projectsFile)
	if err != nil {
		t.Fatalf("failed to read projects file: %v", err)
	}
	content := string(data)
	if !strings.HasSuffix(content, "\n") {
		t.Errorf("expected file to end with newline, got %q", content)
	}
}
