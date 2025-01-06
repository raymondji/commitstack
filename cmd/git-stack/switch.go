package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/inference"
	"github.com/spf13/cobra"
)

var switchBranchFlag bool

func init() {
	switchCmd.Flags().BoolVarP(&switchBranchFlag, "branch", "b", false, "Switch branches in the current stack")
}

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"s"},
	Short:   "Switch stacks or branches within the current stack",
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

		var currCommit string
		var stacks []inference.Stack
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
				stacks, err = inference.InferStacks(log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("switchCmd", "got curr commit and stack inference")

		var currStack *inference.Stack
		if switchBranchFlag {
			stack, err := inference.GetCurrent(stacks, currCommit)
			if err != nil {
				return err
			}
			currStack = &stack
		}

		var target string
		var formTitle string
		var opts []huh.Option[string]
		if switchBranchFlag {
			branches, err := currStack.TotalOrderedBranches()
			var errNoTotalOrder inference.NoTotalOrderError
			if errors.As(err, &errNoTotalOrder) {
				// TODO: check for this specific error type
				fmt.Printf("Warning: stack %s does not have a total order\n", currStack.Name)
				fmt.Println("Branches are displayed in lexicographic order")
				branches = currStack.Branches()
			} else if err != nil {
				return err
			}

			formTitle = "Choose branch"
			for _, b := range branches {
				opts = append(opts, huh.NewOption(b, b))
			}
		} else {
			formTitle = "Choose stack"
			for _, s := range stacks {
				opts = append(opts, huh.NewOption(s.Name, s.Name))
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
