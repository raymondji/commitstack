package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log [stack]",
	Aliases: []string{"lo"},
	Short:   "Log commits in the stack",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch
		benchmarkPoint("logCmd", "got deps")

		var currBranch, currCommit string
		var stacks []stackparser.Stack
		var mergedBranches []string
		err = concurrent.Run(
			context.Background(),
			func(ctx context.Context) error {
				var err error
				currCommit, err = git.GetShortCommitHash("HEAD")
				return err
			},
			func(ctx context.Context) error {
				var err error
				mergedBranches, err = git.GetMergedBranches(defaultBranch)
				return err
			},
			func(ctx context.Context) error {
				var err error
				currBranch, err = git.GetCurrentBranch()
				return err
			},
			func(ctx context.Context) error {
				log, err := git.LogAll(defaultBranch)
				if err != nil {
					return err
				}
				stacks, err = stackparser.InferStacks(log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("logCmd", "got curr commit, curr branch, and stack stackparser")

		var stack stackparser.Stack
		if len(args) == 0 {
			if slices.Contains(mergedBranches, currBranch) {
				fmt.Printf("error: the current branch is not a valid stack (it's merged into %s)\n", defaultBranch)
				return nil
			}
			stack, err = stackparser.GetCurrent(stacks, currCommit)
			if err != nil {
				return err
			}
		} else {
			wantStack := args[0]
			var found bool
			for _, s := range stacks {
				if s.Name == wantStack {
					stack = s
					found = true
				}
			}
			if !found {
				return fmt.Errorf("no stack named: %s", wantStack)
			}
		}
		defer func() {
			printProblems([]stackparser.Stack{stack}, deps.theme)
		}()
		benchmarkPoint("logCmd", "got desired stack")
		if err := git.LogOneline(defaultBranch, stack.Name); err != nil {
			return err
		}
		benchmarkPoint("logCmd", "done")

		return nil
	},
}
