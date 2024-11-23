package gitplatform

import "errors"

// PullRequest is the most popular name so using this here, but this represents an
// umbrella including pull requests (github), merge requests (gitlab), diffs (phabricator), etc.
type PullRequest struct {
	Description  string
	SourceBranch string
	TargetBranch string
	WebURL       string
}

var ErrDoesNotExist = errors.New("does not exist")

type GitPlatform interface {
	// Returns ErrDoesNotExist if no pull request exists for the given sourceBranch
	GetPullRequest(sourceBranch string) (PullRequest, error)
	UpdatePullRequest(r PullRequest) (PullRequest, error)
	CreatePullRequest(r PullRequest) (PullRequest, error)
}
