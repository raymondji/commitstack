package gitlab

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/githost/internal"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type gitlabClient struct {
	client *gitlab.Client
}

func New(personalAccessToken string) (internal.Host, error) {
	client, err := gitlab.NewClient(personalAccessToken)
	if err != nil {
		return gitlabClient{}, fmt.Errorf("failed to create client: %v", err)
	}
	return gitlabClient{
		client: client,
	}, nil
}

func (g gitlabClient) GetVocabulary() internal.Vocabulary {
	return internal.Vocabulary{
		ChangeRequestNameCapitalized: "Merge request",
		ChangeRequestName:            "merge request",
		ChangeRequestNamePlural:      "merge requests",
		ChangeRequestNameShort:       "mr",
		ChangeRequestNameShortPlural: "mrs",
	}
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

func (g gitlabClient) GetChangeReqeuest(repoPath string, sourceBranch string) (internal.ChangeRequest, error) {
	opts := &gitlab.ListProjectMergeRequestsOptions{
		State:        gitlab.Ptr("opened"),
		SourceBranch: &sourceBranch,
	}
	mergeRequests, _, err := g.client.MergeRequests.ListProjectMergeRequests(repoPath, opts)
	if err != nil {
		return internal.ChangeRequest{}, fmt.Errorf("failed to list merge requests: %w", err)
	}
	switch len(mergeRequests) {
	case 0:
		return internal.ChangeRequest{}, fmt.Errorf("%w, source branch: %s", internal.ErrDoesNotExist, sourceBranch)
	case 1:
		mr := mergeRequests[0]
		return convertMR(mr), nil
	default:
		return internal.ChangeRequest{}, fmt.Errorf("found multiple merge requests for source branch: %s", sourceBranch)
	}
}

func (g gitlabClient) CreateChangeRequest(repoPath string, pr internal.ChangeRequest) (internal.ChangeRequest, error) {
	if pr.Title == "" {
		return internal.ChangeRequest{}, fmt.Errorf("merge request title cannot be empty")
	}

	opts := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.Ptr(fmt.Sprintf("Draft: %s", pr.Title)),
		Description:  &pr.Description,
		SourceBranch: &pr.SourceBranch,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.CreateMergeRequest(repoPath, opts)
	if err != nil {
		return internal.ChangeRequest{}, fmt.Errorf("failed to create merge request: %w", err)
	}

	return convertMR(mr), nil
}

func (g gitlabClient) UpdateChangeRequest(repoPath string, pr internal.ChangeRequest) (internal.ChangeRequest, error) {
	if pr.ID == 0 {
		return internal.ChangeRequest{}, fmt.Errorf("merge request ID must be set")
	}
	if pr.Title == "" {
		return internal.ChangeRequest{}, fmt.Errorf("merge request title cannot be empty")
	}

	opts := &gitlab.UpdateMergeRequestOptions{
		Title:        &pr.Title,
		Description:  &pr.Description,
		TargetBranch: &pr.TargetBranch,
	}

	mr, _, err := g.client.MergeRequests.UpdateMergeRequest(repoPath, pr.ID, opts)
	if err != nil {
		return internal.ChangeRequest{}, fmt.Errorf("failed to update merge request: %w, mr: %+v", err, pr)
	}

	return convertMR(mr), nil
}

func (g gitlabClient) CloseChangeRequest(repoPath string, pr internal.ChangeRequest) (internal.ChangeRequest, error) {
	if pr.ID == 0 {
		return internal.ChangeRequest{}, fmt.Errorf("merge request ID must be set")
	}
	if pr.Title == "" {
		return internal.ChangeRequest{}, fmt.Errorf("merge request title cannot be empty")
	}

	opts := &gitlab.UpdateMergeRequestOptions{
		StateEvent: gitlab.Ptr("close"),
	}

	mr, _, err := g.client.MergeRequests.UpdateMergeRequest(repoPath, pr.ID, opts)
	if err != nil {
		return internal.ChangeRequest{}, fmt.Errorf("failed to update merge request: %w, mr: %+v", err, pr)
	}

	return convertMR(mr), nil
}

func convertMR(mr *gitlab.MergeRequest) internal.ChangeRequest {
	return internal.ChangeRequest{
		ID:             mr.IID,
		SourceBranch:   mr.SourceBranch,
		TargetBranch:   mr.TargetBranch,
		Description:    mr.Description,
		WebURL:         mr.WebURL,
		MarkdownWebURL: fmt.Sprintf("%s+", mr.WebURL),
		Title:          mr.Title,
	}
}
