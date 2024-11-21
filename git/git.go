package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Git struct{}

// GetRootDir returns the root directory of the current Git repository.
func (g Git) GetRootDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Git root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (g Git) GetCurrentBranch() (string, error) {
	return "", nil
}

func (g Git) GetHash(branch string) (string, error) {
	return "", nil
}

func (g Git) GetLocalBranches() ([]string, error) {
	out, err := runCommand("git", "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	return lines, nil
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

// GetLog retrieves the git log and parses branches for each commit
func (g Git) GetLog(rangeStart, rangeEnd string) (Log, error) {
	out, err := runCommand(
		"git", "log", `--pretty=format:"%h-----%p-----%D"`,
		"--decorate=full", fmt.Sprintf("%s..%s", rangeStart, rangeEnd))
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
		return "", fmt.Errorf("error running cmd %s: %s, err: %v", name, strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}
