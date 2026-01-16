package theia

import (
	"github.com/spf13/cobra"

	"github.com/es6kr/pmux/internal/plugin"
	"github.com/es6kr/pmux/internal/theia"
)

func init() {
	plugin.Register(&plugin.Plugin{
		Name:        "theia",
		Description: "Theia IDE (lightweight web IDE)",
		Binary:      "npm",
		RegisterCmd: registerCommands,
	})
}

func registerCommands(root *cobra.Command) {
	var port int

	theiaCmd := &cobra.Command{
		Use:   "theia [directory]",
		Short: "Start Theia IDE",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ""
			if len(args) > 0 {
				dir = args[0]
			}
			return theia.Start(dir, port)
		},
	}
	theiaCmd.Flags().IntVarP(&port, "port", "p", 3000, "Port for Theia IDE")

	root.AddCommand(theiaCmd)
}
