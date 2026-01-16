package theia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const packageJSON = `{
  "name": "pmux-theia",
  "version": "1.0.0",
  "dependencies": {
    "@theia/core": "latest",
    "@theia/editor": "latest",
    "@theia/filesystem": "latest",
    "@theia/monaco": "latest",
    "@theia/navigator": "latest",
    "@theia/terminal": "latest",
    "@theia/workspace": "latest",
    "@theia/vsx-registry": "latest",
    "@theia/plugin-ext": "latest",
    "@theia/plugin-ext-vscode": "latest"
  },
  "devDependencies": {
    "@theia/cli": "latest"
  },
  "theia": {
    "frontend": {
      "config": {
        "applicationName": "pmux IDE"
      }
    }
  }
}
`

func theiaDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pmux", "theia")
}

// EnsureInstalled ensures Theia is installed and built
func EnsureInstalled() error {
	dir := theiaDir()

	// Create directory if not exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create theia directory: %w", err)
	}

	pkgPath := filepath.Join(dir, "package.json")
	libPath := filepath.Join(dir, "lib")

	// Check if package.json exists
	needsInstall := false
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		needsInstall = true
		fmt.Println("Setting up Theia IDE (first run)...")

		// Write package.json
		if err := os.WriteFile(pkgPath, []byte(packageJSON), 0644); err != nil {
			return fmt.Errorf("failed to write package.json: %w", err)
		}
	}

	// Check if node_modules exists
	nodeModulesPath := filepath.Join(dir, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		needsInstall = true
	}

	if needsInstall {
		// Install dependencies
		fmt.Println("Installing dependencies (this may take a while)...")
		cmd := exec.Command("npm", "install")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}

	// Check if lib directory exists (build output)
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		// Build Theia
		fmt.Println("Building Theia...")
		buildCmd := exec.Command("npx", "theia", "build")
		buildCmd.Dir = dir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			return fmt.Errorf("failed to build theia: %w", err)
		}

		fmt.Println("✓ Theia IDE setup complete")
	}

	return nil
}

// Start starts Theia IDE for a directory
func Start(dir string, port int) error {
	if err := EnsureInstalled(); err != nil {
		return err
	}

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist", absDir)
	}

	fmt.Printf("Starting Theia IDE at http://localhost:%d\n", port)
	fmt.Printf("Working directory: %s\n", absDir)

	cmd := exec.Command("npx", "theia", "start", absDir,
		"--hostname", "0.0.0.0",
		"--port", fmt.Sprintf("%d", port),
	)
	cmd.Dir = theiaDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
