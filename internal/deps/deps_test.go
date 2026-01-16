package deps

import (
	"runtime"
	"testing"
)

func TestCheck(t *testing.T) {
	deps := Check()

	if len(deps) != 3 {
		t.Errorf("Check() returned %d deps, want 3", len(deps))
	}

	// Verify expected dependency names
	expectedNames := map[string]bool{
		"tmux":       false,
		"tmuxinator": false,
		"ttyd":       false,
	}

	for _, d := range deps {
		if _, ok := expectedNames[d.Name]; !ok {
			t.Errorf("Unexpected dependency: %s", d.Name)
		}
		expectedNames[d.Name] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Missing dependency: %s", name)
		}
	}
}

func TestInstallCmdMac(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	deps := Check()

	for _, d := range deps {
		if d.Name == "tmuxinator" {
			if d.InstallCmd != "brew install tmuxinator" {
				t.Errorf("tmuxinator InstallCmd = %q, want brew install", d.InstallCmd)
			}
		}
	}
}

func TestRequiredInstalled(t *testing.T) {
	// Just verify it doesn't panic
	_ = RequiredInstalled()
}
