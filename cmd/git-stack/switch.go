package main

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var switchBranchFlag bool

func init() {
	switchCmd.Flags().BoolVarP(&switchBranchFlag, "branch", "b", false, "Switch between branches in the current stack")
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to another stack, or to another branch within the current stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

		stacks, err := commitstack.InferStacks(git, defaultBranch)
		if err != nil {
			return err
		}

		var target, formTitle string
		var opts []huh.Option[string]
		if switchBranchFlag {
			stack, err := stacks.GetCurrent()
			if errors.Is(err, commitstack.ErrUnableToInferCurrentStack) {
				fmt.Println(err.Error())
				printProblems(stacks)
				return nil
			} else if err != nil {
				return err
			}

			formTitle = "Choose branch"
			for _, b := range stack.LocalBranches() {
				opts = append(opts, huh.NewOption(b.Name, b.Name))
			}
		} else {
			formTitle = "Choose stack"
			for _, s := range stacks.Entries {
				opts = append(opts, huh.NewOption(s.Name(), s.Name()))
			}
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(formTitle).
					Options(opts...).
					Filtering(true).
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
