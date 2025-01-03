package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:          "stack",
		Short:        "CLI for managing stacked Git branches",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		initCmd,
		switchCmd,
		pushCmd,
		rebaseCmd,
		fixupCmd,
		showCmd,
		listCmd,
		learnCmd,
		versionCmd,
	)

	rootCmd.Execute()
}
