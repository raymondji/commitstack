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

		currBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}
		log, err := git.LogAll(defaultBranch)
		if err != nil {
			return err
		}
		inference, err := commitstack.InferStacks(git, log)
		if err != nil {
			return err
		}
		defer func() {
			printProblems(inference)
		}()

		var target, formTitle string
		var opts []huh.Option[string]
		if switchBranchFlag {
			stack, err := commitstack.GetCurrent(inference.InferredStacks, currBranch)
			if errors.Is(err, commitstack.ErrUnableToInferCurrentStack) {
				fmt.Println(err.Error())
				return nil
			} else if err != nil {
				return err
			}

			formTitle = "Choose branch"
			for _, b := range stack.LocalBranches() {
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
