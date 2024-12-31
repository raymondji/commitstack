package main

import (
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
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

		stacks, err := commitstack.ComputeAll(git, defaultBranch)
		if err != nil {
			return err
		}

		for _, s := range stacks.Entries {
			var name string
			if s.Current() {
				name = "* " + theme.PrimaryColor.Render(s.Name())
			} else {
				name = "  " + s.Name()
			}

			fmt.Printf("%s (%d branches)\n", name, len(s.LocalBranches()))
		}

		printProblems(stacks)
		return nil
	},
}
