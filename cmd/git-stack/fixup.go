package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/commitstack/commitstack"
	"github.com/raymondji/commitstack/libgit"
	"github.com/spf13/cobra"
)

var fixupAddFlag bool
var fixupRebaseFlag bool

func init() {
	fixupCmd.Flags().BoolVarP(&fixupAddFlag, "add", "a", false, "Equivalent to git commit -a")
	fixupCmd.Flags().BoolVarP(&fixupRebaseFlag, "rebase", "r", false, "Perform a git rebase after")
}

var fixupCmd = &cobra.Command{
	Use:   "fixup",
	Short: "Create a commit to fixup a branch in the stack",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch := deps.git, deps.repoCfg.DefaultBranch

		stacks, err := commitstack.ComputeAll(git, defaultBranch)
		if err != nil {
			return err
		}
		stack, err := stacks.GetCurrent()
		if err != nil {
			return err
		}

		var branchToFix string
		if len(args) == 1 {
			branchToFix = args[0]
		} else {
			var opts []huh.Option[string]
			for _, b := range stack.LocalBranches() {
				opts = append(opts, huh.NewOption(b.Name, b.Name))
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

		hash, err := git.GetCommitHash(branchToFix)
		if err != nil {
			return err
		}

		res, err := git.CommitFixup(hash, fixupAddFlag)
		if err != nil {
			return err
		}
		fmt.Println(res)

		if fixupRebaseFlag {
			// Hack(raymond): --autosquash only works with interactive rebase, so use
			// GIT_SEQUENCE_EDITOR=true to accept the changes automatically.
			res, err := git.Rebase(defaultBranch, libgit.RebaseOpts{
				Env:            []string{"GIT_SEQUENCE_EDITOR=true"},
				AdditionalArgs: []string{"--keep-base", "--autosquash"},
				Interactive:    true,
			})
			if err != nil {
				return err
			}
			fmt.Println(res)
		}
		return nil
	},
}
