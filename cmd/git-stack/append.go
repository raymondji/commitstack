package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var appendCmd = &cobra.Command{
	Use:   "append [branch_name]",
	Short: "Checkout a new branch and create an initial commit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch
		currBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}

		if currBranch != defaultBranch {
			stacks, err := commitstack.InferStacks(git, defaultBranch)
			if err != nil {
				return err
			}
			stack, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			if currBranch != stack.Name() {
				return fmt.Errorf("must be on the tip of the stack to add another branch, currently checked out: %s, tip: %s",
					currBranch, stack.Name())
			}
		}

		branchName := args[0]
		if err := git.CreateBranch(branchName); err != nil {
			return err
		}
		fmt.Printf("Switched to a new branch '%s'\n", branchName)
		return git.CommitEmpty(fmt.Sprintf("Start of %s", branchName))
	},
}
