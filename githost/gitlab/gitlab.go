package gitlab

import (
	"fmt"

	"github.com/raymondji/git-stack/githost"
	gitlabSDK "gitlab.com/gitlab-org/api/client-go"
)

type gitlab struct {
	client *gitlabSDK.Client
}

func New(personalAccessToken string) (githost.Host, error) {
	client, err := gitlabSDK.NewClient(personalAccessToken)
	if err != nil {
		return gitlab{}, fmt.Errorf("Failed to create client: %v", err)
	}
	return gitlab{
		client: client,
	}, nil
}

// e.g. for https://gitlab.com/raymondji/git-stacked-gitlab-test, the path is raymondji/git-stacked-gitlab-test
func (g gitlab) GetRepo(repoPath string) (githost.Repo, error) {
	project, _, err := g.client.Projects.GetProject(repoPath, &gitlabSDK.GetProjectOptions{})
	if err != nil {
		return githost.Repo{}, err
	}

	return githost.Repo{
		DefaultBranch: project.DefaultBranch,
	}, nil
}

func (g gitlab) GetPullRequest(repoPath string, sourceBranch string) (githost.PullRequest, error) {
	opts := &gitlabSDK.ListProjectMergeRequestsOptions{
		SourceBranch: &sourceBranch,
	}
	mergeRequests, _, err := g.client.MergeRequests.ListProjectMergeRequests(repoPath, opts)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to list merge requests: %w", err)
	}
	switch len(mergeRequests) {
	case 0:
		return githost.PullRequest{}, fmt.Errorf("%w, source branch: %s", githost.ErrDoesNotExist, sourceBranch)
	case 1:
		mr := mergeRequests[0]
		return convertMR(mr), nil
	default:
		return githost.PullRequest{}, fmt.Errorf("found multiple merge requests for source branch: %s", sourceBranch)
	}
}

func (g gitlab) CreatePullRequest(repoPath string, pr githost.PullRequest) (githost.PullRequest, error) {
	if pr.Title == "" {
		return githost.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	// TODO: make it a draft
	opts := &gitlabSDK.CreateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		SourceBranch: &pr.SourceBranch,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.CreateMergeRequest(repoPath, opts)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to create merge request: %w", err)
	}

	return convertMR(mr), nil
}

func (g gitlab) UpdatePullRequest(repoPath string, pr githost.PullRequest) (githost.PullRequest, error) {
	if pr.ID == 0 {
		return githost.PullRequest{}, fmt.Errorf("pull request ID must be set")
	}
	if pr.Title == "" {
		return githost.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	opts := &gitlabSDK.UpdateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.UpdateMergeRequest(repoPath, pr.ID, opts)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to update merge request: %w, mr: %+v", err, pr)
	}

	return convertMR(mr), nil
}

func convertMR(mr *gitlabSDK.MergeRequest) githost.PullRequest {
	return githost.PullRequest{
		ID:             mr.IID,
		SourceBranch:   mr.SourceBranch,
		TargetBranch:   mr.TargetBranch,
		Description:    mr.Description,
		WebURL:         mr.WebURL,
		MarkdownWebURL: fmt.Sprintf("%s+", mr.WebURL),
		Title:          mr.Title,
	}
}
