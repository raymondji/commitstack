package main

import (
	"errors"
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newLogCmd(git libgit.Git, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Log all commits in a stack",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}

			var stack commitstack.Stack
			if len(args) == 0 {
				stack, err = stacks.GetCurrent()
				if errors.Is(err, commitstack.ErrNotInAStack) {
					fmt.Println("Not in a stack")
					printProblems(stacks)
					return nil
				} else if err != nil {
					printProblems(stacks)
					return err
				}
			} else {
				wantStack := args[0]
				var found bool
				for _, s := range stacks.Entries {
					if s.Name() == wantStack {
						stack = s
						found = true
					}
				}
				if !found {
					return fmt.Errorf("no stack named: %s", wantStack)
				}
			}

			for _, c := range stack.Commits {
				var hereMarker string
				if c.LocalBranch != nil && c.LocalBranch.Current {
					hereMarker = "*"
				} else {
					hereMarker = " "
				}
				var branchCol string
				if c.LocalBranch != nil {
					branchCol = fmt.Sprintf("(%s) ", secondaryColor.Render(c.LocalBranch.Name))
				}

				fmt.Printf("%s %s %s%s\n", hereMarker, primaryColor.Render(c.Hash), branchCol, c.Subject)
			}
			return nil
		},
	}
}
