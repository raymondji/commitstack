package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/spf13/cobra"
)

var fixupAddFlag bool
var fixupRebaseFlag bool

func init() {
	fixupCmd.Flags().BoolVarP(&fixupAddFlag, "add", "a", false, "Equivalent to git commit -a")
	fixupCmd.Flags().BoolVarP(&fixupRebaseFlag, "rebase", "r", false, "Perform a git rebase after")
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
		stacks, err := commitstack.InferStacks(git, log)
		if err != nil {
			return err
		}
		currCommit, err := git.GetShortCommitHash("HEAD")
		if err != nil {
			return err
		}
		stack, err := commitstack.GetCurrent(stacks.InferredStacks, currCommit)
		if err != nil {
			return err
		}

		var branchToFix string
		if len(args) == 1 {
			branchToFix = args[0]
		} else {
			var opts []huh.Option[string]
			for _, b := range stack.AllBranches() {
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
				Interactive: true,
				Autosquash:  true,
				UpdateRefs:  true,
				KeepBase:    true,
			})
			if err != nil {
				return err
			}
			fmt.Println(res)
		}
		return nil
	},
}
