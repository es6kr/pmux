package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	configDir  = ".pmux"
	configFile = "plugins.json"
)

// Plugin represents a pmux plugin
type Plugin struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Binary      string `json:"binary,omitempty"`
	// RegisterCmd registers plugin's commands to root
	RegisterCmd func(root *cobra.Command) `json:"-"`
}

// Registry holds all available plugins
var Registry = make(map[string]*Plugin)

// Register adds a plugin to the registry
func Register(p *Plugin) {
	Registry[p.Name] = p
}

// Config holds plugin configuration
type Config struct {
	Plugins map[string]bool `json:"plugins"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configDir, configFile)
}

func ensureConfigDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, configDir)
	return os.MkdirAll(dir, 0755)
}

// LoadConfig loads plugin configuration
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Plugins: make(map[string]bool),
	}

	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SaveConfig saves plugin configuration
func SaveConfig(cfg *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0644)
}

// IsInstalled checks if a plugin is installed (enabled)
func IsInstalled(name string) bool {
	cfg, err := LoadConfig()
	if err != nil {
		return false
	}
	return cfg.Plugins[name]
}

// Install enables a plugin
func Install(name string) error {
	p, exists := Registry[name]
	if !exists {
		return fmt.Errorf("unknown plugin: %s", name)
	}

	// Check if required binary exists
	if p.Binary != "" {
		if _, err := findBinary(p.Binary); err != nil {
			return fmt.Errorf("plugin '%s' requires '%s' to be installed\n  Install with: brew install %s", name, p.Binary, p.Binary)
		}
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if cfg.Plugins[name] {
		return fmt.Errorf("plugin '%s' is already installed", name)
	}

	cfg.Plugins[name] = true
	if err := SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Plugin '%s' installed\n", name)
	return nil
}

// Remove disables a plugin
func Remove(name string) error {
	if _, exists := Registry[name]; !exists {
		return fmt.Errorf("unknown plugin: %s", name)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	if !cfg.Plugins[name] {
		return fmt.Errorf("plugin '%s' is not installed", name)
	}

	delete(cfg.Plugins, name)
	if err := SaveConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Plugin '%s' removed\n", name)
	return nil
}

// List shows all available plugins and their status
func List() {
	cfg, _ := LoadConfig()

	fmt.Printf("%-15s %-12s %-10s %s\n", "PLUGIN", "STATUS", "TYPE", "DESCRIPTION")

	// Built-in plugins
	for name, p := range Registry {
		status := "not installed"
		if cfg.Plugins[name] {
			status = "installed"
		}
		fmt.Printf("%-15s %-12s %-10s %s\n", name, status, "builtin", p.Description)
	}

	// External plugins
	for _, name := range DiscoverExternal() {
		if _, exists := Registry[name]; exists {
			continue // Skip if same name as builtin
		}
		fmt.Printf("%-15s %-12s %-10s %s\n", name, "available", "external", "pmux-"+name)
	}
}

// GetAll returns all plugins with their status
func GetAll() []Plugin {
	cfg, _ := LoadConfig()
	var plugins []Plugin

	for name, p := range Registry {
		plugin := *p
		plugin.Enabled = cfg.Plugins[name]
		plugins = append(plugins, plugin)
	}
	return plugins
}

// RegisterCommands registers all installed plugins' commands
func RegisterCommands(root *cobra.Command) {
	cfg, _ := LoadConfig()

	for name, p := range Registry {
		if cfg.Plugins[name] && p.RegisterCmd != nil {
			p.RegisterCmd(root)
		}
	}
}

func findBinary(name string) (string, error) {
	paths := []string{
		"/usr/local/bin/" + name,
		"/opt/homebrew/bin/" + name,
		"/usr/bin/" + name,
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("%s not found", name)
}

// DiscoverExternal finds external plugins (executables named pmux-*)
func DiscoverExternal() []string {
	var plugins []string
	seen := make(map[string]bool)

	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	// Also check ~/.pmux/plugins/
	home, _ := os.UserHomeDir()
	paths = append(paths, filepath.Join(home, configDir, "plugins"))

	for _, dir := range paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			name := f.Name()
			if !strings.HasPrefix(name, "pmux-") {
				continue
			}

			// Check if executable
			fullPath := filepath.Join(dir, name)
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			if info.Mode()&0111 == 0 {
				continue
			}

			pluginName := strings.TrimPrefix(name, "pmux-")
			if !seen[pluginName] {
				seen[pluginName] = true
				plugins = append(plugins, pluginName)
			}
		}
	}

	return plugins
}

// RegisterExternalCommands registers external plugin commands
func RegisterExternalCommands(root *cobra.Command) {
	for _, name := range DiscoverExternal() {
		// Skip if built-in plugin with same name exists and is installed
		if _, exists := Registry[name]; exists {
			continue
		}

		pluginName := name
		cmd := &cobra.Command{
			Use:                name,
			Short:              fmt.Sprintf("External plugin: pmux-%s", name),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				pluginPath, err := exec.LookPath("pmux-" + pluginName)
				if err != nil {
					// Try ~/.pmux/plugins/
					home, _ := os.UserHomeDir()
					pluginPath = filepath.Join(home, configDir, "plugins", "pmux-"+pluginName)
					if _, err := os.Stat(pluginPath); err != nil {
						return fmt.Errorf("plugin not found: pmux-%s", pluginName)
					}
				}

				pluginCmd := exec.Command(pluginPath, args...)
				pluginCmd.Stdin = os.Stdin
				pluginCmd.Stdout = os.Stdout
				pluginCmd.Stderr = os.Stderr
				return pluginCmd.Run()
			},
		}
		root.AddCommand(cmd)
	}
}
