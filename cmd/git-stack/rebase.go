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
	Use:     "rebase [newbase]",
	Aliases: []string{"r"},
	Short:   "Rebase the current stack",
	Long:    "A convenience wrapper around git rebase, with some nicer defaults for stacked branches.",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

		log, err := git.LogAll(defaultBranch)
		if err != nil {
			return err
		}
		inference, err := commitstack.InferStacks(git, log)
		if err != nil {
			return err
		}
		currCommit, err := git.GetShortCommitHash("HEAD")
		if err != nil {
			return err
		}
		s, err := commitstack.GetCurrent(inference.InferredStacks, currCommit)
		if err != nil {
			return err
		}
		currBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}
		if currBranch != s.Name() {
			return fmt.Errorf("must be on the tip of the stack to edit, currently checked out: %s, tip: %s",
				currBranch, s.Name())
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
			fmt.Printf("Successfully rebased %s on %s\n", s.Name(), newBase)
		}
		return nil
	},
}
