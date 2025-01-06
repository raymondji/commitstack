package main

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	benchmarkFlag bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&benchmarkFlag, "benchmark", false, "Benchmark commands")
	rootCmd.AddCommand(
		branchCmd,
		fixupCmd,
		initCmd,
		learnCmd,
		listCmd,
		logCmd,
		pushCmd,
		rebaseCmd,
		switchCmd,
		versionCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:          "stack",
	Short:        "CLI for managing stacked Git branches",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if benchmarkFlag {
			benchmarkCheckpoint = time.Now()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		benchmarkPoint("rootCmd", "done")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
