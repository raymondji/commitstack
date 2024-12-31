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
		versionCmd,
		initCmd,
		addCmd,
		switchCmd,
		pushCmd,
		pullCmd,
		editCmd,
		fixupCmd,
		showCmd,
		listCmd,
		logCmd,
	)

	rootCmd.Execute()
}
