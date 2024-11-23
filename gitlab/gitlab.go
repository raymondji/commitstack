package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"github.com/raymondji/git-stacked/exec"
)

type Gitlab struct{}

var ErrNoMRForBranch = errors.New("no MR exists for the branch")

type GitlabMR struct {
	TargetBranch string `json:"target_branch"`
}

// GetMRTargetBranch returns ErrNoMRForBranch if no MR exists for the branch
func (g Gitlab) GetMRTargetBranch(branchName string) (string, error) {
	output, err := exec.Command(
		"glab", "mr", "view", branchName, "--output=json",
	)
	if err != nil {
		if strings.Contains(err.Error(), "no open merge request available") {
			return "", ErrNoMRForBranch
		}
		return "", fmt.Errorf("error checking MRs for branch %s, output: %s, err: %v", branchName, output, err)
	}

	var mr GitlabMR
	if err := json.Unmarshal([]byte(output), &mr); err != nil {
		return "", err
	}
	return mr.TargetBranch, nil
}

func (g Gitlab) SetMRTargetBranch(sourceBranch string, targetBranch string) (string, error) {
	output, err := exec.Command("glab", "mr", "update", sourceBranch, "--target-branch", targetBranch)
	if err != nil {
		return "", fmt.Errorf("error setting target branch for MR from %s to %s: %v", sourceBranch, targetBranch, err)
	}

	return output, nil
}

func (g Gitlab) CreateMR(sourceBranch string, targetBranch string) (string, error) {
	output, err := exec.Command(
		"glab", "mr", "create",
		"--source-branch", sourceBranch,
		"--target-branch", targetBranch,
		"--title", sourceBranch,
		"--draft", "--description", "Sample description",
	)
	if err != nil {
		return "", fmt.Errorf("error creating merge request from %s to %s: %v", sourceBranch, targetBranch, err)
	}

	return output, nil
}
