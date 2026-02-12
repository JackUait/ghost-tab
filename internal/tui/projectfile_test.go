package tui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
	"github.com/jackuait/ghost-tab/internal/tui"
)

func TestAppendProject(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")

	err := tui.AppendProject("my-app", "/home/user/my-app", file)
	if err != nil {
		t.Fatalf("AppendProject: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "my-app:/home/user/my-app\n" {
		t.Errorf("File content: expected 'my-app:/home/user/my-app\\n', got %q", string(data))
	}
}

func TestAppendProject_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("first:/tmp/first\n"), 0644)

	err := tui.AppendProject("second", "/tmp/second", file)
	if err != nil {
		t.Fatalf("AppendProject: %v", err)
	}

	data, _ := os.ReadFile(file)
	expected := "first:/tmp/first\nsecond:/tmp/second\n"
	if string(data) != expected {
		t.Errorf("File content: expected %q, got %q", expected, string(data))
	}
}

func TestAppendProject_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sub", "dir", "projects")

	err := tui.AppendProject("test", "/tmp/test", file)
	if err != nil {
		t.Fatalf("AppendProject should create parent dirs: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "test:/tmp/test\n" {
		t.Errorf("File content: expected 'test:/tmp/test\\n', got %q", string(data))
	}
}

func TestRemoveProject(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("first:/tmp/first\nsecond:/tmp/second\nthird:/tmp/third\n"), 0644)

	err := tui.RemoveProject("second:/tmp/second", file)
	if err != nil {
		t.Fatalf("RemoveProject: %v", err)
	}

	data, _ := os.ReadFile(file)
	expected := "first:/tmp/first\nthird:/tmp/third\n"
	if string(data) != expected {
		t.Errorf("File content: expected %q, got %q", expected, string(data))
	}
}

func TestRemoveProject_SingleEntry(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("only:/tmp/only\n"), 0644)

	err := tui.RemoveProject("only:/tmp/only", file)
	if err != nil {
		t.Fatalf("RemoveProject: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "" {
		t.Errorf("File content should be empty, got %q", string(data))
	}
}

func TestRemoveProject_NoMatch(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	original := "first:/tmp/first\nsecond:/tmp/second\n"
	os.WriteFile(file, []byte(original), 0644)

	err := tui.RemoveProject("nonexistent:/tmp/nope", file)
	if err != nil {
		t.Fatalf("RemoveProject with no match: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != original {
		t.Errorf("File should be unchanged, got %q", string(data))
	}
}

func TestRemoveProject_PartialMatchNotDeleted(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	original := "app:/tmp/app\napp-long:/tmp/app-long\n"
	os.WriteFile(file, []byte(original), 0644)

	err := tui.RemoveProject("app:/tmp/app", file)
	if err != nil {
		t.Fatalf("RemoveProject: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "app-long:/tmp/app-long\n" {
		t.Errorf("Partial match should survive, got %q", string(data))
	}
}

func TestIsDuplicateProject(t *testing.T) {
	projects := []models.Project{
		{Name: "app", Path: "/home/user/app"},
		{Name: "web", Path: "/home/user/web"},
	}

	if !tui.IsDuplicateProject("/home/user/app", projects) {
		t.Error("Should detect duplicate path")
	}
	if tui.IsDuplicateProject("/home/user/new", projects) {
		t.Error("Should not flag non-duplicate")
	}
}

func TestIsDuplicateProject_TrailingSlash(t *testing.T) {
	projects := []models.Project{
		{Name: "app", Path: "/home/user/app"},
	}

	if !tui.IsDuplicateProject("/home/user/app/", projects) {
		t.Error("Should detect duplicate even with trailing slash")
	}
}

func TestAppendProject_ErrorOnUnwritablePath(t *testing.T) {
	// /dev/null is not a directory, so MkdirAll should fail
	err := tui.AppendProject("test", "/tmp/test", "/dev/null/sub/projects")
	if err == nil {
		t.Error("Should return error when parent cannot be created")
	}
}

func TestRemoveProject_ErrorOnMissingFile(t *testing.T) {
	err := tui.RemoveProject("foo:/tmp/foo", "/nonexistent/projects")
	if err == nil {
		t.Error("Should return error when file doesn't exist")
	}
}

// --- AppendProject edge cases (ported from project-actions.bats) ---

func TestAppendProject_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		projPath string
		expected string
	}{
		{
			name:     "handles paths with spaces",
			projName: "myapp",
			projPath: "/tmp/path with spaces",
			expected: "myapp:/tmp/path with spaces\n",
		},
		{
			name:     "handles name with spaces",
			projName: "my app",
			projPath: "/path/to/app",
			expected: "my app:/path/to/app\n",
		},
		{
			name:     "handles path with quotes",
			projName: "app",
			projPath: `/path/with"quotes`,
			expected: "app:/path/with\"quotes\n",
		},
		{
			name:     "handles path with unicode",
			projName: "app",
			projPath: "/path/\u00e9moji/\U0001F47B",
			expected: "app:/path/\u00e9moji/\U0001F47B\n",
		},
		{
			name:     "handles very long paths",
			projName: "app",
			projPath: strings.Repeat("/very/long/path", 50),
			expected: "app:" + strings.Repeat("/very/long/path", 50) + "\n",
		},
		{
			name:     "handles name with colons",
			projName: "app:v2.0",
			projPath: "/path/to/app",
			expected: "app:v2.0:/path/to/app\n",
		},
		{
			name:     "handles special characters in name",
			projName: "app-v1.0_test",
			projPath: "/path/to/app",
			expected: "app-v1.0_test:/path/to/app\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "projects")

			err := tui.AppendProject(tt.projName, tt.projPath, file)
			if err != nil {
				t.Fatalf("AppendProject: %v", err)
			}

			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("File content: expected %q, got %q", tt.expected, string(data))
			}
		})
	}
}

