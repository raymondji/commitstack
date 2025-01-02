package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to a stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("0.3.1")
		return nil
	},
}
