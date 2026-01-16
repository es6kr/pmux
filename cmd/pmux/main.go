package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/es6kr/pmux/internal/deps"
	"github.com/es6kr/pmux/internal/plugin"
	_ "github.com/es6kr/pmux/internal/plugins" // register built-in plugins
	"github.com/es6kr/pmux/internal/server"
	"github.com/es6kr/pmux/internal/session"
	"github.com/es6kr/pmux/internal/tmuxinator"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:     "pmux",
		Short:   "Plugin-based tmux session manager",
		Version: version,
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start [channel]",
		Short: "Start a new tmux session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := "main"
			if len(args) > 0 {
				channel = args[0]
			}

			sess, err := session.Create(channel)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			fmt.Printf("Session '%s' started at %s\n", sess.Channel, sess.Socket)
			return nil
		},
	}

	// Stop command
	stopCmd := &cobra.Command{
		Use:   "stop [channel]",
		Short: "Stop a tmux session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := "main"
			if len(args) > 0 {
				channel = args[0]
			}

			if err := session.Kill(channel); err != nil {
				return fmt.Errorf("failed to stop session: %w", err)
			}

			fmt.Printf("Session '%s' stopped\n", channel)
			return nil
		},
	}

	// Attach command
	attachCmd := &cobra.Command{
		Use:   "attach [channel]",
		Short: "Attach to an existing session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := "main"
			if len(args) > 0 {
				channel = args[0]
			}

			return session.Attach(channel)
		},
	}

	// List command
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "List all sessions",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := session.List()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No active sessions")
				return nil
			}

			fmt.Printf("%-15s %-30s %-10s\n", "CHANNEL", "SOCKET", "CLIENTS")
			for _, s := range sessions {
				fmt.Printf("%-15s %-30s %-10d\n", s.Channel, s.Socket, s.Clients)
			}
			return nil
		},
	}

	// Web command
	var webPort int
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Start web management UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Start(webPort)
		},
	}
	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to listen on")

	// Doctor command
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check dependencies and plugins",
		Run: func(cmd *cobra.Command, args []string) {
			deps.PrintStatus()
		},
	}

	// Plugin subcommands
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	pluginListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available plugins",
		Run: func(cmd *cobra.Command, args []string) {
			plugin.List()
		},
	}

	pluginInstallCmd := &cobra.Command{
		Use:   "install <name>",
		Short: "Install a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return plugin.Install(args[0])
		},
	}

	pluginRemoveCmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "uninstall"},
		Short:   "Remove a plugin",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return plugin.Remove(args[0])
		},
	}

	pluginCmd.AddCommand(pluginListCmd, pluginInstallCmd, pluginRemoveCmd)

	// tmuxinator subcommands (alias: project, proj)
	muxCmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj", "mux"},
		Short:   "Manage tmuxinator projects",
	}

	muxListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tmuxinator projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := tmuxinator.List()
			if err != nil {
				return err
			}

			if len(projects) == 0 {
				fmt.Println("No tmuxinator projects")
				return nil
			}

			fmt.Printf("%-20s %s\n", "NAME", "PATH")
			for _, p := range projects {
				fmt.Printf("%-20s %s\n", p.Name, p.Path)
			}
			return nil
		},
	}

	muxStartCmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start a tmuxinator project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmuxinator.Start(args[0])
		},
	}

	muxStopCmd := &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop a tmuxinator project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmuxinator.Stop(args[0])
		},
	}

	muxNewCmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new tmuxinator project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmuxinator.New(args[0])
		},
	}

	muxEditCmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit a tmuxinator project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmuxinator.Edit(args[0])
		},
	}

	muxDeleteCmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm"},
		Short:   "Delete a tmuxinator project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return tmuxinator.Delete(args[0])
		},
	}

	muxCmd.AddCommand(muxListCmd, muxStartCmd, muxStopCmd, muxNewCmd, muxEditCmd, muxDeleteCmd)

	rootCmd.AddCommand(startCmd, stopCmd, attachCmd, listCmd, webCmd, doctorCmd, pluginCmd, muxCmd)

	// Register installed plugin commands (builtin + external)
	plugin.RegisterCommands(rootCmd)
	plugin.RegisterExternalCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
