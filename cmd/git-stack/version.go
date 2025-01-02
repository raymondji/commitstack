package main

import (
	"fmt"

	"github.com/raymondji/commitstack/release/releasevars"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the application version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf(releasevars.Version)
		return nil
	},
}
