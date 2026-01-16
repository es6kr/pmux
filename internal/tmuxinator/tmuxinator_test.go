package tmuxinator

import (
	"os"
	"testing"
)

func TestInstalled(t *testing.T) {
	// Just verify it returns a boolean without error
	result := Installed()
	t.Logf("tmuxinator installed: %v", result)
}

func TestList(t *testing.T) {
	if !Installed() {
		t.Skip("tmuxinator not installed")
	}

	projects, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	t.Logf("Found %d projects", len(projects))
	for _, p := range projects {
		t.Logf("  - %s: %s", p.Name, p.Path)
	}
}

func TestGetProjectPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	path := getProjectPath("test")

	// Should contain either .config/tmuxinator or .tmuxinator
	if path == "" {
		t.Error("getProjectPath returned empty string")
	}

	if len(path) < len(home) {
		t.Errorf("getProjectPath returned invalid path: %s", path)
	}
}
