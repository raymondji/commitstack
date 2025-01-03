package main

import (
	"fmt"
	"slices"
	"strings"

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

		var stack commitstack.Stack
		if len(args) == 0 {
			stack, err = commitstack.GetCurrent(inference.InferredStacks, currBranch)
			if err != nil {
				return err
			}
		} else {
			wantStack := args[0]
			var found bool
			for _, s := range inference.InferredStacks {
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
			if slices.Contains(c.LocalBranches, currBranch) {
				hereMarker = "*"
			} else {
				hereMarker = " "
			}
			var branchCol string
			if len(c.LocalBranches) > 0 {
				branchCol = fmt.Sprintf("(%s) ", theme.SecondaryColor.Render(strings.Join(c.LocalBranches, ", ")))
			}

			fmt.Printf("%s %s %s%s\n", hereMarker, theme.PrimaryColor.Render(c.Hash), branchCol, c.Subject)
		}

		return nil
	},
}
