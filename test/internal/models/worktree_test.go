package models_test

import (
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func TestParseWorktreeListPorcelain(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int // number of non-main worktrees
	}{
		{
			name:   "no worktrees (main only)",
			output: "worktree /Users/jack/ghost-tab\nHEAD abc123\nbranch refs/heads/main\n\n",
			want:   0,
		},
		{
			name: "two additional worktrees",
			output: "worktree /Users/jack/ghost-tab\nHEAD abc123\nbranch refs/heads/main\n\n" +
				"worktree /Users/jack/wt/feature-auth\nHEAD def456\nbranch refs/heads/feature/auth\n\n" +
				"worktree /Users/jack/wt/fix-cleanup\nHEAD 789abc\nbranch refs/heads/fix/cleanup\n\n",
			want: 2,
		},
		{
			name: "worktree with detached HEAD",
			output: "worktree /Users/jack/ghost-tab\nHEAD abc123\nbranch refs/heads/main\n\n" +
				"worktree /Users/jack/wt/detached\nHEAD def456\ndetached\n\n",
			want: 1,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worktrees := models.ParseWorktreeListPorcelain(tt.output)
			if len(worktrees) != tt.want {
				t.Errorf("got %d worktrees, want %d", len(worktrees), tt.want)
			}
		})
	}
}

func TestParseWorktreeListPorcelain_BranchNames(t *testing.T) {
	output := "worktree /Users/jack/ghost-tab\nHEAD abc123\nbranch refs/heads/main\n\n" +
		"worktree /Users/jack/wt/feature-auth\nHEAD def456\nbranch refs/heads/feature/auth\n\n" +
		"worktree /Users/jack/wt/fix-cleanup\nHEAD 789abc\nbranch refs/heads/fix/cleanup\n\n"

	worktrees := models.ParseWorktreeListPorcelain(output)
	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}
	if worktrees[0].Branch != "feature/auth" {
		t.Errorf("worktree 0 branch: got %q, want %q", worktrees[0].Branch, "feature/auth")
	}
	if worktrees[0].Path != "/Users/jack/wt/feature-auth" {
		t.Errorf("worktree 0 path: got %q, want %q", worktrees[0].Path, "/Users/jack/wt/feature-auth")
	}
	if worktrees[1].Branch != "fix/cleanup" {
		t.Errorf("worktree 1 branch: got %q, want %q", worktrees[1].Branch, "fix/cleanup")
	}
}

func TestParseWorktreeListPorcelain_DetachedHead(t *testing.T) {
	output := "worktree /main\nHEAD abc\nbranch refs/heads/main\n\n" +
		"worktree /detached\nHEAD def456\ndetached\n\n"

	worktrees := models.ParseWorktreeListPorcelain(output)
	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}
	if worktrees[0].Branch != "(detached)" {
		t.Errorf("detached branch: got %q, want %q", worktrees[0].Branch, "(detached)")
	}
}

func TestLoadProjectsWithWorktrees_NonGitDir(t *testing.T) {
	// A temp dir that's not a git repo should produce 0 worktrees
	tmpDir := t.TempDir()
	projects := []models.Project{
		{Name: "no-git", Path: tmpDir},
	}

	models.PopulateWorktrees(projects)

	if len(projects[0].Worktrees) != 0 {
		t.Errorf("expected 0 worktrees for non-git dir, got %d", len(projects[0].Worktrees))
	}
}
