package main

import (
	"fmt"

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
		&cobra.Command{
			Use:   "version",
			Short: "Prints the application version",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("%s (%s) <%s>\n", version, commit, date)
				return nil
			},
		},
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
