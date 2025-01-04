package main

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "List all stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, theme := deps.git, deps.repoCfg.DefaultBranch, deps.theme

		benchmarkPoint("listCmd", "got deps")
		currCommit, err := git.GetShortCommitHash("HEAD")
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "got current commit")
		log, err := git.LogAll(defaultBranch)
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "done git log")

		inference, err := commitstack.InferStacks(git, log)
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "done stack inference")
		defer func() {
			printProblems(inference)
		}()

		for _, s := range inference.InferredStacks {
			var name, suffix string
			if s.IsCurrent(currCommit) {
				name = "* " + theme.PrimaryColor.Render(s.Name())
			} else {
				name = "  " + s.Name()
			}

			all := s.AllBranches()
			if len(all) == 1 {
				suffix = theme.TertiaryColor.Render("(1 branch)")
			} else {
				suffix = theme.TertiaryColor.Render(fmt.Sprintf("(%d branches)", len(s.AllBranches())))
			}

			fmt.Printf("%s %s\n", name, suffix)
		}
		benchmarkPoint("listCmd", "done")

		return nil
	},
}
