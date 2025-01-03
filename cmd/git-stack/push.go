package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/concurrent"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/slices"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all branches in the stack to the remote and create/update pull requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		git, defaultBranch, host := deps.git, deps.repoCfg.DefaultBranch, deps.host

		ctx := context.Background()
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
		s, err := commitstack.GetCurrent(inference.InferredStacks, currBranch)
		if err != nil {
			return err
		}
		if len(s.ValidationErrors) > 0 {
			fmt.Println("cannot push stack, please resolve validation errors")
			printProblems(inference)
			return nil
		}
		defer func() {
			printProblems(inference)
		}()

		wantTargets := map[string]string{}
		branches, err := s.UniqueBranches()
		if err != nil {
			return err
		}
		for i, b := range branches {
			if i == len(branches)-1 {
				wantTargets[b] = defaultBranch
			} else {
				wantTargets[b] = branches[i+1]
			}
		}

		pushStack := func() ([]githost.PullRequest, error) {
			// For safety, reset the target branch on any existing MRs if they don't match.
			// If any branches have been re-ordered, Gitlab can automatically merge MRs, which is not what we want here.
			prs, err := concurrent.Map(ctx, branches, func(ctx context.Context, branch string) (githost.PullRequest, error) {
				pr, err := host.GetPullRequest(deps.remote.URLPath, branch)
				if errors.Is(err, githost.ErrDoesNotExist) {
					return githost.PullRequest{}, nil
				} else if err != nil {
					return githost.PullRequest{}, err
				}

				if pr.TargetBranch != wantTargets[branch] {
					return host.UpdatePullRequest(deps.remote.URLPath, githost.PullRequest{
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
				_, err := git.PushForceWithLease(branch)
				return err
			})
			if err != nil {
				return nil, fmt.Errorf("failed to force push branches, errors: %v", err.Error())
			}

			// Create any new PRs
			prs, err = concurrent.Map(
				ctx,
				branches,
				func(ctx context.Context, branch string) (githost.PullRequest, error) {
					if pr, ok := prsBySourceBranch[branch]; ok {
						return pr, nil
					}

					return host.CreatePullRequest(deps.remote.URLPath, githost.PullRequest{
						Title:        branch,
						SourceBranch: branch,
						TargetBranch: wantTargets[branch],
					})
				})
			if err != nil {
				return nil, fmt.Errorf("failed to create new PRs, errors: %v", err.Error())
			}

			// Update PRs with correct target branches and stack info.
			return concurrent.Map(ctx, prs, func(ctx context.Context, pr githost.PullRequest) (githost.PullRequest, error) {
				desc := formatPullRequestDescription(pr, prs)
				pr, err := host.UpdatePullRequest(deps.remote.URLPath, githost.PullRequest{
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
		for _, pr := range prs {
			fmt.Printf("Pushed %s: %s\n", pr.SourceBranch, pr.WebURL)
		}
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
				prefix = "Current: "
			} else if i == currIndex-1 {
				prefix = "Next: "
			} else if i == currIndex+1 {
				prefix = "Prev: "
			}
			newStackDescParts = append(newStackDescParts, fmt.Sprintf("- %s%s", prefix, pr.MarkdownWebURL))
		}

		newStackDesc = "Merge request stack:\n" + strings.Join(newStackDescParts, "\n")
	}

	beginMarker := "<!-- DO NOT EDIT: generated by git stack push (start)-->"
	endMarker := "<!-- DO NOT EDIT: generated by git stack push (end) -->"
	newSection := fmt.Sprintf("%s\n%s\n%s", beginMarker, newStackDesc, endMarker)
	sectionPattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker))

	if sectionPattern.MatchString(currPR.Description) {
		return sectionPattern.ReplaceAllString(currPR.Description, newSection)
	} else {
		return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(currPR.Description), newSection)
	}
}
