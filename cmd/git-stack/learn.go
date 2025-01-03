package main

import (
	"errors"
	"fmt"

	"github.com/raymondji/git-stack-cli/sampleusage"
	"github.com/spf13/cobra"
)

var learnChapterFlag int
var learnModeFlag string

const (
	learnModePrint   string = "print"
	learnModeExec    string = "exec"
	learnModeCleanup string = "cleanup"
)

func init() {
	learnCmd.Flags().IntVar(&learnChapterFlag, "chapter", 0, "Which chapter of the tutorial to continue from")
	learnCmd.Flags().StringVar(&learnModeFlag, "mode", learnModePrint, "Which mode to use.")
}

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Learn how to use git stack through interactive tutorials",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}

		var sample sampleusage.Sample
		switch learnChapterFlag {
		case 0:
			fmt.Println("Welcome to git stack! The following tutorial(s) will explain the core functionality and show how to use various sample commands.")
			fmt.Println()
			fmt.Println("Recommended: git stack can execute the sample commands in the tutorial automatically and show you their output live. To continue with this option, run:")
			fmt.Println(deps.theme.TertiaryColor.Render("git stack learn --chapter 1 --mode=exec"))
			fmt.Println()
			fmt.Println("Alternative: you can also view the text-only version and follow along if desired by copying the commands and running them yourself. To continue with this option, run:")
			fmt.Println(deps.theme.TertiaryColor.Render("git stack learn --chapter 1"))
			return nil
		case 1:
			sample = sampleusage.Basics(deps.git, deps.host, deps.repoCfg.DefaultBranch, deps.theme)
		case 2:
			sample = sampleusage.Advanced(deps.git, deps.host, deps.repoCfg.DefaultBranch, deps.theme)
		default:
			return errors.New("invalid tutorial chapter number")
		}

		switch learnModeFlag {
		case learnModePrint:
			fmt.Println(sample.String())
			fmt.Println("----")
			fmt.Println()
			fmt.Println("To automatically execute the sample commands in this tutorial and see their output live, run:")
			fmt.Println(deps.theme.TertiaryColor.Render(fmt.Sprintf(
				"git stack learn --chapter %d --mode=exec", learnChapterFlag,
			)))
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
