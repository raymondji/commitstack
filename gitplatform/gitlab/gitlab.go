package gitlab

import (
	"encoding/json"
	"fmt"

	"strings"

	"github.com/raymondji/git-stacked/exec"
	"github.com/raymondji/git-stacked/gitplatform"
)

type Gitlab struct{}

var _ gitplatform.GitPlatform = Gitlab{}

type GitlabMR struct {
	Title        string `json:"title"`
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
	Description  string `json:"description"`
	WebURL       string `json:"web_url"`
	IID          int    `json:"iid"`
}

func (g Gitlab) GetPullRequest(sourceBranch string) (gitplatform.PullRequest, error) {
	output, err := exec.Command(
		"glab", "mr", "view", sourceBranch, "--output=json",
	)
	if err != nil {
		if strings.Contains(err.Error(), "no open merge request available") {
			return gitplatform.PullRequest{}, gitplatform.ErrDoesNotExist
		}
		return gitplatform.PullRequest{}, fmt.Errorf(
			"error checking MRs for branch %s, output: %s, err: %v", sourceBranch, output, err)
	}

	var mr GitlabMR
	if err := json.Unmarshal([]byte(output), &mr); err != nil {
		return gitplatform.PullRequest{}, err
	}

	return gitplatform.PullRequest{
		SourceBranch:   mr.SourceBranch,
		TargetBranch:   mr.TargetBranch,
		Description:    mr.Description,
		WebURL:         mr.WebURL,
		MarkdownWebURL: fmt.Sprintf("%s+", mr.WebURL),
		Title:          mr.Title,
	}, nil
}

func (g Gitlab) UpdatePullRequest(pr gitplatform.PullRequest) (gitplatform.PullRequest, error) {
	_, err := exec.Command("glab", "mr", "update", pr.SourceBranch, "--target-branch", pr.TargetBranch, "--description", pr.Description)
	if err != nil {
		return gitplatform.PullRequest{}, fmt.Errorf("error updating pr: %s, err: %v", pr.SourceBranch, err)
	}

	return g.GetPullRequest(pr.SourceBranch)
}

func (g Gitlab) CreatePullRequest(pr gitplatform.PullRequest) (gitplatform.PullRequest, error) {
	_, err := exec.Command(
		"glab", "mr", "create",
		"--source-branch", pr.SourceBranch,
		"--target-branch", pr.TargetBranch,
		"--title", pr.SourceBranch,
		"--description", pr.Description,
		"--draft",
	)
	if err != nil {
		return gitplatform.PullRequest{}, fmt.Errorf("error creating merge request: %+v, err: %v", pr, err)
	}

	return g.GetPullRequest(pr.SourceBranch)
}
