package sampleusage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/raymondji/commitstack/config"
	"github.com/raymondji/commitstack/exec"
	"github.com/raymondji/commitstack/githost"
	"github.com/raymondji/commitstack/libgit"
)

type Samples struct {
	theme         config.Theme
	defaultBranch string
	host          githost.Host
	git           libgit.Git
}

func New(theme config.Theme, defaultBranch string, git libgit.Git, host githost.Host) Samples {
	return Samples{
		theme:         theme,
		defaultBranch: defaultBranch,
		host:          host,
		git:           git,
	}
}

func (s Samples) Cleanup() error {
	remote, err := s.git.GetRemote()
	if err != nil {
		return err
	}

	if err := s.git.Checkout(s.defaultBranch); err != nil {
		return err
	}

	return s.cleanupBranches(remote.URLPath, "learncommitstack", "learncommitstack2")
}

func (s Samples) cleanupBranches(repoPath string, names ...string) error {
	for _, name := range names {
		hasPR := true
		pr, err := s.host.GetPullRequest(repoPath, name)
		if errors.Is(err, githost.ErrDoesNotExist) {
			hasPR = false
		} else if err != nil {
			return err
		}

		if hasPR {
			_, err = s.host.ClosePullRequest(repoPath, pr)
			if err != nil {
				return err
			}
		}

		if err := s.git.DeleteBranchIfExists(name); err != nil {
			return err
		}
	}

	return nil
}

func (s Samples) Part1() Sample {
	lines := parseLines(
		"Welcome to commitstack!",
		"Here is a quick tutorial on how to use the CLI.",
		"First, let's start on the default branch:",
		shellCmd(fmt.Sprintf("git checkout %s", s.defaultBranch)),
		"",
		"Next, let's create our first branch:",
		shellCmd(
			"git checkout -b learncommitstack && \\\n"+
				"echo 'hello world' > learncommitstack.txt && \\\n"+
				"git add . && \\\n"+
				"git commit -m 'hello world'",
		),
		"",
		"Now let's stack a second branch on top of our first:",
		shellCmd(
			"git checkout -b learncommitstack-pt2 && \\\n"+
				"echo 'have a break' >> learncommitstack.txt && \\\n"+
				"git commit -am 'break' && \\\n"+
				"echo 'have a kitkat' >> learncommitstack.txt && \\\n"+
				"git commit -am 'kitkat'",
		),
		"",
		"So far everything we've done has been normal Git. Let's see what commitstack can do for us!",
		"Our current stack has two branches in it, which we can see with:",
		shellCmd(`git stack show`),
		"Our current stack has 3 commits in it, which we can see with:",
		shellCmd(`git stack log`),
		"",
		"We can easily push all branches in the stack up as separate PRs:",
		"commitstack automatically sets the target branches for you on the PRs.",
		shellCmd(`git stack push`),
		"We can quickly view the PRs in the stack using:",
		shellCmd(`git stack show --prs`),
		"",
		"Nice! All done part 1 of the tutorial. In part 2 we'll learn how to make more changes to a stack.",
		"Once you're ready, continue the tutorial using:",
		shellCmd("git stack learn --part 2"),
	)
	return newSample(lines, s.theme)
}

type Sample struct {
	lines []line
	theme config.Theme
}

func newSample(lines []line, theme config.Theme) Sample {
	return Sample{
		lines: lines,
		theme: theme,
	}
}

func (s Sample) Execute() error {
	for _, l := range s.lines {
		if err := l.RunAsShellCmd(); err != nil {
			return err
		}
	}
	return nil
}

func (s Sample) String() string {
	var sb strings.Builder
	for _, l := range s.lines {
		sb.Write([]byte(l.String() + "\n"))
	}
	return sb.String()
}

type line interface {
	fmt.Stringer
	RunAsShellCmd() error
}

func parseLines(lines ...any) []line {
	var out []line
	for _, l := range lines {
		switch t := l.(type) {
		case string:
			out = append(out, text(t))
		case shellCmdLine:
			out = append(out, t)
		default:
			panic(fmt.Sprintf("invalid line %v", l))
		}
	}

	return out
}

type textLine string

func text(s string) textLine {
	return textLine(s)
}

func (t textLine) String() string {
	return string(t)
}

func (t textLine) RunAsShellCmd() error {
	_, err := exec.Run("echo", exec.WithArgs(strings.Fields(string(t))...))
	return err
}

type shellCmdLine struct {
	cmd  string
	args []string
}

func shellCmd(s string) line {
	if strings.TrimSpace(s) == "" {
		panic(fmt.Sprintf("shellCmd passed invalid string: %q", s))
	}
	return shellCmdLine{
		cmd:  "bash",
		args: []string{"-c", s},
	}
}

func (s shellCmdLine) String() string {
	return fmt.Sprintf("%s %s", s.cmd, strings.Join(s.args, " "))
}

func (s shellCmdLine) RunAsShellCmd() error {
	_, err := exec.Run(s.cmd, exec.WithArgs(s.args...))
	return err
}
