package main

import (
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newListCmd(git libgit.Git, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}

			for _, s := range stacks.Entries {
				var name string
				if s.Current() {
					name = "* " + primaryColor.Render(s.Name())
				} else {
					name = "  " + s.Name()
				}

				fmt.Printf("%s (%d branches)\n", name, len(s.LocalBranches()))
			}

			printProblems(stacks)
			return nil
		},
	}
}
