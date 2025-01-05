package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/inference"
	"github.com/spf13/cobra"
)

var showPRsFlag bool
var showLogFlag bool

func init() {
	showCmd.Flags().BoolVar(&showPRsFlag, "prs", false, "Whether to show PRs for each branch")
	showCmd.Flags().BoolVar(&showPRsFlag, "mrs", false, "Whether to show MRs for each branch. Alias for --prs")
	showCmd.Flags().BoolVarP(&showLogFlag, "log", "l", false, "Whether to log all commits in the stack")
}

var showCmd = &cobra.Command{
	Use:     "show [stack]",
	Aliases: []string{"sh"},
	Short:   "Show information about the current or specified stack",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if showPRsFlag && showLogFlag {
			return fmt.Errorf("--log and --prs are incompatible")
		}

		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, host, theme := deps.git, deps.repoCfg.DefaultBranch, deps.host, deps.theme
		benchmarkPoint("listCmd", "got deps")

		var currBranch, currCommit string
		var stacks []inference.Stack
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
				stacks, err = inference.InferStacks(log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "got curr commit, curr branch, and stack inference")

		var stack inference.Stack
		if len(args) == 0 {
			if slices.Contains(mergedBranches, currBranch) {
				fmt.Printf("error: the current branch is not a valid stack (it's merged into %s)\n", defaultBranch)
				return nil
			}
			stack, err = inference.GetCurrent(stacks, currCommit)
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
			printProblems([]inference.Stack{stack})
		}()
		benchmarkPoint("listCmd", "got desired stack")

		if len(args) == 0 {
			fmt.Printf("In stack %s\n", stack.Name)
		}
		totalOrder := true
		branches, err := stack.TotalOrderedBranches()
		var errNoTotalOrder inference.NoTotalOrderError
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
		if showPRsFlag {
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

		if showLogFlag {
			return errors.New("unimplemented")
			// fmt.Println("Commits in stack:")
			// for i, c := range stack.Commits {
			// 	var hereMarker, topMarker string
			// 	if i == 0 {
			// 		topMarker = fmt.Sprintf(" (%s)", theme.TertiaryColor.Render("top"))
			// 	} else {
			// 		topMarker = strings.Repeat(" ", 6)
			// 	}
			// 	if currCommit == c.Hash {
			// 		hereMarker = "* "
			// 	} else {
			// 		hereMarker = "  "
			// 	}

			// 	var branchCol string
			// 	if len(c.LocalBranches) > 0 {
			// 		var branchParts []string
			// 		const headMarker = "HEAD ->"
			// 		for _, b := range c.LocalBranches {
			// 			if b == currBranch {
			// 				branchParts = append(
			// 					[]string{
			// 						fmt.Sprintf("%s %s", theme.PrimaryColor.Render(headMarker), theme.SecondaryColor.Render(b)),
			// 					},
			// 					branchParts...,
			// 				)
			// 			} else {
			// 				branchParts = append(branchParts, theme.SecondaryColor.Render(b))
			// 			}
			// 		}
			// 		branchCol = fmt.Sprintf("(%s) ", strings.Join(branchParts, ", "))
			// 	} else if currCommit == c.Hash {
			// 		branchCol = fmt.Sprintf("(%s) ", theme.PrimaryColor.Render("HEAD"))
			// 	}

			// 	commitHash := theme.QuaternaryColor.Render(c.Hash)
			// 	fmt.Printf("%s%s %s%s%s\n", hereMarker, commitHash, branchCol, c.Subject, topMarker)
			// }
			// benchmarkPoint("listCmd", "done printing log")
		} else {
			fmt.Println("Branches in stack:")
			for i, branch := range branches {
				// if len(c.LocalBranches) == 0 && c.Hash == currCommit {
				// 	fmt.Println("* " + theme.PrimaryColor.Render(fmt.Sprintf("(HEAD detached at %s)", c.Hash)))
				// 	continue
				// } else if len(c.LocalBranches) == 0 {
				// 	continue
				// }
				var prefix, branchesSegment, suffix string
				if i == 0 && totalOrder {
					suffix = fmt.Sprintf(" (%s)", theme.TertiaryColor.Render("top"))
				}
				if branch == currBranch {
					prefix = "*"
					branchesSegment = theme.PrimaryColor.Render(branch)
				} else {
					prefix = " "
					branchesSegment = branch
				}

				fmt.Printf("%s %s%s\n", prefix, branchesSegment, suffix)
				if showPRsFlag {
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
		}

		return nil
	},
}
