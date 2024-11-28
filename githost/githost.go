package githost

import "errors"

// PullRequest is the most popular name so using this here, but this represents an
// umbrella including pull requests (github), merge requests (gitlab), diffs (phabricator), etc.
type PullRequest struct {
	Title          string
	Description    string
	SourceBranch   string
	TargetBranch   string
	WebURL         string
	MarkdownWebURL string
}

var ErrDoesNotExist = errors.New("does not exist")

type GitHost interface {
	// Returns ErrDoesNotExist if no pull request exists for the given sourceBranch
	GetPullRequest(sourceBranch string) (PullRequest, error)
	UpdatePullRequest(r PullRequest) (PullRequest, error)
	CreatePullRequest(r PullRequest) (PullRequest, error)
}
