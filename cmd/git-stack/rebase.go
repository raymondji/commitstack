package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/spf13/cobra"
)

var rebaseInteractiveFlag bool
var rebaseKeepBaseFlag bool

func init() {
	rebaseCmd.Flags().BoolVarP(&rebaseInteractiveFlag, "interactive", "i", false, "Use interactive rebase")
	rebaseCmd.Flags().BoolVarP(&rebaseKeepBaseFlag, "keep-base", "k", false, "See git rebase --keep-base flag")
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase [newbase]",
	Short: "Rebase the current stack",
	Long:  "A convenience wrapper around git rebase, with some nicer defaults for stacked branches.",
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

		rebaseOpts := libgit.RebaseOpts{
			UpdateRefs:  true,
			Autosquash:  true,
			Interactive: rebaseInteractiveFlag,
			KeepBase:    rebaseKeepBaseFlag,
		}
		newBase := defaultBranch
		if len(args) == 1 {
			newBase = args[0]
		}
		if _, err := git.Rebase(newBase, rebaseOpts); err != nil {
			return err
		}
		if !rebaseInteractiveFlag {
			fmt.Printf("Successfully rebased %s on %s\n", stack.Name(), defaultBranch)
		}
		return nil
	},
}
