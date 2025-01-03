package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/spf13/cobra"
)

var rebaseNoEditFlag bool

func init() {
	rebaseCmd.Flags().BoolVar(&rebaseNoEditFlag, "no-edit", false, "Disable interactive editing")
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase [newbase]",
	Short: "Rebase the current stack",
	Args:  cobra.MaximumNArgs(1),
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

		if len(args) == 0 {
			_, err := git.Rebase(defaultBranch, libgit.RebaseOpts{
				Interactive: !rebaseNoEditFlag,
				UpdateRefs:  true,
				Autosquash:  true,
				KeepBase:    true,
			})
			if err != nil {
				return err
			}
		} else {
			newBase := args[0]
			_, err := git.Rebase(newBase, libgit.RebaseOpts{
				Interactive: !rebaseNoEditFlag,
				UpdateRefs:  true,
				Autosquash:  true,
			})
			if err != nil {
				return err
			}
		}

		if rebaseNoEditFlag {
			fmt.Printf("Successfully rebased %s on %s\n", stack.Name(), defaultBranch)
		}
		return nil
	},
}
