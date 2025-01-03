package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:          "stack",
		Short:        "A CLI tool for managing stacked Git branches.",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		initCmd,
		appendCmd,
		switchCmd,
		pushCmd,
		rebaseCmd,
		fixupCmd,
		showCmd,
		listCmd,
		logCmd,
		learnCmd,
		versionCmd,
	)

	rootCmd.Execute()
}
