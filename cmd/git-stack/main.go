package main

import (
	"github.com/spf13/cobra"
)

// Set by goreleaser
var (
	version = "unknown"
	commit  = "unknown commit"
	date    = "unknown date"
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
		learnCmd,
	)

	rootCmd.Execute()
}
