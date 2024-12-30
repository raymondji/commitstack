package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newSwitchCmd(git libgit.Git, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "switch",
		Short: "Switch to a stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}

			var target string
			var opts []huh.Option[string]
			for _, s := range stacks.Entries {
				opts = append(opts, huh.NewOption(s.Name(), s.Name()))
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose stack").
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
}
