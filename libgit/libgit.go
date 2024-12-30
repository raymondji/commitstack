package libgit

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/raymondji/git-stack/exec"
)

type Git struct{}

type Upstream struct {
	Remote     string
	BranchName string
}

func (g Git) GetUpstream(branch string) (Upstream, error) {
	output, err := exec.Run(
		"git",
		exec.WithArgs(
			"for-each-ref", "--format=%(upstream:short)", fmt.Sprintf("refs/heads/%s", branch),
		),
	)
	if err != nil {
		return Upstream{}, fmt.Errorf("failed to get upstream, err: %v", err)
	}

	lines := output.Lines()
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

func (g Git) GetRootDir() (string, error) {
	// Use the helper function to run the git command
	output, err := exec.Run("git", exec.WithArgs("rev-parse", "--show-toplevel"))
	if err != nil {
		return "", fmt.Errorf("failed to get Git root dir, err: %v", err)
	}
	return output.Stdout, nil
}

func (g Git) CommitFixup(commitHash string, add bool) (string, error) {
	args := []string{"commit", "--fixup", commitHash}
	if add {
		args = append(args, "-a")
	}
	output, err := exec.Run("git", exec.WithArgs(args...))
	if err != nil {
		return "", fmt.Errorf("failed to git commit --fixup, err: %v", err)
	}
	return output.Stdout, nil
}

func (g Git) CommitEmpty(msg string) error {
	_, err := exec.Run("git", exec.WithArgs("commit", "--allow-empty", "-m", msg))
	if err != nil {
		return fmt.Errorf("failed to commit, err: %v", err)
	}
	return nil
}

func (g Git) GetCurrentBranch() (string, error) {
	output, err := exec.Run("git", exec.WithArgs("rev-parse", "--abbrev-ref", "HEAD"))
	if err != nil {
		return "", fmt.Errorf("failed to get current branch, err: %v", err)
	}
	return output.Stdout, nil
}

func (g Git) GetCommitHash(branch string) (string, error) {
	output, err := exec.Run("git", exec.WithArgs("rev-parse", branch))
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash for branch %s, err: %v", branch, err)
	}
	return output.Stdout, nil
}

func (g *Git) PushForceWithLease(branchName string) (string, error) {
	res, err := exec.Run("git", exec.WithArgs("push", "--force-with-lease", "origin", branchName))
	if err != nil {
		return "", fmt.Errorf("failed to force push branch %s: %w", branchName, err)
	}

	return fmt.Sprintf("Force pushing branch %s\n%s", branchName, res), nil
}

func (g *Git) Fetch(repo string, refspec string) error {
	_, err := exec.Run("git", exec.WithArgs("fetch", repo, refspec))
	if err != nil {
		return fmt.Errorf("failed to fetch, err: %v", err)
	}

	return nil
}

type RebaseOpts struct {
	Interactive    bool
	Env            []string
	AdditionalArgs []string
}

func (g *Git) Rebase(branch string, opts RebaseOpts) (string, error) {
	args := []string{"rebase", branch, "--update-refs"}
	args = append(args, opts.AdditionalArgs...)
	if opts.Interactive {
		args = append(args, "-i")
	}
	output, err := exec.Run(
		"git",
		exec.WithArgs(args...),
		exec.WithEnv(opts.Env...),
		exec.WithInteractive(opts.Interactive))
	if err != nil {
		return "", fmt.Errorf("failed to rebase, err: %v", err)
	}
	return output.Stdout, nil
}

func (g Git) CreateBranch(name string) error {
	_, err := exec.Run("git", exec.WithArgs("checkout", "-b", name))
	if err != nil {
		return fmt.Errorf("failed to create branch, err: %v", err)
	}
	return nil
}

func (g Git) Checkout(name string) error {
	_, err := exec.Run("git", exec.WithArgs("checkout", name))
	if err != nil {
		return fmt.Errorf("failed to checkout branch, err: %v", err)
	}
	return nil
}

type Log struct {
	// Commits are ordered from oldest to newest
	Commits []Commit
}

type Commit struct {
	Date          string
	Subject       string
	Author        string
	Hash          string
	ParentHashes  []string
	LocalBranches []string
}

func (g Git) LogAll(notReachableFrom string) (Log, error) {
	output, err := exec.Run(
		"git",
		exec.WithArgs(
			"log",
			`--pretty=format:%h-----%p-----%D-----%an-----%ar-----%s`,
			"--decorate=full",
			"--branches", fmt.Sprintf("^%s", notReachableFrom),
		),
	)
	if err != nil {
		return Log{}, fmt.Errorf("failed to retrieve git log: %v", err)
	}
	lines := output.Lines()
	slog.Debug("git.LogAll", "output", lines, "len output", len(lines))

	var commits []Commit
	for _, line := range lines {
		parts := strings.Split(line, "-----")
		if len(parts) != 6 {
			return Log{}, fmt.Errorf("unexpected git log line: %s", line)
		}
		commit := Commit{
			Hash:         parts[0],
			ParentHashes: strings.Fields(parts[1]),
			Author:       parts[3],
			Date:         parts[4],
			Subject:      parts[5],
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

	slog.Debug("git.LogAll", "commits", commits)
	return Log{
		Commits: commits,
	}, nil
}
