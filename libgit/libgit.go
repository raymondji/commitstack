package libgit

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/raymondji/git-stack-cli/exec"
	"github.com/raymondji/git-stack-cli/githost"
)

// When --update-refs was introduced
var minGitVersion = version{major: 2, minor: 38}

type Git interface {
	ValidateGitInstall() error
	IsRepoClean() (bool, error)
	GetRemote() (Remote, error)
	GetRootDir() (string, error)
	CommitFixup(commitHash string, add bool) (string, error)
	CommitEmpty(msg string) error
	GetMergedBranches(ref string) ([]string, error)
	GetCurrentBranch() (string, error)
	GetShortCommitHash(branch string) (string, error)
	Push(branchName string, opts PushOpts) (string, error)
	Rebase(branch string, opts RebaseOpts) (string, error)
	CreateBranch(name string, startPoint string) error
	DeleteBranchIfExists(name string) error
	DeleteRemoteBranchIfExists(name string) error
	Checkout(name string) error
	LogAll(notReachableFrom string) (Log, error)
	LogOneline(from string, to string) error
}

type git struct{}

func New() Git {
	return git{}
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

func (g git) ValidateGitInstall() error {
	ok, err := exec.InPath("git")
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("git is not installed")
	}

	v, err := g.getVersion()
	if err != nil {
		return fmt.Errorf("failed to get git version")
	}
	if v.lessThan(minGitVersion) {
		return fmt.Errorf("the minimum supported git version is %s, yours is %s", minGitVersion, v)
	}
	return nil
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

func (g git) IsRepoClean() (bool, error) {
	output, err := exec.Run("git", exec.WithArgs("status", "--porcelain"))
	if err != nil {
		return false, fmt.Errorf("failed to run git status: %v", err)
	}

	return len(output.Stdout) == 0, nil
}

type Remote struct {
	Kind    githost.Kind
	URLPath string // e.g. raymondji/git-stack-cli
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
	case strings.Contains(output.Stdout, "github.com"):
		kind = githost.Github
	default:
		return Remote{}, fmt.Errorf("unsupported git host: %s", output.Stdout)
	}

	// Extract the repository name
	path, err := parseRepoPathFromRemoteURL(output.Stdout)
	if err != nil {
		return Remote{}, err
	}
	return Remote{
		URLPath: path,
		Kind:    kind,
	}, nil
}

func parseRepoPathFromRemoteURL(url string) (string, error) {
	u, err := giturls.Parse(url)
	if err != nil {
		return "", fmt.Errorf("failed to parse origin url %q, err: %v", url, err)
	}

	path := strings.TrimSuffix(u.Path, ".git")
	path = strings.TrimPrefix(path, "/")
	return path, nil
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
	args := []string{"commit", "-m", fmt.Sprintf("fixup! %s", commitHash)}
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

func (g git) Commit(msg string) error {
	_, err := exec.Run("git", exec.WithArgs("commit", "-a", "-m", msg))
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

func (g git) GetShortCommitHash(branch string) (string, error) {
	output, err := exec.Run("git", exec.WithArgs("rev-parse", "--short", branch))
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash for branch %s, err: %v", branch, err)
	}
	return output.Stdout, nil
}

type PushOpts struct {
	Force           bool
	ForceWithLease  bool
	ForceIfIncludes bool
}

func (g git) Push(branchName string, opts PushOpts) (string, error) {
	args := []string{"push", "origin", branchName}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.ForceWithLease {
		args = append(args, "--force-with-lease")
	}
	if opts.ForceIfIncludes {
		args = append(args, "--force-if-includes")
	}

	output, err := exec.Run("git", exec.WithArgs(args...))
	if err != nil {
		return "", fmt.Errorf("failed to push branch, args: %v, %w", args, err)
	}

	return output.Stdout, nil
}

type RebaseOpts struct {
	Interactive bool
	Autosquash  bool
	KeepBase    bool
	UpdateRefs  bool
}

func (g git) Rebase(branch string, opts RebaseOpts) (string, error) {
	env := []string{}
	args := []string{"rebase", branch}
	if opts.KeepBase {
		args = append(args, "--keep-base")
	}
	if opts.UpdateRefs {
		args = append(args, "--update-refs")
	}
	if opts.Interactive {
		args = append(args, "-i")
	}
	if opts.Autosquash {
		if !opts.Interactive {
			// Hack(raymond): --autosquash only works with interactive rebase, so use
			// GIT_SEQUENCE_EDITOR=true to accept the changes automatically.
			env = append(env, "GIT_SEQUENCE_EDITOR=true")
			args = append(args, "-i")
		}
		args = append(args, "--autosquash")
	}

	output, err := exec.Run(
		"git",
		exec.WithArgs(args...),
		exec.WithEnv(env...),
		exec.WithInteractive(opts.Interactive))
	if err != nil {
		return "", fmt.Errorf("failed to rebase, err: %v", err)
	}
	return output.Stdout, nil
}

func (g git) GetMergedBranches(ref string) ([]string, error) {
	output, err := exec.Run("git", exec.WithArgs("branch", "--merged", ref, "--format=%(refname:short)"))
	if err != nil {
		return nil, fmt.Errorf("failed to find branches merged into %s, err: %v", ref, err)
	}

	var branches []string
	for _, line := range output.Lines() {
		if line == "" {
			continue
		}
		branches = append(branches, line)
	}

	return branches, nil
}

func (g git) CreateBranch(name string, startPoint string) error {
	_, err := exec.Run("git", exec.WithArgs("branch", name, startPoint))
	if err != nil {
		return fmt.Errorf("failed to create branch, err: %v", err)
	}
	return nil
}

func (g git) DeleteBranchIfExists(name string) error {
	_, err := exec.Run("git", exec.WithArgs("branch", "-D", name))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete branch, err: %v", err)
	}
	return nil
}

func (g git) DeleteRemoteBranchIfExists(name string) error {
	_, err := exec.Run("git", exec.WithArgs("push", "origin", "--delete", name))
	if err != nil {
		if strings.Contains(err.Error(), "remote ref does not exist") {
			return nil
		}
		return fmt.Errorf("failed to delete remote branch, err: %v", err)
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

func (g git) LogOneline(from string, to string) error {
	_, err := exec.Run(
		"git",
		exec.WithArgs(
			"log", "--oneline", fmt.Sprintf("%s..%s", from, to),
		),
		exec.WithOSStdout(),
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve git log: %v", err)
	}
	return nil
}

// Is there any advantage to using git rev-list --parents --branches instead?
// Seems to be about the same, git git rev-list would need to do a separate
// git branch call to map branch refs to commit hashes
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
