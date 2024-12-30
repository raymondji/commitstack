package githost

import "errors"

// Using PullRequest since it's the most widely used term, but this represents an
// umbrella term including pull requests (github), merge requests (gitlab), diffs (phabricator), etc.
type PullRequest struct {
	Title          string
	Description    string
	SourceBranch   string
	TargetBranch   string
	WebURL         string
	MarkdownWebURL string
}

type Repo struct {
	DefaultBranch string
}

var ErrDoesNotExist = errors.New("does not exist")

type Host interface {
	GetRepo() (Repo, error)
	// Returns ErrDoesNotExist if no pull request exists for the given sourceBranch
	GetPullRequest(sourceBranch string) (PullRequest, error)
	UpdatePullRequest(r PullRequest) (PullRequest, error)
	CreatePullRequest(r PullRequest) (PullRequest, error)
}

type Kind string

const (
	Gitlab Kind = "GITLAB"
)
