package main

import (
	"errors"
	"fmt"

	"github.com/raymondji/commitstack/sampleusage"
	"github.com/spf13/cobra"
)

var learnPartFlag int
var learnExecFlag bool
var learnCleanupFlag bool

func init() {
	learnCmd.Flags().IntVar(&learnPartFlag, "part", 1, "Which part of the tutorial to continue from")
	learnCmd.Flags().BoolVar(&learnExecFlag, "exec", false, "Whether to execute the commands in the tutorial")
	learnCmd.Flags().BoolVar(&learnCleanupFlag, "cleanup", false, "Whether to cleanup everything created as part of the tutorial")
}

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Prints sample commands to learn how to use commitstack",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		samples := sampleusage.New(deps.theme, deps.repoCfg.DefaultBranch, deps.git, deps.host)
		if learnCleanupFlag {
			if err := samples.Cleanup(); err != nil {
				return err
			}
		}

		switch learnPartFlag {
		case 1:
			if learnExecFlag {
				if err := samples.Part1().Execute(); err != nil {
					return err
				}
			} else {
				fmt.Println(samples.Part1().String())
			}
		case 2:
			fmt.Println("TODO")
		default:
			return errors.New("invalid tutorial part")
		}

		return nil
	},
}
