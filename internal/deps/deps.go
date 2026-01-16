package deps

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/es6kr/pmux/internal/plugin"
)

// Dependency represents an external dependency
type Dependency struct {
	Name       string
	Command    string
	Version    string
	Installed  bool
	InstallCmd string
	Required   bool // true = required, false = optional
}

// Check checks all dependencies and returns their status
func Check() []Dependency {
	isMac := runtime.GOOS == "darwin"

	deps := []Dependency{
		{
			Name:       "tmux",
			Command:    "tmux",
			InstallCmd: getInstallCmd("tmux", isMac),
			Required:   true,
		},
		{
			Name:       "tmuxinator",
			Command:    "tmuxinator",
			InstallCmd: getTmuxinatorInstallCmd(isMac),
			Required:   false,
		},
		{
			Name:       "ttyd",
			Command:    "ttyd",
			InstallCmd: getInstallCmd("ttyd", isMac),
			Required:   false,
		},
	}

	for i := range deps {
		deps[i].Version, deps[i].Installed = checkCommand(deps[i].Command)
	}

	return deps
}

func checkCommand(cmd string) (version string, installed bool) {
	// Try to get version
	var versionFlag string
	switch cmd {
	case "tmux":
		versionFlag = "-V"
	case "tmuxinator":
		versionFlag = "version"
	case "ttyd":
		versionFlag = "--version"
	default:
		versionFlag = "--version"
	}

	out, err := exec.Command(cmd, versionFlag).Output()
	if err != nil {
		return "", false
	}

	version = strings.TrimSpace(string(out))
	// Extract just the version number
	parts := strings.Fields(version)
	if len(parts) > 1 {
		version = parts[len(parts)-1]
	}

	return version, true
}

func getInstallCmd(pkg string, isMac bool) string {
	if isMac {
		return fmt.Sprintf("brew install %s", pkg)
	}

	// Linux fallback
	switch pkg {
	case "tmux":
		return "apt install tmux  # or: yum install tmux"
	case "ttyd":
		return "# See: https://github.com/tsl0922/ttyd#installation"
	default:
		return fmt.Sprintf("# Install %s manually", pkg)
	}
}

func getTmuxinatorInstallCmd(isMac bool) string {
	if isMac {
		return "brew install tmuxinator"
	}
	return "gem install tmuxinator"
}

// RequiredInstalled returns true if all required dependencies are installed
func RequiredInstalled() bool {
	for _, d := range Check() {
		if d.Required && !d.Installed {
			return false
		}
	}
	return true
}

// PrintStatus prints dependency and plugin status to stdout
func PrintStatus() {
	deps := Check()
	plugins := plugin.GetAll()

	// === Dependencies ===
	fmt.Println("=== Dependencies ===")

	requiredOk := true
	for _, d := range deps {
		tag := "[optional]"
		if d.Required {
			tag = "[required]"
		}

		if d.Installed {
			fmt.Printf("✓ %-15s %-15s %s\n", d.Name, "("+d.Version+")", tag)
		} else {
			fmt.Printf("✗ %-15s %-15s %s\n", d.Name, "not found", tag)
			if d.Required {
				fmt.Printf("  → Run: %s\n", d.InstallCmd)
				requiredOk = false
			}
		}
	}

	// === Plugins ===
	fmt.Println()
	fmt.Println("=== Plugins ===")

	for _, p := range plugins {
		if p.Enabled {
			fmt.Printf("✓ %-15s [installed]\n", p.Name)
		} else {
			fmt.Printf("✗ %-15s [not installed]\n", p.Name)
		}
	}

	// Summary
	fmt.Println()
	if requiredOk {
		installedCount := 0
		for _, p := range plugins {
			if p.Enabled {
				installedCount++
			}
		}
		if installedCount == 0 {
			fmt.Println("Run 'pmux plugin install <name>' to enable plugins")
		}
	} else {
		fmt.Println("⚠ Install required dependencies first")
	}
}
