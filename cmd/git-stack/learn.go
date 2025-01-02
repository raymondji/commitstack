package main

import (
	"errors"
	"fmt"

	"github.com/raymondji/commitstack/sampleusage"
	"github.com/spf13/cobra"
)

var learnPartFlag int
var learnModeFlag string

const (
	learnModePrint   string = "print"
	learnModeExec    string = "exec"
	learnModeCleanup string = "cleanup"
)

func init() {
	learnCmd.Flags().IntVar(&learnPartFlag, "part", 1, "Which part of the tutorial to continue from")
	learnCmd.Flags().StringVar(&learnModeFlag, "mode", learnModePrint, "Which mode to use.")
}

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Prints sample commands to learn how to use commitstack",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}

		var sample sampleusage.Sample
		switch learnPartFlag {
		case 1:
			sample = sampleusage.Basics(deps.git, deps.host, deps.repoCfg.DefaultBranch, deps.theme)
		case 2:
			sample = sampleusage.Basics(deps.git, deps.host, deps.repoCfg.DefaultBranch, deps.theme)
		default:
			return errors.New("invalid tutorial part number")
		}

		switch learnModeFlag {
		case learnModePrint:
			fmt.Println(sample.String())
		case learnModeExec:
			if err := sample.Cleanup(); err != nil {
				return err
			}
			if err := sample.Execute(); err != nil {
				return err
			}
		case learnModeCleanup:
			if err := sample.Cleanup(); err != nil {
				return err
			}
		default:
			return errors.New("invalid mode flag")
		}

		return nil
	},
}
