package ttyd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPidFile(t *testing.T) {
	result := pidFile("test")
	expected := filepath.Join(pidDir, "test.json")

	if result != expected {
		t.Errorf("pidFile(%q) = %q, want %q", "test", result, expected)
	}
}

func TestGetInstanceNotFound(t *testing.T) {
	inst, err := GetInstance("nonexistent-channel-12345")
	if err != nil {
		t.Fatalf("GetInstance() failed: %v", err)
	}
	if inst != nil {
		t.Error("GetInstance() should return nil for nonexistent channel")
	}
}

func TestPidDirExists(t *testing.T) {
	// init() should create the pid directory
	if _, err := os.Stat(pidDir); os.IsNotExist(err) {
		t.Errorf("pidDir %q does not exist", pidDir)
	}
}
