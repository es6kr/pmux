package tmuxinator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Project represents a tmuxinator project
type Project struct {
	Name string
	Path string
}

// List returns all tmuxinator projects
func List() ([]Project, error) {
	cmd := exec.Command("tmuxinator", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tmuxinator list failed: %w", err)
	}

	// Parse output: "tmuxinator projects:\nproject1 project2 project3"
	lines := strings.Split(string(output), "\n")
	var projects []Project

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "tmuxinator") {
			continue
		}

		// Projects are space-separated
		names := strings.Fields(line)
		for _, name := range names {
			projects = append(projects, Project{
				Name: name,
				Path: getProjectPath(name),
			})
		}
	}

	return projects, nil
}

// Start starts a tmuxinator project
func Start(name string) error {
	cmd := exec.Command("tmuxinator", "start", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Stop stops a tmuxinator project
func Stop(name string) error {
	cmd := exec.Command("tmuxinator", "stop", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Edit opens a project config in the default editor
func Edit(name string) error {
	cmd := exec.Command("tmuxinator", "edit", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// New creates a new tmuxinator project
func New(name string) error {
	cmd := exec.Command("tmuxinator", "new", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Delete removes a tmuxinator project
func Delete(name string) error {
	cmd := exec.Command("tmuxinator", "delete", name, "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Debug shows the shell commands that would be run
func Debug(name string) (string, error) {
	cmd := exec.Command("tmuxinator", "debug", name)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmuxinator debug failed: %w", err)
	}
	return string(output), nil
}

// getProjectPath returns the path to a project's config file
func getProjectPath(name string) string {
	home, _ := os.UserHomeDir()

	// Check XDG config first
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}

	paths := []string{
		filepath.Join(xdgConfig, "tmuxinator", name+".yml"),
		filepath.Join(home, ".tmuxinator", name+".yml"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return filepath.Join(home, ".tmuxinator", name+".yml")
}

// Installed checks if tmuxinator is installed
func Installed() bool {
	_, err := exec.LookPath("tmuxinator")
	return err == nil
}
