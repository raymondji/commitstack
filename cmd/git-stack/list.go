package main

import (
	"context"
	"fmt"

	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/inference"
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

		var currCommit string
		var mergedBranches []string
		var stacks []inference.Stack
		err = concurrent.Run(
			context.Background(),
			func(ctx context.Context) error {
				var err error
				currCommit, err = git.GetShortCommitHash("HEAD")
				return err
			},
			func(ctx context.Context) error {
				var err error
				mergedBranches, err = git.GetMergedBranches(defaultBranch)
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
		benchmarkPoint("listCmd", "got curr commit and stack inference")
		defer func() {
			printMergedBranches(mergedBranches, defaultBranch, theme)
		}()
		defer func() {
			printProblems(stacks, theme)
		}()

		for _, s := range stacks {
			var name, suffix string
			if s.IsCurrent(currCommit) {
				name = "* " + theme.PrimaryColor.Render(s.Name)
			} else {
				name = "  " + s.Name
			}

			all := s.Branches()
			if len(all) == 1 {
				suffix = theme.TertiaryColor.Render("(1 branch)")
			} else {
				suffix = theme.TertiaryColor.Render(fmt.Sprintf("(%d branches)", len(all)))
			}

			fmt.Printf("%s %s\n", name, suffix)
		}
		benchmarkPoint("listCmd", "done")

		return nil
	},
}
