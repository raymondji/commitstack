package gitlib

import (
	"fmt"
	"strings"

	"github.com/raymondji/git-stacked/exec"
)

const DEBUG = false

type Git struct{}

type Upstream struct {
	Remote     string
	BranchName string
}

func (g Git) GetUpstream(branch string) (Upstream, error) {
	output, err := exec.Command("git", "for-each-ref", "--format=%(upstream:short)", fmt.Sprintf("refs/heads/%s", branch))
	if err != nil {
		return Upstream{}, fmt.Errorf("failed to get upstream, err: %v", err)
	}

	lines := strings.Split(output, "\n")
	switch len(lines) {
	case 0:
		return Upstream{}, fmt.Errorf("branch %s does not have an upstream branch", branch)
	case 1:
		parts := strings.Split(lines[0], "/")
		if len(parts) != 2 {
			return Upstream{}, fmt.Errorf("failed to get remote for branch %s, unexpected output: %v", branch, output)
		}
		return Upstream{
			Remote:     parts[0],
			BranchName: parts[1],
		}, nil
	default:
		return Upstream{}, fmt.Errorf("branch %s matched multiple upstream branches: %v", branch, lines)
	}
}

// GetRootDir returns the root directory of the current Git repository.
func (g Git) GetRootDir() (string, error) {
	// Use the helper function to run the git command
	output, err := exec.Command("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get Git root dir, err: %v", err)
	}
	return output, nil
}

func (g Git) GetCurrentBranch() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch, err: %v", err)
	}
	return output, nil
}

func (g *Git) PushForceWithLease(branchName string) (string, error) {
	res, err := exec.Command("git", "push", "--force-with-lease", "origin", branchName)
	if err != nil {
		return "", fmt.Errorf("failed to force push branch %s: %w", branchName, err)
	}

	return fmt.Sprintf("Force pushing branch %s\n%s", branchName, res), nil
}

func (g *Git) Fetch(repo string, refspec string) error {
	_, err := exec.Command("git", "fetch", repo, refspec)
	if err != nil {
		return fmt.Errorf("failed to fetch, err: %v", err)
	}

	return nil
}

func (g *Git) Rebase(branch string) (string, error) {
	res, err := exec.Command("git", "rebase", "--update-refs", branch)
	if err != nil {
		return "", fmt.Errorf("failed to rebase, err: %v", err)
	}
	return res, nil
}

func (g *Git) RebaseInteractiveKeepBase(branch string) error {
	err := exec.InteractiveCommand("git", "rebase", "--update-refs", "--keep-base", "-i", branch)
	if err != nil {
		return fmt.Errorf("failed to rebase, err: %v", err)
	}
	return nil
}

func (g Git) CreateBranch(name string) error {
	_, err := exec.Command("git", "checkout", "-b", name)
	if err != nil {
		return fmt.Errorf("failed to create branch, err: %v", err)
	}
	return nil
}

func (g Git) CommitEmpty(msg string) error {
	_, err := exec.Command("git", "commit", "--allow-empty", "-m", msg)
	if err != nil {
		return fmt.Errorf("failed to commit, err: %v", err)
	}
	return nil
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
	out, err := exec.Command(
		"git", "log", `--pretty=format:%h-----%p-----%D`, "--decorate=full",
		"--branches", fmt.Sprintf("^%s", notReachableFrom))
	if err != nil {
		return Log{}, fmt.Errorf("failed to retrieve git log: %v", err)
	}
	if DEBUG {
		fmt.Println("log output")
		fmt.Println(out)
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
