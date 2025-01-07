package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/raymondji/git-stack-cli/stackparser"
	"github.com/spf13/cobra"
)

var fixupAddFlag bool
var fixupRebaseFlag bool

func init() {
	fixupCmd.Flags().BoolVarP(&fixupAddFlag, "add", "a", false, "Use git commit -a")
	fixupCmd.Flags().BoolVarP(&fixupRebaseFlag, "rebase", "r", false, "Automatically perform a git rebase")
}

var fixupCmd = &cobra.Command{
	Use:     "fixup [branch]",
	Aliases: []string{"f"},
	Short:   "Create a commit to fixup a branch in the current stack",
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
		stacks, err := stackparser.InferStacks(log)
		if err != nil {
			return err
		}
		currCommit, err := git.GetShortCommitHash("HEAD")
		if err != nil {
			return err
		}
		stack, err := stackparser.GetCurrent(stacks, currCommit)
		if err != nil {
			return err
		}

		var branchToFix string
		if len(args) == 1 {
			branchToFix = args[0]
		} else {
			branches, err := stack.TotalOrderedBranches()
			if err != nil {
				return err
			}

			var opts []huh.Option[string]
			for _, b := range branches {
				opts = append(opts, huh.NewOption(b, b))
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose which branch to fixup").
						Options(opts...).
						Filtering(true).
						Value(&branchToFix),
				),
			)
			err = form.Run()
			if err != nil {
				return err
			}
		}

		hash, err := git.GetShortCommitHash(branchToFix)
		if err != nil {
			return err
		}

		res, err := git.CommitFixup(hash, fixupAddFlag)
		if err != nil {
			return err
		}
		fmt.Println(res)

		if fixupRebaseFlag {
			res, err := git.Rebase(defaultBranch, libgit.RebaseOpts{
				Autosquash: true,
				UpdateRefs: true,
				KeepBase:   true,
			})
			if err != nil {
				return err
			}
			fmt.Println(res)
		}
		return nil
	},
}
