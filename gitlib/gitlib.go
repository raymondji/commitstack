package gitlib

import (
	"fmt"
	"os/exec"
	"strings"
)

type Git struct{}

// GetRootDir returns the root directory of the current Git repository.
func (g Git) GetRootDir() (string, error) {
	// Use the helper function to run the git command
	output, err := runCommand("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get Git root dir, err: %v", err)
	}
	return output, nil
}

func (g Git) GetCurrentBranch() (string, error) {
	output, err := runCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch, err: %v", err)
	}
	return output, nil
}

type Log struct {
	// Commits are ordered from oldest to newest
	Commits []Commit
}

type Commit struct {
	Hash          string
	ParentHashes  []string
	LocalBranches []string
}

func (g Git) LogAll(notReachableFrom string) (Log, error) {
	out, err := runCommand(
		"git", "log", `--pretty=format:%h-----%p-----%D`, "--decorate=full",
		"--branches", fmt.Sprintf("^%s", notReachableFrom),
		"--not", "--remotes")
	if err != nil {
		return Log{}, fmt.Errorf("failed to retrieve git log: %v", err)
	}

	lines := strings.Split(out, "\n")
	var commits []Commit
	for _, line := range lines {
		parts := strings.SplitN(line, "-----", 3)
		commit := Commit{
			Hash:         parts[0],
			ParentHashes: strings.Fields(parts[1]),
		}
		branchParts := strings.Fields(parts[2])
		for _, part := range branchParts {
			part = strings.TrimSpace(part)
			part = strings.TrimSuffix(part, "HEAD ->")
			part = strings.TrimSuffix(part, ",")

			if strings.HasPrefix(part, "refs/heads/") {
				branch := strings.TrimPrefix(part, "refs/heads/")
				commit.LocalBranches = append(commit.LocalBranches, branch)
			}
		}

		commits = append(commits, commit)
	}

	return Log{
		Commits: commits,
	}, nil
}

func runCommand(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", fmt.Errorf("error running cmd %s %s, err: %v", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}