// --- RemoveProject edge cases (ported from project-actions.bats) ---

func TestRemoveProject_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		remove   string
		expected string
	}{
		{
			name:     "handles paths with spaces",
			initial:  "app1:/path/app1\napp2:/tmp/path with spaces\napp3:/path/app3\n",
			remove:   "app2:/tmp/path with spaces",
			expected: "app1:/path/app1\napp3:/path/app3\n",
		},
		{
			name:     "handles entry with quotes",
			initial:  "app1:/path/app1\napp2:/path/with\"quotes\napp3:/path/app3\n",
			remove:   "app2:/path/with\"quotes",
			expected: "app1:/path/app1\napp3:/path/app3\n",
		},
		{
			name:     "handles entry with unicode",
			initial:  "app1:/path/app1\napp2:/path/\u00e9moji/\U0001F47B\napp3:/path/app3\n",
			remove:   "app2:/path/\u00e9moji/\U0001F47B",
			expected: "app1:/path/app1\napp3:/path/app3\n",
		},
		{
			name:     "does not delete partial matches",
			initial:  "app:/path/app\napp-long:/path/app-longer-name\n",
			remove:   "app:/path/app",
			expected: "app-long:/path/app-longer-name\n",
		},
		{
			name:     "handles very long entries",
			initial:  "app1:/path/app1\napp2:" + strings.Repeat("/very/long/path", 50) + "\napp3:/path/app3\n",
			remove:   "app2:" + strings.Repeat("/very/long/path", 50),
			expected: "app1:/path/app1\napp3:/path/app3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "projects")
			os.WriteFile(file, []byte(tt.initial), 0644)

			err := tui.RemoveProject(tt.remove, file)
			if err != nil {
				t.Fatalf("RemoveProject: %v", err)
			}

			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("File content: expected %q, got %q", tt.expected, string(data))
			}
		})
	}
}

// --- IsDuplicateProject edge cases (ported from project-actions.bats) ---

func TestIsDuplicateProject_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		projects []models.Project
		expected bool
	}{
		{
			name:     "returns false for empty projects list",
			path:     "/home/user/myapp",
			projects: []models.Project{},
			expected: false,
		},
		{
			name:     "returns false for nil projects list",
			path:     "/home/user/myapp",
			projects: nil,
			expected: false,
		},
		{
			name:     "handles path with spaces",
			path:     "/tmp/path with spaces",
			projects: []models.Project{{Name: "app", Path: "/tmp/path with spaces"}},
			expected: true,
		},
		{
			name:     "handles path with single quotes",
			path:     "/tmp/path'with'quotes",
			projects: []models.Project{{Name: "app", Path: "/tmp/path'with'quotes"}},
			expected: true,
		},
		{
			name:     "handles path with double quotes",
			path:     `/tmp/path"with"quotes`,
			projects: []models.Project{{Name: "app", Path: `/tmp/path"with"quotes`}},
			expected: true,
		},
		{
			name:     "handles path with unicode",
			path:     "/tmp/\u00e9moji\U0001F47B",
			projects: []models.Project{{Name: "app", Path: "/tmp/\u00e9moji\U0001F47B"}},
			expected: true,
		},
		{
			name:     "handles multiple trailing slashes",
			path:     "/home/user/myapp///",
			projects: []models.Project{{Name: "app", Path: "/home/user/myapp"}},
			expected: true,
		},
		{
			name:     "detects duplicate with different trailing slash variations",
			path:     "/home/user/myapp///",
			projects: []models.Project{{Name: "app", Path: "/home/user/myapp/"}},
			expected: true,
		},
		{
			name:     "trailing slash on existing path only",
			path:     "/home/user/myapp",
			projects: []models.Project{{Name: "app", Path: "/home/user/myapp/"}},
			expected: true,
		},
		{
			name:     "non-duplicate with spaces is not flagged",
			path:     "/tmp/path with spaces/other",
			projects: []models.Project{{Name: "app", Path: "/tmp/path with spaces"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.IsDuplicateProject(tt.path, tt.projects)
			if result != tt.expected {
				t.Errorf("IsDuplicateProject(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
