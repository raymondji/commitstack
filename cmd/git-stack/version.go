package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	version = "0.2.7"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the application version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version)
		return nil
	},
}
