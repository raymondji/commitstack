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
		var inference commitstack.InferenceResult
		err = concurrent.Run(
			context.Background(),
			func(ctx context.Context) error {
				var err error
				currCommit, err = git.GetShortCommitHash("HEAD")
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
				inference, err = commitstack.InferStacks(git, log)
				return err
			},
		)
		if err != nil {
			return err
		}
		benchmarkPoint("listCmd", "got curr commit, curr branch, and stack inference")
		defer func() {
			printProblems(inference)
		}()

		var stack commitstack.Stack
		if len(args) == 0 {
			stack, err = commitstack.GetCurrent(inference.InferredStacks, currCommit)
			if err != nil {
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
		benchmarkPoint("listCmd", "got desired stack")

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
			benchmarkPoint("listCmd", "fetched pull requests")
		}

		if len(args) == 0 {
			fmt.Printf("In stack %s\n", stack.Name())
		}

		if showLogFlag {
			fmt.Println("Commits in stack:")
			for i, c := range stack.Commits {
				var hereMarker, topMarker string
				if i == 0 {
					topMarker = fmt.Sprintf(" (%s)", theme.TertiaryColor.Render("top"))
				} else {
					topMarker = strings.Repeat(" ", 6)
				}
				if currCommit == c.Hash {
					hereMarker = "* "
				} else {
					hereMarker = "  "
				}

				var branchCol string
				if len(c.LocalBranches) > 0 {
					var branchParts []string
					for _, b := range c.LocalBranches {
						if b == currBranch {
							branchParts = append(branchParts, fmt.Sprintf("%s %s", theme.PrimaryColor.Render("HEAD ->"), theme.SecondaryColor.Render(b)))
						} else {
							branchParts = append(branchParts, theme.SecondaryColor.Render(b))
						}
					}
					branchCol = fmt.Sprintf("(%s) ", strings.Join(branchParts, ", "))
				} else if currCommit == c.Hash {
					branchCol = fmt.Sprintf("(%s) ", theme.PrimaryColor.Render("HEAD"))
				}

				commitHash := theme.QuaternaryColor.Render(c.Hash)
				fmt.Printf("%s%s %s%s%s\n", hereMarker, commitHash, branchCol, c.Subject, topMarker)
			}
			benchmarkPoint("listCmd", "done printing log")
		} else {
			fmt.Println("Branches in stack:")
			for i, c := range stack.Commits {
				var prefix, branchesSegment, suffix string
				if len(c.LocalBranches) == 0 && c.Hash == currCommit {
					fmt.Println("* " + theme.PrimaryColor.Render(fmt.Sprintf("(HEAD detached at %s)", c.Hash)))
					continue
				} else if len(c.LocalBranches) == 0 {
					continue
				} else {
					var names []string
					for _, b := range c.LocalBranches {
						if b == currBranch {
							names = append(names, theme.PrimaryColor.Render(b))
						} else {
							names = append(names, b)
						}
					}
					branchesSegment = strings.Join(names, ", ")
				}
				if i == 0 {
					suffix = fmt.Sprintf(" (%s)", theme.TertiaryColor.Render("top"))
				}
				if slices.Contains(c.LocalBranches, currBranch) {
					prefix = "*"
				} else {
					prefix = " "
				}

				fmt.Printf("%s %s%s\n", prefix, branchesSegment, suffix)
				if showPRsFlag {
					for _, b := range c.LocalBranches {
						if pr, ok := prsBySrcBranch[b]; ok {
							fmt.Printf("  └── %s\n", pr.WebURL)
						} else {
							fmt.Printf("  └── Not created yet\n")
						}
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
