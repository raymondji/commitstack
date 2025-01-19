package main

import (
	"context"
	"fmt"

	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/raymondji/git-stack-cli/stackparser"
	"github.com/spf13/cobra"
)

var rebaseInteractiveFlag bool
var rebaseKeepBaseFlag bool
var rebaseForceFlag bool

func init() {
	rebaseCmd.Flags().BoolVarP(&rebaseInteractiveFlag, "interactive", "i", false, "Use interactive rebase")
	rebaseCmd.Flags().BoolVarP(&rebaseKeepBaseFlag, "keep-base", "k", false, "See git rebase --keep-base flag")
	rebaseCmd.Flags().BoolVarP(&rebaseForceFlag, "force", "f", false, "Allow rebasing when some loss of state could occur")
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
		var currBranch, currCommit string
		var log libgit.Log
		err = concurrent.Run(
			context.Background(),
			func(ctx context.Context) error {
				var err error
				currCommit, err = git.GetShortCommitHash("HEAD")
				return err
			},
			func(ctx context.Context) error {
				var err error
				currBranch, err = git.GetCurrentBranch()
				return err
			},
			func(ctx context.Context) error {
				var err error
				log, err = git.LogAll(defaultBranch)
				return err
			},
		)
		if err != nil {
			return err
		}
		stacks, err := stackparser.ParseStacks(log)
		if err != nil {
			return err
		}
		currStack, err := stackparser.GetCurrent(stacks, currCommit)
		if err != nil {
			return err
		}

		if !rebaseForceFlag {
			if currBranch != currStack.Name {
				return fmt.Errorf("warning: rebase from the tip of the stack to maintain the association between branches, use --force to allow")
			}
			if len(currStack.DivergesFrom()) > 0 {
				return fmt.Errorf("warning: rebasing may lose the association between divergent stacks, use --force to allow")
			}
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
			fmt.Printf("Successfully rebased %s on %s\n", currStack.Name, newBase)
		}
		return nil
	},
}
