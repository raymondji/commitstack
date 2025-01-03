package main

import (
	"errors"
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Log all commits in a stack",
	Args:  cobra.MaximumNArgs(1),
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

		var stack commitstack.Stack
		if len(args) == 0 {
			stack, err = stacks.GetCurrent()
			if errors.Is(err, commitstack.ErrUnableToInferCurrentStack) {
				fmt.Println(err.Error())
				printProblems(stacks)
				return nil
			} else if err != nil {
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
				branchCol = fmt.Sprintf("(%s) ", theme.SecondaryColor.Render(c.LocalBranch.Name))
			}

			fmt.Printf("%s %s %s%s\n", hereMarker, theme.PrimaryColor.Render(c.Hash), branchCol, c.Subject)
		}

		printProblem(stack)
		return nil
	},
}
