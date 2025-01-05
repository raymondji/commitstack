package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/spf13/cobra"
)

var switchBranchFlag bool

func init() {
	switchCmd.Flags().BoolVarP(&switchBranchFlag, "branch", "b", false, "Switch branches in the current stack")
}

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw"},
	Short:   "Switch stacks or branches within the current stack",
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

		var currCommit string
		var inference commitstack.InferenceResult
		err = concurrent.Run(
			context.Background(),
			func(ctx context.Context) error {
				var err error
				currCommit, err = git.GetShortCommitHash("HEAD")
				return err
			},
			func(ctx context.Context) error {
				log, err := git.LogAll(defaultBranch)
				if err != nil {
					return err
				}
				inference, err = commitstack.InferStacks(git, log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("switchCmd", "got curr commit and stack inference")
		defer func() {
			printProblems(inference)
		}()

		var currStack *commitstack.Stack
		if switchBranchFlag {
			stack, err := commitstack.GetCurrent(inference.InferredStacks, currCommit)
			if err != nil {
				return err
			}
			currStack = &stack
		}

		var target string
		var formTitle string
		var opts []huh.Option[string]
		if switchBranchFlag {
			formTitle = "Choose branch"
			for _, b := range currStack.OrderedBranches() {
				opts = append(opts, huh.NewOption(b, b))
			}
		} else {
			formTitle = "Choose stack"
			for _, s := range inference.InferredStacks {
				opts = append(opts, huh.NewOption(s.Name(), s.Name()))
			}
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(formTitle).
					Options(opts...).
					Value(&target),
			),
		)
		err = form.Run()
		if err != nil {
			return err
		}

		if err := git.Checkout(target); err != nil {
			return err
		}
		fmt.Printf("Switched to branch '%s'\n", target)
		return nil
	},
}
