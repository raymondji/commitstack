package githost

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/config"
	"github.com/raymondji/git-stack-cli/githost/github"
	"github.com/raymondji/git-stack-cli/githost/gitlab"
	"github.com/raymondji/git-stack-cli/githost/internal"
)

type (
	Host        = internal.Host
	PullRequest = internal.PullRequest
	Repo        = internal.Repo
)

var (
	ErrDoesNotExist = internal.ErrDoesNotExist
)

type Kind string

const (
	Gitlab Kind = "GITLAB"
	Github Kind = "GITHUB"
)

func New(kind Kind, repoCfg config.RepoConfig) (Host, error) {
	switch kind {
	case Gitlab:
		host, err := gitlab.New(repoCfg.Gitlab.PersonalAccessToken)
		if err != nil {
			return host, fmt.Errorf("failed to init gitlab client, err: %v", err)
		}
		return host, nil
	case Github:
		host, err := github.New(repoCfg.Github.PersonalAccessToken)
		if err != nil {
			return host, fmt.Errorf("failed to init github client, err: %v", err)
		}
		return host, nil
	default:
		var host Host
		return host, fmt.Errorf("unsupported git host %s", kind)
	}
}
