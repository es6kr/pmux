package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	socketPath   = "/tmp"
	socketPrefix = "pmux-"
)

// Session represents a pmux session
type Session struct {
	Channel   string
	Socket    string
	CreatedAt time.Time
	Clients   int
	Active    bool
}

// GetSocketPath returns the socket path for a channel
func GetSocketPath(channel string) string {
	return filepath.Join(socketPath, socketPrefix+channel)
}

// Create creates a new tmux session
func Create(channel string) (*Session, error) {
	socket := GetSocketPath(channel)

	// Check if session already exists
	if _, err := os.Stat(socket); err == nil {
		sessions, _ := List()
		for _, s := range sessions {
			if s.Channel == channel {
				return &s, nil
			}
		}
	}

	// Create new tmux session
	cmd := exec.Command("tmux", "-S", socket, "new-session", "-d", "-s", channel)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tmux new-session failed: %w", err)
	}

	// Set permissions for multi-user access
	if err := os.Chmod(socket, 0777); err != nil {
		return nil, fmt.Errorf("chmod failed: %w", err)
	}

	return &Session{
		Channel:   channel,
		Socket:    socket,
		CreatedAt: time.Now(),
		Clients:   0,
		Active:    true,
	}, nil
}

// Attach attaches to an existing session (replaces current process)
func Attach(channel string) error {
	socket := GetSocketPath(channel)

	if _, err := os.Stat(socket); os.IsNotExist(err) {
		return fmt.Errorf("session '%s' does not exist", channel)
	}

	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Replace current process with tmux attach
	return syscall.Exec(tmuxPath, []string{
		"tmux", "-S", socket, "attach-session", "-t", channel,
	}, os.Environ())
}

// Kill terminates a session
func Kill(channel string) error {
	socket := GetSocketPath(channel)

	if _, err := os.Stat(socket); os.IsNotExist(err) {
		return fmt.Errorf("session '%s' does not exist", channel)
	}

	cmd := exec.Command("tmux", "-S", socket, "kill-session", "-t", channel)
	if err := cmd.Run(); err != nil {
		// Try to clean up socket anyway
		os.Remove(socket)
		return nil
	}

	// Clean up socket
	os.Remove(socket)
	return nil
}

// List returns all active sessions
func List() ([]Session, error) {
	pattern := filepath.Join(socketPath, socketPrefix+"*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, socket := range matches {
		channel := strings.TrimPrefix(filepath.Base(socket), socketPrefix)

		// Check if tmux server is running on this socket
		cmd := exec.Command("tmux", "-S", socket, "list-sessions")
		output, err := cmd.Output()

		sess := Session{
			Channel: channel,
			Socket:  socket,
			Active:  err == nil && len(output) > 0,
		}

		if sess.Active {
			// Get client count
			clientCmd := exec.Command("tmux", "-S", socket, "list-clients")
			clientOutput, _ := clientCmd.Output()
			if len(clientOutput) > 0 {
				sess.Clients = len(strings.Split(strings.TrimSpace(string(clientOutput)), "\n"))
			}
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// SendKeys sends keys to a session
func SendKeys(channel, keys string) error {
	socket := GetSocketPath(channel)

	if _, err := os.Stat(socket); os.IsNotExist(err) {
		return fmt.Errorf("session '%s' does not exist", channel)
	}

	cmd := exec.Command("tmux", "-S", socket, "send-keys", "-t", channel, keys, "Enter")
	return cmd.Run()
}

// CapturePane captures the current pane content
func CapturePane(channel string) (string, error) {
	socket := GetSocketPath(channel)

	if _, err := os.Stat(socket); os.IsNotExist(err) {
		return "", fmt.Errorf("session '%s' does not exist", channel)
	}

	cmd := exec.Command("tmux", "-S", socket, "capture-pane", "-t", channel, "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
