package internal

import "errors"

// Using ChangeRequest as a generic term to represent pull requests (github),
// merge requests (gitlab), diffs (phabricator), etc.
type ChangeRequest struct {
	ID             int
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

type Vocabulary struct {
	ChangeRequestNameCapitalized string
	ChangeRequestName            string
	ChangeRequestNamePlural      string
	ChangeRequestNameShort       string
	ChangeRequestNameShortPlural string
}

type Host interface {
	GetVocabulary() Vocabulary
	GetRepo(repoPath string) (Repo, error)
	// Returns ErrDoesNotExist if no change request exists for the given sourceBranch
	GetChangeReqeuest(repoPath string, sourceBranch string) (ChangeRequest, error)
	UpdateChangeRequest(repoPath string, r ChangeRequest) (ChangeRequest, error)
	CreateChangeRequest(repoPath string, r ChangeRequest) (ChangeRequest, error)
	CloseChangeRequest(repoPath string, r ChangeRequest) (ChangeRequest, error)
}
