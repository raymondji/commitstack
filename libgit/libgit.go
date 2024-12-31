package libgit

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/raymondji/git-stack/exec"
	"github.com/raymondji/git-stack/githost"
)

// When --update-refs was introduced
var minGitVersion = version{major: 2, minor: 38}

type Git interface {
	GetRemote() (Remote, error)
	GetUpstream(branch string) (Upstream, error)
	GetRootDir() (string, error)
	CommitFixup(commitHash string, add bool) (string, error)
	CommitEmpty(msg string) error
	GetCurrentBranch() (string, error)
	GetCommitHash(branch string) (string, error)
	PushForceWithLease(branchName string) (string, error)
	Fetch(repo string, refspec string) error
	Rebase(branch string, opts RebaseOpts) (string, error)
	CreateBranch(name string) error
	Checkout(name string) error
	LogAll(notReachableFrom string) (Log, error)
}

type git struct{}

func New() (Git, error) {
	ok, err := exec.InPath("git")
	if err != nil {
		return git{}, err
	}
	if !ok {
		return git{}, fmt.Errorf("git is not installed")
	}

	g := git{}
	v, err := g.getVersion()
	if err != nil {
		return git{}, fmt.Errorf("failed to get git version")
	}
	if v.lessThan(minGitVersion) {
		return git{}, fmt.Errorf("the minimum supported git version is %s, yours is %s", minGitVersion, v)
	}
	return g, nil
}

type Upstream struct {
	Remote     string
	BranchName string
}

type version struct {
	major int
	minor int
}

func (v version) lessThan(other version) bool {
	return v.major < other.major || (v.major == other.major && v.minor < other.minor)
}

func (v version) String() string {
	return fmt.Sprintf("%d.%d.0", v.major, v.minor)
}

func (g git) getVersion() (version, error) {
	output, err := exec.Run(
		"git", exec.WithArgs("-v"),
	)
	if err != nil {
		return version{}, fmt.Errorf("failed to get upstream, err: %v", err)
	}

	fields := strings.Fields(output.Stdout)
	if len(fields) < 3 {
		return version{}, fmt.Errorf("unexpected git -v output: %v", output.Stdout)
	}

	vParts := strings.Split(fields[2], ".")
	if len(vParts) != 3 {
		return version{}, fmt.Errorf("unexpected git version string: %v", fields[2])
	}
	major, err := strconv.Atoi(vParts[0])
	if err != nil {
		return version{}, err
	}
	minor, err := strconv.Atoi(vParts[1])
	if err != nil {
		return version{}, err
	}

	return version{
		major: major,
		minor: minor,
	}, nil
}

type Remote struct {
	Kind     githost.Kind
	RepoPath string
}

func (g git) GetRemote() (Remote, error) {
	output, err := exec.Run(
		"git",
		exec.WithArgs(
			"remote", "get-url", "origin",
		),
	)
	if err != nil {
		return Remote{}, fmt.Errorf("failed to get upstream, err: %v", err)
	}

	var kind githost.Kind
	switch {
	case strings.Contains(output.Stdout, "gitlab.com"):
		kind = githost.Gitlab
	default:
		return Remote{}, errors.New(fmt.Sprintf("unsupported git host: %s", output.Stdout))
	}

	// Extract the repository name
	path, err := parseRepoPathFromRemoteURL(output.Stdout)
	if err != nil {
		return Remote{}, err
	}
	return Remote{
		RepoPath: path,
		Kind:     kind,
	}, nil
}

func parseRepoPathFromRemoteURL(url string) (string, error) {
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to parse repo path from url: %s", url)
	}

	return parts[1], nil
}

func (g git) GetUpstream(branch string) (Upstream, error) {
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

func (g git) GetRootDir() (string, error) {
	// Use the helper function to run the git command
	output, err := exec.Run("git", exec.WithArgs("rev-parse", "--show-toplevel"))
	if err != nil {
		return "", fmt.Errorf("failed to get Git root dir, err: %v", err)
	}
	return output.Stdout, nil
}

func (g git) CommitFixup(commitHash string, add bool) (string, error) {
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

func (g git) CommitEmpty(msg string) error {
	_, err := exec.Run("git", exec.WithArgs("commit", "--allow-empty", "-m", msg))
	if err != nil {
		return fmt.Errorf("failed to commit, err: %v", err)
	}
	return nil
}

func (g git) GetCurrentBranch() (string, error) {
	output, err := exec.Run("git", exec.WithArgs("rev-parse", "--abbrev-ref", "HEAD"))
	if err != nil {
		return "", fmt.Errorf("failed to get current branch, err: %v", err)
	}
	return output.Stdout, nil
}

func (g git) GetCommitHash(branch string) (string, error) {
	output, err := exec.Run("git", exec.WithArgs("rev-parse", branch))
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash for branch %s, err: %v", branch, err)
	}
	return output.Stdout, nil
}

func (g git) PushForceWithLease(branchName string) (string, error) {
	res, err := exec.Run("git", exec.WithArgs("push", "--force-with-lease", "origin", branchName))
	if err != nil {
		return "", fmt.Errorf("failed to force push branch %s: %w", branchName, err)
	}

	return fmt.Sprintf("Force pushing branch %s\n%s", branchName, res), nil
}

func (g git) Fetch(repo string, refspec string) error {
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

func (g git) Rebase(branch string, opts RebaseOpts) (string, error) {
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

func (g git) CreateBranch(name string) error {
	_, err := exec.Run("git", exec.WithArgs("checkout", "-b", name))
	if err != nil {
		return fmt.Errorf("failed to create branch, err: %v", err)
	}
	return nil
}

func (g git) Checkout(name string) error {
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

func (g git) LogAll(notReachableFrom string) (Log, error) {
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
