package session

import (
	"os"
	"testing"
)

func TestGetSocketPath(t *testing.T) {
	tests := []struct {
		channel  string
		expected string
	}{
		{"main", "/tmp/pmux-main"},
		{"dev", "/tmp/pmux-dev"},
		{"test-session", "/tmp/pmux-test-session"},
	}

	for _, tt := range tests {
		t.Run(tt.channel, func(t *testing.T) {
			result := GetSocketPath(tt.channel)
			if result != tt.expected {
				t.Errorf("GetSocketPath(%q) = %q, want %q", tt.channel, result, tt.expected)
			}
		})
	}
}

func TestCreateAndKill(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI environment (requires tmux)")
	}

	channel := "pmux-test-session"

	// Create session
	sess, err := Create(channel)
	if err != nil {
		t.Fatalf("Create(%q) failed: %v", channel, err)
	}

	if sess.Channel != channel {
		t.Errorf("sess.Channel = %q, want %q", sess.Channel, channel)
	}

	if sess.Socket != GetSocketPath(channel) {
		t.Errorf("sess.Socket = %q, want %q", sess.Socket, GetSocketPath(channel))
	}

	// List sessions
	sessions, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s.Channel == channel {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Session %q not found in List()", channel)
	}

	// Kill session
	if err := Kill(channel); err != nil {
		t.Fatalf("Kill(%q) failed: %v", channel, err)
	}

	// Verify killed
	if _, err := os.Stat(GetSocketPath(channel)); !os.IsNotExist(err) {
		t.Errorf("Socket file still exists after Kill()")
	}
}

func TestListEmpty(t *testing.T) {
	// This test just verifies List() doesn't error on empty results
	_, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
}
