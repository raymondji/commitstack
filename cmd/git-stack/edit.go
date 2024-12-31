package main

import (
	"fmt"

	"github.com/raymondji/commitstack/commitstack"
	"github.com/raymondji/commitstack/libgit"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the stack using interactive rebase",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

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
			return fmt.Errorf("must be on the tip of the stack to edit, currently checked out: %s, tip: %s",
				currBranch, stack.Name())
		}

		if _, err := git.Rebase(defaultBranch, libgit.RebaseOpts{
			Interactive:    true,
			AdditionalArgs: []string{"--keep-base", "--autosquash"},
		}); err != nil {
			return err
		}
		return nil
	},
}
