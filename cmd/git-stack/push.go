package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/raymondji/git-stack-cli/slices"
	"github.com/raymondji/git-stack-cli/stackparser"
	"github.com/raymondji/git-stack-cli/ui"
	"github.com/spf13/cobra"
)

var pushSaferForceFlag bool
var pushForceFlag bool
var pushCreatePRsFlag bool

func init() {
	pushCmd.Flags().BoolVarP(&pushSaferForceFlag, "safer-force", "f", false, "see git push --force-with-lease and --force-if-includes")
	pushCmd.Flags().BoolVarP(&pushForceFlag, "force", "F", false, "see git push --force")
	pushCmd.Flags().BoolVarP(&pushCreatePRsFlag, "create-prs", "c", false, "Don't create new pull requests. Existing ones are always updated.")
}

var pushCmd = &cobra.Command{
	Use:     "push",
	Aliases: []string{"p"},
	Short:   "Push all branches in the current stack and create/update pull requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, host := deps.git, deps.repoCfg.DefaultBranch, deps.host

		ctx := context.Background()
		log, err := git.LogAll(defaultBranch)
		if err != nil {
			return err
		}
		stacks, err := stackparser.ParseStacks(log)
		if err != nil {
			return err
		}
		currCommit, err := git.GetShortCommitHash("HEAD")
		if err != nil {
			return err
		}
		currBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}
		s, err := stackparser.GetCurrent(stacks, currCommit)
		if err != nil {
			return err
		}

		wantTargets := map[string]string{}
		branches, err := s.TotalOrderedBranches()
		if err != nil {
			return err
		}
		if len(s.DivergesFrom()) > 0 {
			fmt.Println("error: cannot push divergent stacks")
			printProblems([]stackparser.Stack{s}, deps.theme)
			return nil
		}
		for i, b := range branches {
			if i == len(branches)-1 {
				wantTargets[b] = defaultBranch
			} else {
				wantTargets[b] = branches[i+1]
			}
		}

		pushStack := func() ([]githost.PullRequest, error) {
			// Before pushing branches, reset the target branch on any existing MRs if they don't match what we want.
			// If mergeRequestA is from branchA -> branchB, and the branches have been re-ordered to branchB -> branchA,
			// Gitlab will automatically mark mergeRequestA as merged after we push branchA and branchB.
			// We don't want this behaviour.
			prs, err := concurrent.Map(ctx, branches, func(ctx context.Context, branch string) (githost.PullRequest, error) {
				pr, err := host.GetChangeReqeuest(deps.remote.URLPath, branch)
				if errors.Is(err, githost.ErrDoesNotExist) {
					return githost.PullRequest{}, nil
				} else if err != nil {
					return githost.PullRequest{}, err
				}

				if pr.TargetBranch != wantTargets[branch] {
					return host.UpdateChangeRequest(deps.remote.URLPath, githost.PullRequest{
						ID:           pr.ID,
						Title:        pr.Title,
						Description:  pr.Description,
						SourceBranch: branch,
						TargetBranch: defaultBranch,
					})
				}

				return pr, nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to reset target branches on existing MRs, errors: %v", err)
			}
			prs = slices.Filter(prs, func(pr githost.PullRequest) bool {
				return pr.ID != 0
			})
			prsBySourceBranch := slices.ToMap(prs, func(pr githost.PullRequest) string {
				return pr.SourceBranch
			})

			// Push all branches.
			err = concurrent.ForEach(ctx, branches, func(ctx context.Context, branch string) error {
				_, err := git.Push(branch, libgit.PushOpts{
					Force:           pushForceFlag,
					ForceWithLease:  pushSaferForceFlag,
					ForceIfIncludes: pushSaferForceFlag,
				})
				return err
			})
			if err != nil {
				return nil, fmt.Errorf("failed to force push branches, errors: %v", err.Error())
			}

			// Create any new PRs
			if pushCreatePRsFlag {
				prs, err = concurrent.Map(
					ctx,
					branches,
					func(ctx context.Context, branch string) (githost.PullRequest, error) {
						if pr, ok := prsBySourceBranch[branch]; ok {
							return pr, nil
						}

						return host.CreateChangeRequest(deps.remote.URLPath, githost.PullRequest{
							Title:        branch,
							SourceBranch: branch,
							TargetBranch: wantTargets[branch],
						})
					})
				if err != nil {
					return nil, fmt.Errorf("failed to create new PRs, errors: %v", err.Error())
				}
			}

			// Update PRs with correct target branches and stack info.
			return concurrent.Map(ctx, prs, func(ctx context.Context, pr githost.PullRequest) (githost.PullRequest, error) {
				desc := formatPullRequestDescription(pr, prs)
				pr, err := host.UpdateChangeRequest(deps.remote.URLPath, githost.PullRequest{
					ID:           pr.ID,
					Title:        pr.Title,
					Description:  desc,
					SourceBranch: pr.SourceBranch,
					TargetBranch: wantTargets[pr.SourceBranch],
				})
				return pr, err
			})
		}

		var prs []githost.PullRequest
		var actionErr error
		action := func() {
			prs, actionErr = pushStack()
		}
		if err = spinner.New().Title("Pushing stack...").Action(action).Run(); err != nil {
			return err
		}
		if actionErr != nil {
			return actionErr
		}
		prsBySourceBranch := map[string]githost.PullRequest{}
		for _, pr := range prs {
			prsBySourceBranch[pr.SourceBranch] = pr
		}

		fmt.Println("Pushed branches:")
		ui.PrintBranchesInStack(
			branches,
			true,
			currBranch,
			deps.theme,
			prsBySourceBranch,
			true,
			host.GetVocabulary(),
		)
		return nil
	},
}

func formatPullRequestDescription(
	currPR githost.PullRequest, prs []githost.PullRequest,
) string {
	var newStackDesc string
	if len(prs) == 1 {
		// (raymond):
	} else {
		var newStackDescParts []string
		currIndex := slices.IndexFunc(prs, func(pr githost.PullRequest) bool {
			return pr.SourceBranch == currPR.SourceBranch
		})

		for i, pr := range prs {
			var prefix string
			if i == currIndex {
				prefix = "**Current**: "
			} else if i == currIndex-1 {
				prefix = "**Next**: "
			} else if i == currIndex+1 && i == 1 {
				prefix = "**Previous**: "
			}
			newStackDescParts = append(newStackDescParts, fmt.Sprintf("- %s%s", prefix, pr.MarkdownWebURL))
		}

		newStackDesc = "This is **part of a stack**:\n" + strings.Join(newStackDescParts, "\n")
	}

	beginMarker := "<!-- DO NOT EDIT: generated by git stack push (start)-->"
	endMarker := "<!-- DO NOT EDIT: generated by git stack push (end) -->"
	newSection := fmt.Sprintf("%s\n------\n%s\n%s", beginMarker, newStackDesc, endMarker)
	sectionPattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker))

	if sectionPattern.MatchString(currPR.Description) {
		return sectionPattern.ReplaceAllString(currPR.Description, newSection)
	} else {
		return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(currPR.Description), newSection)
	}
}
