package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
	"github.com/raymondji/git-stack/githost"
)

type githubClient struct {
	client *github.Client
}

func New(personalAccessToken string) (githost.Host, error) {
	client := github.NewClient(nil).WithAuthToken(personalAccessToken)
	return &githubClient{
		client: client,
	}, nil
}

func (g *githubClient) GetRepo(repoPath string) (githost.Repo, error) {
	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return githost.Repo{}, err
	}

	repository, _, err := g.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return githost.Repo{}, fmt.Errorf("failed to get repository: %w", err)
	}

	return githost.Repo{
		DefaultBranch: *repository.DefaultBranch,
	}, nil
}

// GetPullRequest retrieves a pull request by its source branch.
func (g *githubClient) GetPullRequest(repoPath string, sourceBranch string) (githost.PullRequest, error) {
	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return githost.PullRequest{}, err
	}

	opts := &github.PullRequestListOptions{
		State: "open",
		Head:  sourceBranch,
	}
	prs, _, err := g.client.PullRequests.List(context.Background(), owner, repo, opts)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to list pull requests: %w", err)
	}

	switch len(prs) {
	case 0:
		return githost.PullRequest{}, fmt.Errorf("%w, source branch: %s", githost.ErrDoesNotExist, sourceBranch)
	case 1:
		pr := prs[0]
		return convertPR(pr), nil
	default:
		return githost.PullRequest{}, fmt.Errorf("found multiple pull requests for source branch: %s", sourceBranch)
	}
}

func (g *githubClient) CreatePullRequest(repoPath string, pr githost.PullRequest) (githost.PullRequest, error) {
	if pr.Title == "" {
		return githost.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return githost.PullRequest{}, err
	}

	newPR := &github.NewPullRequest{
		Title: github.Ptr(pr.Title),
		Head:  github.Ptr(pr.SourceBranch),
		Base:  github.Ptr(pr.TargetBranch),
		Body:  github.Ptr(pr.Description),
		Draft: github.Ptr(true),
	}

	createdPR, _, err := g.client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to create pull request: %w", err)
	}

	return convertPR(createdPR), nil
}

func (g *githubClient) UpdatePullRequest(repoPath string, pr githost.PullRequest) (githost.PullRequest, error) {
	if pr.ID == 0 {
		return githost.PullRequest{}, fmt.Errorf("pull request ID must be set")
	}
	if pr.Title == "" {
		return githost.PullRequest{}, fmt.Errorf("pull request title cannot be empty")
	}

	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return githost.PullRequest{}, err
	}

	updatedPR := &github.PullRequest{
		Title: github.Ptr(pr.Title),
		Body:  github.Ptr(pr.Description),
		Base: &github.PullRequestBranch{
			Ref: github.Ptr(pr.TargetBranch),
		},
	}

	prResult, _, err := g.client.PullRequests.Edit(context.Background(), owner, repo, int(pr.ID), updatedPR)
	if err != nil {
		return githost.PullRequest{}, fmt.Errorf("failed to update pull request: %w", err)
	}

	return convertPR(prResult), nil
}

func convertPR(pr *github.PullRequest) githost.PullRequest {
	return githost.PullRequest{
		ID:             *pr.Number,
		SourceBranch:   *pr.Head.Ref,
		TargetBranch:   *pr.Base.Ref,
		Description:    *pr.Body,
		WebURL:         *pr.HTMLURL,
		MarkdownWebURL: fmt.Sprintf("%s+", *pr.HTMLURL),
		Title:          *pr.Title,
	}
}

// parseRepoPath returns (owner, name, error)
func parseRepoPath(repoPath string) (string, string, error) {
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo path: %s", repoPath)
	}
	return parts[0], parts[1], nil
}
