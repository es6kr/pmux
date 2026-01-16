package ttyd

import (
	"github.com/spf13/cobra"

	"github.com/es6kr/pmux/internal/plugin"
	"github.com/es6kr/pmux/internal/ttyd"
)

func init() {
	plugin.Register(&plugin.Plugin{
		Name:        "ttyd",
		Description: "Web terminal using ttyd",
		Binary:      "ttyd",
		RegisterCmd: registerCommands,
	})
}

func registerCommands(root *cobra.Command) {
	var ttydPort int

	ttydCmd := &cobra.Command{
		Use:   "ttyd",
		Short: "Manage web terminal (ttyd)",
	}

	ttydStartCmd := &cobra.Command{
		Use:   "start [channel]",
		Short: "Start ttyd for a session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := "main"
			if len(args) > 0 {
				channel = args[0]
			}
			return ttyd.Start(channel, ttydPort)
		},
	}
	ttydStartCmd.Flags().IntVarP(&ttydPort, "port", "p", 7681, "Port for ttyd")

	ttydStopCmd := &cobra.Command{
		Use:   "stop [channel]",
		Short: "Stop ttyd for a session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := "main"
			if len(args) > 0 {
				channel = args[0]
			}
			return ttyd.Stop(channel)
		},
	}

	ttydListCmd := &cobra.Command{
		Use:     "list",
		Short:   "List running ttyd instances",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			ttyd.List()
		},
	}

	ttydCmd.AddCommand(ttydStartCmd, ttydStopCmd, ttydListCmd)
	root.AddCommand(ttydCmd)
}
