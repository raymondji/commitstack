package gitlab

import (
	"fmt"

	"github.com/raymondji/commitstack/githost/internal"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type gitlabClient struct {
	client *gitlab.Client
}

func New(personalAccessToken string) (internal.Host, error) {
	client, err := gitlab.NewClient(personalAccessToken)
	if err != nil {
		return gitlabClient{}, fmt.Errorf("Failed to create client: %v", err)
	}
	return gitlabClient{
		client: client,
	}, nil
}

// e.g. for https://gitlab.com/raymondji/git-stacked-gitlab-test, the path is raymondji/git-stacked-gitlab-test
func (g gitlabClient) GetRepo(repoPath string) (internal.Repo, error) {
	project, _, err := g.client.Projects.GetProject(repoPath, &gitlab.GetProjectOptions{})
	if err != nil {
		return internal.Repo{}, err
	}

	return internal.Repo{
		DefaultBranch: project.DefaultBranch,
	}, nil
}

func (g gitlabClient) GetPullRequest(repoPath string, sourceBranch string) (internal.PullRequest, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.Ptr("opened"),
		SourceBranch: &sourceBranch,
	}
	mergeRequests, _, err := g.client.MergeRequests.ListProjectMergeRequests(repoPath, opts)
	if err != nil {
		return internal.PullRequest{}, fmt.Errorf("failed to list merge requests: %w", err)
	}
	switch len(mergeRequests) {
	case 0:
		return internal.PullRequest{}, fmt.Errorf("%w, source branch: %s", internal.ErrDoesNotExist, sourceBranch)
	case 1:
		mr := mergeRequests[0]
		return convertMR(mr), nil
	default:
		return internal.PullRequest{}, fmt.Errorf("found multiple merge requests for source branch: %s", sourceBranch)
	}
}

func (g gitlabClient) CreatePullRequest(repoPath string, pr internal.PullRequest) (internal.PullRequest, error) {
	if pr.Title == "" {
		return internal.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	// TODO: make it a draft
	opts := &gitlab.CreateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		SourceBranch: &pr.SourceBranch,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.CreateMergeRequest(repoPath, opts)
	if err != nil {
		return internal.PullRequest{}, fmt.Errorf("failed to create merge request: %w", err)
	}

	return convertMR(mr), nil
}

func (g gitlabClient) UpdatePullRequest(repoPath string, pr internal.PullRequest) (internal.PullRequest, error) {
	if pr.ID == 0 {
		return internal.PullRequest{}, fmt.Errorf("pull request ID must be set")
	}
	if pr.Title == "" {
		return internal.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	opts := &gitlab.UpdateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.UpdateMergeRequest(repoPath, pr.ID, opts)
	if err != nil {
		return internal.PullRequest{}, fmt.Errorf("failed to update merge request: %w, mr: %+v", err, pr)
	}

	return convertMR(mr), nil
}

func (g gitlabClient) ClosePullRequest(repoPath string, pr internal.PullRequest) (internal.PullRequest, error) {
	if pr.ID == 0 {
		return internal.PullRequest{}, fmt.Errorf("pull request ID must be set")
	}
	if pr.Title == "" {
		return internal.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	opts := &gitlab.UpdateMergeRequestOptions{
		StateEvent: gitlab.Ptr("closed"),
	}

	mr, _, err := g.client.MergeRequests.UpdateMergeRequest(repoPath, pr.ID, opts)
	if err != nil {
		return internal.PullRequest{}, fmt.Errorf("failed to update merge request: %w, mr: %+v", err, pr)
	}

	return convertMR(mr), nil
}

func convertMR(mr *gitlab.MergeRequest) internal.PullRequest {
	return internal.PullRequest{
		ID:             mr.IID,
		SourceBranch:   mr.SourceBranch,
		TargetBranch:   mr.TargetBranch,
		Description:    mr.Description,
		WebURL:         mr.WebURL,
		MarkdownWebURL: fmt.Sprintf("%s+", mr.WebURL),
		Title:          mr.Title,
	}
}
