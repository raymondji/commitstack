package commitstack

import (
	"fmt"

	"github.com/raymondji/git-stack-cli/libgit"
)

type Git interface {
	GetCurrentBranch() (string, error)
	GetBranchesContainingCommit(commitHash string) ([]string, error)
	LogAll(notReachableFrom string) (libgit.Log, error)
}

type FakeGit struct {
	Log                            libgit.Log
	CurrentBranch                  string
	CommitHashToContainingBranches map[string][]string
}

func (fg *FakeGit) LogAll(notReachableFrom string) (libgit.Log, error) {
	return fg.Log, nil
}

func (fg *FakeGit) GetCurrentBranch() (string, error) {
	return fg.CurrentBranch, nil
}

func (fg *FakeGit) GetBranchesContainingCommit(commit string) ([]string, error) {
	if got, ok := fg.CommitHashToContainingBranches[commit]; !ok {
		return nil, fmt.Errorf("no entry found for commit %s", commit)
	} else {
		return got, nil
	}
}
