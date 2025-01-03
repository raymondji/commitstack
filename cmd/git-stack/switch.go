package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var switchBranchFlag bool
var switchCreateFlag bool

func init() {
	switchCmd.Flags().BoolVarP(&switchBranchFlag, "branch", "b", false, "Switch branches in the current stack")
	switchCmd.Flags().BoolVarP(&switchCreateFlag, "create", "c", false, "Create a new stack (or branch in the current stack) and switch to it")
}

var switchCmd = &cobra.Command{
	Use:   "switch [target]",
	Short: "Switch stacks or branches within the current stack",
	Args:  cobra.MaximumNArgs(1),
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

		var currStack *commitstack.Stack
		if switchBranchFlag {
			stack, err := commitstack.GetCurrent(inference.InferredStacks, currBranch)
			if err != nil {
				return err
			}
			currStack = &stack
		}
		if switchBranchFlag && switchCreateFlag {
			if currBranch != currStack.Name() {
				return fmt.Errorf("must be on the tip of the stack, currently checked out: %s, tip: %s",
					currBranch, currStack.Name())
			}
		}

		var target string
		if len(args) == 1 {
			target = args[0]
		} else {
			var formTitle string
			var opts []huh.Option[string]
			if switchBranchFlag {
				formTitle = "Choose branch"
				for _, b := range currStack.AllBranches() {
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
		}

		if switchBranchFlag {
			if switchCreateFlag {
				err := git.CreateBranch(target, currStack.Name())
				if err != nil {
					return err
				}
			} else {
				// Verify this stack contains a branch named <target>
				var found bool
				for _, b := range currStack.AllBranches() {
					if b == target {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("no branch named %s in the current stack %s", target, currStack.Name())
				}
			}
		} else {
			// target represents a stack
			if switchCreateFlag {
				err := git.CreateBranch(target, defaultBranch)
				if err != nil {
					return err
				}
			} else {
				// Verify there exists a stack named <target>
				var found bool
				for _, s := range inference.InferredStacks {
					if s.Name() == target {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("no stack named: %s", target)
				}
			}
		}

		if err := git.Checkout(target); err != nil {
			return err
		}
		if switchCreateFlag {
			err := git.CommitEmpty(fmt.Sprintf("Start of %s", target))
			if err != nil {
				return err
			}
		}

		fmt.Printf("Switched to branch '%s'\n", target)
		return nil
	},
}
