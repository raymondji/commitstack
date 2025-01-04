package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version.Version)
		return nil
	},
}
