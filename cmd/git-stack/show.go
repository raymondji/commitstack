package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/githost"
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

		var stack commitstack.Stack
		if len(args) == 0 {
			stack, err = commitstack.GetCurrent(inference.InferredStacks, currBranch)
			if errors.Is(err, commitstack.ErrUnableToInferCurrentStack) {
				fmt.Println(err.Error())
				return nil
			} else if err != nil {
				return err
			}
		} else {
			wantStack := args[0]
			var found bool
			for _, s := range inference.InferredStacks {
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
				prs, err := concurrent.Map(ctx, stack.AllBranches(), func(ctx context.Context, branch string) (githost.PullRequest, error) {
					pr, err := host.GetPullRequest(deps.remote.URLPath, branch)
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
		for i, c := range stack.Commits {
			var prefix, name, suffix string
			switch len(c.LocalBranches) {
			case 0:
				continue
			case 1:
				name = c.LocalBranches[0]
			default:
				name = fmt.Sprintf("%s", strings.Join(c.LocalBranches, ", "))
			}
			if i == 0 {
				suffix = "(top)"
			}
			if slices.Contains(c.LocalBranches, currBranch) {
				prefix = "*"
				name = theme.PrimaryColor.Render(name)
			} else {
				prefix = " "
			}

			fmt.Printf("%s %s %s\n", prefix, name, suffix)
			if showPRsFlag {
				for _, b := range c.LocalBranches {
					if pr, ok := prsBySrcBranch[b]; ok {
						fmt.Printf("  └── %s\n", pr.WebURL)
					}
				}

				fmt.Println()
			}
		}

		return nil
	},
}
