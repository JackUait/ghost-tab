package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackuait/ghost-tab/internal/models"
)

func TestLoadProjects_StaleField_ExistingPath(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "myproject")
	os.MkdirAll(realDir, 0755)
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("myproject:"+realDir+"\n"), 0644)

	projects, err := models.LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Stale {
		t.Error("expected Stale=false for existing path")
	}
}

func TestLoadProjects_StaleField_MissingPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "projects")
	os.WriteFile(file, []byte("ghost:/nonexistent/path/xyz\n"), 0644)

	projects, err := models.LoadProjects(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if !projects[0].Stale {
		t.Error("expected Stale=true for missing path")
	}
}
