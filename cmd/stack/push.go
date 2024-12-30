package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh/spinner"
	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/concurrent"
	"github.com/raymondji/git-stack/githost"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newPushCmd(git libgit.Git, host githost.Host, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push the stack to the remote and create/update pull requests",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}
			s, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			if s.Error != nil {
				return fmt.Errorf("cannot push when stack has an error: %v", s.Error)
			}

			wantTargets := map[string]string{}
			lb := s.LocalBranches()
			for i, b := range lb {
				if i == len(lb)-1 {
					wantTargets[b.Name] = defaultBranch
				} else {
					wantTargets[b.Name] = lb[i+1].Name
				}
			}

			pushStack := func() ([]githost.PullRequest, error) {
				// Create any missing pull requests.
				// For safety, also reset the target branch on any existing MRs if they don't match.
				// If any branches have been re-ordered, Gitlab can automatically merge MRs, which is not what we want here.
				prs, err := concurrent.Map(ctx, lb, func(ctx context.Context, branch commitstack.Branch) (githost.PullRequest, error) {
					pr, err := host.GetPullRequest(branch.Name)
					if errors.Is(err, githost.ErrDoesNotExist) {
						return host.CreatePullRequest(githost.PullRequest{
							SourceBranch: branch.Name,
							TargetBranch: wantTargets[branch.Name],
							Description:  "",
						})
					} else if err != nil {
						return githost.PullRequest{}, err
					}

					if pr.TargetBranch != wantTargets[branch.Name] {
						return host.UpdatePullRequest(githost.PullRequest{
							SourceBranch: branch.Name,
							TargetBranch: defaultBranch,
							Description:  pr.Description,
						})
					}

					return pr, nil
				})
				if err != nil {
					return nil, fmt.Errorf("failed to force push branches, errors: %v", err)
				}

				// Push all branches.
				localBranches := s.LocalBranches()
				err = concurrent.ForEach(ctx, localBranches, func(ctx context.Context, branch commitstack.Branch) error {
					_, err := git.PushForceWithLease(branch.Name)
					return err
				})
				if err != nil {
					return nil, fmt.Errorf("failed to force push branches, errors: %v", err.Error())
				}

				// Update PRs with correct target branches and stack info.
				return concurrent.Map(ctx, prs, func(ctx context.Context, pr githost.PullRequest) (githost.PullRequest, error) {
					desc := formatPullRequestDescription(pr, prs)
					pr, err := host.UpdatePullRequest(githost.PullRequest{
						SourceBranch: pr.SourceBranch,
						TargetBranch: wantTargets[pr.SourceBranch],
						Description:  desc,
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
}
