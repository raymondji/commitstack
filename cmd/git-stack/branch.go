package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/stackparser"
	"github.com/spf13/cobra"
)

var branchPRsFlag bool

func init() {
	branchCmd.Flags().BoolVar(&branchPRsFlag, "prs", false, "Whether to show PRs for each branch")
	branchCmd.Flags().BoolVar(&branchPRsFlag, "mrs", false, "Whether to show MRs for each branch. Alias for --prs")
}

var branchCmd = &cobra.Command{
	Use:     "branch [stack]",
	Aliases: []string{"b"},
	Short:   "List branches in the stack",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// if branchPRsFlag && branchLogFlag {
		// 	return fmt.Errorf("--log and --prs are incompatible")
		// }

		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, host, theme := deps.git, deps.repoCfg.DefaultBranch, deps.host, deps.theme
		benchmarkPoint("listCmd", "got deps")

		var currBranch, currCommit string
		var stacks []stackparser.Stack
		var mergedBranches []string
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
				var err error
				currBranch, err = git.GetCurrentBranch()
				return err
			},
			func(ctx context.Context) error {
				log, err := git.LogAll(defaultBranch)
				if err != nil {
					return err
				}
				stacks, err = stackparser.InferStacks(log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "got curr commit, curr branch, and stack stackparser")

		var stack stackparser.Stack
		if len(args) == 0 {
			if slices.Contains(mergedBranches, currBranch) {
				fmt.Printf("error: the current branch is not a valid stack (it's merged into %s)\n", defaultBranch)
				return nil
			}
			stack, err = stackparser.GetCurrent(stacks, currCommit)
			if err != nil {
				return err
			}
		} else {
			wantStack := args[0]
			var found bool
			for _, s := range stacks {
				if s.Name == wantStack {
					stack = s
					found = true
				}
			}
			if !found {
				return fmt.Errorf("no stack named: %s", wantStack)
			}
		}
		defer func() {
			printProblems([]stackparser.Stack{stack}, deps.theme)
		}()
		benchmarkPoint("listCmd", "got desired stack")

		totalOrder := true
		branches, err := stack.TotalOrderedBranches()
		var errNoTotalOrder stackparser.NoTotalOrderError
		if errors.As(err, &errNoTotalOrder) {
			// TODO: check for this specific error type
			fmt.Printf("Warning: stack %s does not have a total order\n", stack.Name)
			fmt.Println("Branches are displayed in reverse lexicographic order.")
			fmt.Println()

			branches = stack.Branches()
			totalOrder = false
		} else if err != nil {
			return err
		}
		ctx := context.Background()
		prsBySrcBranch := map[string]githost.PullRequest{}
		if branchPRsFlag {
			var actionErr error
			action := func() {
				prs, err := concurrent.Map(ctx, branches, func(ctx context.Context, branch string) (githost.PullRequest, error) {
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
			benchmarkPoint("listCmd", "fetched pull requests")
		}

		// TODO: if no heremarker, e.g. if I'm on another branch,
		// render all branch names without indentation. Looks less ugly
		for i, branch := range branches {
			// if len(c.LocalBranches) == 0 && c.Hash == currCommit {
			// 	fmt.Println("* " + theme.PrimaryColor.Render(fmt.Sprintf("(HEAD detached at %s)", c.Hash)))
			// 	continue
			// } else if len(c.LocalBranches) == 0 {
			// 	continue
			// }
			var hereMarker, branchesSegment, suffix string
			if totalOrder && i == 0 {
				suffix = fmt.Sprintf(" (%s)", theme.TertiaryColor.Render("top"))
			}
			if branch == currBranch {
				hereMarker = "*"
				branchesSegment = theme.PrimaryColor.Render(branch)
			} else {
				hereMarker = " "
				if totalOrder && i == 0 && branch != currBranch {
					branchesSegment = theme.TertiaryColor.Render(branch)
				} else {
					branchesSegment = branch
				}
			}

			fmt.Printf("%s %s%s\n", hereMarker, branchesSegment, suffix)
			if branchPRsFlag {
				if pr, ok := prsBySrcBranch[branch]; ok {
					fmt.Printf("  └── %s\n", pr.WebURL)
				} else {
					fmt.Printf("  └── Not created yet\n")
				}

				if i != len(stack.Commits)-1 {
					fmt.Println()
				}
			}
		}
		benchmarkPoint("listCmd", "done printing branches")

		return nil
	},
}
