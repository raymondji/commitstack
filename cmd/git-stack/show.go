package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/commitstack/commitstack"
	"github.com/raymondji/commitstack/concurrent"
	"github.com/raymondji/commitstack/githost"
	"github.com/spf13/cobra"
)

var showPRsFlag bool

func init() {
	showCmd.Flags().BoolVar(&showPRsFlag, "prs", false, "Whether to show PRs for each branch")
	showCmd.Flags().BoolVar(&showPRsFlag, "mrs", false, "Whether to show MRs for each branch. Alias for --prs")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show information about the current stack",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, host, theme := deps.git, deps.repoCfg.DefaultBranch, deps.host, deps.theme

		stacks, err := commitstack.ComputeAll(git, defaultBranch)
		if err != nil {
			return err
		}

		var stack commitstack.Stack
		if len(args) == 0 {
			stack, err = stacks.GetCurrent()
			if errors.Is(err, commitstack.ErrNotInAStack) {
				fmt.Println("Not in a stack")
				printProblems(stacks)
				return nil
			} else if err != nil {
				printProblems(stacks)
				return err
			}
		} else {
			wantStack := args[0]
			var found bool
			for _, s := range stacks.Entries {
				if s.Name() == wantStack {
					stack = s
					found = true
				}
			}
			if !found {
				return fmt.Errorf("no stack named: %s", wantStack)
			}
		}

		ctx := context.Background()
		prsBySrcBranch := map[string]githost.PullRequest{}
		if showPRsFlag {
			var actionErr error
			action := func() {
				prs, err := concurrent.Map(ctx, stack.LocalBranches(), func(ctx context.Context, branch commitstack.Branch) (githost.PullRequest, error) {
					pr, err := host.GetPullRequest(deps.remote.RepoPath, branch.Name)
					if errors.Is(err, githost.ErrDoesNotExist) {
						return githost.PullRequest{}, nil
					} else if err != nil {
						return githost.PullRequest{}, err
					}
					return pr, nil
				})
				if err != nil {
					actionErr = err
					return
				}
				for _, pr := range prs {
					if pr.SourceBranch == "" {
						continue
					}
					prsBySrcBranch[pr.SourceBranch] = pr
				}
			}

			err := spinner.New().Title("Fetching MRs...").Action(action).Run()
			if err != nil {
				return err
			}
			if actionErr != nil {
				return actionErr
			}
		}

		if len(args) == 0 {
			fmt.Printf("On stack %s\n", stack.Name())
		}
		fmt.Println("Branches in stack:")
		for i, b := range stack.LocalBranches() {
			var name, suffix string
			if i == 0 {
				suffix = "(top)"
			}
			if b.Current {
				name = "* " + theme.PrimaryColor.Render(b.Name)
			} else {
				name = "  " + b.Name
			}

			fmt.Printf("%s %s\n", name, suffix)
			if showPRsFlag {
				if pr, ok := prsBySrcBranch[b.Name]; ok {
					fmt.Printf("  └── %s\n", pr.WebURL)
					fmt.Println()
				} else {
					fmt.Println()
				}
			}
		}

		printProblem(stack)
		return nil
	},
}

func printProblem(stack commitstack.Stack) {
	if stack.Error != nil {
		fmt.Println()
		fmt.Println("Problems detected:")
		fmt.Printf("  %s\n", stack.Error.Error())
	}
}
