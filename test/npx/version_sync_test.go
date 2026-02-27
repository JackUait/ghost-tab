package npx_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageJSON_version_matches_VERSION_file(t *testing.T) {
	root := projectRoot(t)

	versionBytes, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(string(versionBytes))

	pkgBytes, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pkg struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(pkgBytes, &pkg); err != nil {
		t.Fatal(err)
	}

	if pkg.Version != expected {
		t.Errorf("package.json version = %q, VERSION file = %q", pkg.Version, expected)
	}
}
