package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, theme := deps.git, deps.repoCfg.DefaultBranch, deps.theme

		currBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}
		log, err := git.LogAll(defaultBranch)
		if err != nil {
			return err
		}
		inference, err := commitstack.InferStacks(git, log)
		if err != nil {
			return err
		}
		defer func() {
			printProblems(inference)
		}()

		for _, s := range inference.InferredStacks {
			var name string
			if s.IsCurrent(currBranch) {
				name = "* " + theme.PrimaryColor.Render(s.Name())
			} else {
				name = "  " + s.Name()
			}

			fmt.Printf("%s (%d branches)\n", name, len(s.LocalBranches()))
		}

		return nil
	},
}
