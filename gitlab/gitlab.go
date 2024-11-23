package gitlab

import (
	"fmt"

	"strings"

	"github.com/raymondji/git-stacked/exec"
)

type Gitlab struct{}

func (g Gitlab) CreateMergeRequest() {

}

func (g Gitlab) HasMR(branchName string) (bool, error) {
	output, err := exec.Command("glab", "mr", "view", branchName, "--output=json")
	if err != nil {
		if strings.Contains(err.Error(), "no open merge request available") {
			return false, nil
		}

		return false, fmt.Errorf("error checking MRs for branch %s, output: %s, err: %v", branchName, output, err)
	}

	return true, nil
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
