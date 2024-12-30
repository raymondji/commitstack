package main

import (
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newAddCmd(git libgit.Git, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "add [branch_name]",
		Short: "Start a new stack or add a new branch onto the current stack",
		Long:  "Start a new stack if not currently in one, or add a new branch onto the current stack",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}
			stack, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			currBranch, err := git.GetCurrentBranch()
			if err != nil {
				return err
			}
			if currBranch != stack.Name() {
				return fmt.Errorf("must be on the tip of the stack to add another branch, currently checked out: %s, tip: %s",
					currBranch, stack.Name())
			}

			branchName := args[0]
			if err := git.CreateBranch(branchName); err != nil {
				return err
			}
			fmt.Printf("Switched to a new branch '%s'\n", branchName)
			return git.CommitEmpty(fmt.Sprintf("Start of %s", branchName))
		},
	}
}
