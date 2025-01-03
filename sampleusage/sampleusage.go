package sampleusage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/raymondji/git-stack-cli/config"
	"github.com/raymondji/git-stack-cli/exec"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/libgit"
)

func Basics(git libgit.Git, host githost.Host, defaultBranch string, theme config.Theme) Sample {
	segments := parseLines(
		multiline(
			"Welcome to git stack!",
			"Here is a quick tutorial on how to use the CLI.",
		),
		"Let's start things off on the default branch:",
		shellCmd(fmt.Sprintf("git checkout %s", defaultBranch)),
		"Next, let's create our first branch:",
		shellCmd("git checkout -b myfirststack"),
		shellCmd("echo 'hello world' > myfirststack.txt"),
		shellCmd("git add ."),
		shellCmd("git commit -m 'hello world'"),
		"Now let's stack a second branch on top of our first:",
		shellCmd("git checkout -b myfirststack-pt2"),
		shellCmd("echo 'have a break' >> myfirststack.txt"),
		shellCmd("git commit -am 'break'"),
		shellCmd("echo 'have a kitkat' >> myfirststack.txt"),
		shellCmd("git commit -am 'kitkat'"),
		multiline(
			"So far we've only used standard Git commands. Let's see what git stack can do for us already.",
			"",
			"Our current stack has two branches in it, which we can see with:",
		),
		shellCmd(`git stack show`),
		"Our current stack has 3 commits in it, which we can see with:",
		shellCmd(`git stack log`),
		multiline(
			"We can easily push all branches in the stack up as separate PRs.",
			"git stack automatically sets the target branches for you.",
		),
		shellCmd(`git stack push`),
		"We can quickly view the PRs in the stack using:",
		shellCmd(`git stack show --prs`),
		multiline(
			"To sync the latest changes from the default branch into the stack, you can run:",
			fmt.Sprintf("git rebase %s --update-refs", defaultBranch),
			"Or to avoid having to remember --update-refs, you can do:",
		),
		shellCmd(fmt.Sprintf(`git stack rebase %s --no-edit`, defaultBranch)),
		multiline(
			"Great, we've got the basics down for one stack. How do we deal with multiple stacks?",
			"Let's head back to our default branch and create a second stack.",
		),
		shellCmd(fmt.Sprintf("git checkout %s", defaultBranch)),
		shellCmd("git checkout -b mysecondstack"),
		shellCmd("echo 'buy one get one free' > mysecondstack.txt"),
		shellCmd("git add ."),
		shellCmd("git commit -m 'My second stack'"),
		"To view all the stacks:",
		shellCmd("git stack list"),
		multiline(
			"Nice! All done chapter 1 of the tutorial.",
			"",
			"In chapter 2 we'll see how to make changes to earlier branches in the stack.",
			"Once you're ready, continue the tutorial using:",
			"git stack learn --chapter 2",
			"",
			"To cleanup all the branches/PRs that were created, run:",
			"git stack learn --chapter 1 --cleanup",
		),
	)
	branchesToCleanup := []string{
		"myfirststack", "myfirststack-pt2", "mysecondstack",
	}
	return newSample(git, host, segments, branchesToCleanup, theme, defaultBranch)
}

func Advanced(git libgit.Git, host githost.Host, defaultBranch string, theme config.Theme) Sample {
	segments := parseLines(
		"Coming soon!",
	)
	branchesToCleanup := []string{}
	return newSample(git, host, segments, branchesToCleanup, theme, defaultBranch)
}

type Sample struct {
	defaultBranch     string
	segments          []segment
	theme             config.Theme
	git               libgit.Git
	host              githost.Host
	branchesToCleanup []string
}

func newSample(
	git libgit.Git,
	host githost.Host,
	segments []segment,
	branchesToCleanup []string,
	theme config.Theme,
	defaultBranch string,
) Sample {
	return Sample{
		git:               git,
		host:              host,
		segments:          segments,
		branchesToCleanup: branchesToCleanup,
		theme:             theme,
		defaultBranch:     defaultBranch,
	}
}

func (s Sample) Execute() error {
	if ok, err := s.git.IsRepoClean(); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborting, git repo has changes")
	}

	for _, seg := range s.segments {
		if err := seg.Execute(s.theme); err != nil {
			return err
		}
	}
	return nil
}

func (s Sample) Cleanup() error {
	if ok, err := s.git.IsRepoClean(); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborting, git repo has changes")
	}

	remote, err := s.git.GetRemote()
	if err != nil {
		return err
	}

	if err := s.git.Checkout(s.defaultBranch); err != nil {
		return err
	}

	return s.cleanupBranches(remote.URLPath, s.branchesToCleanup...)
}

func (s Sample) cleanupBranches(repoPath string, names ...string) error {
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

func (s Sample) String() string {
	var sb strings.Builder
	for _, seg := range s.segments {
		sb.Write([]byte(seg.String(s.theme) + "\n"))
	}
	return sb.String()
}

type segment interface {
	String(theme config.Theme) string
	Execute(theme config.Theme) error
}

func parseLines(segments ...any) []segment {
	var out []segment
	for _, seg := range segments {
		switch t := seg.(type) {
		case string:
			out = append(out, text(t))
		case shellCmdLine:
			out = append(out, t)
		default:
			panic(fmt.Sprintf("invalid segment %v", seg))
		}
	}

	return out
}

type textLine string

func text(s string) textLine {
	return textLine(s)
}

func (t textLine) String(theme config.Theme) string {
	if string(t) == "" {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("61")).
		Padding(1).
		Width(50)

	return fmt.Sprint(style.Render(string(t)))
}

func multiline(lines ...string) string {
	return strings.Join(lines, "\n")
}

func (t textLine) Execute(theme config.Theme) error {
	fmt.Println(t.String(theme))
	return nil
}

type shellCmdLine struct {
	text string
}

func shellCmd(s string) segment {
	return shellCmdLine{
		text: s,
	}
}

func (s shellCmdLine) String(theme config.Theme) string {
	return theme.TertiaryColor.Render("> " + s.text)
}

func (s shellCmdLine) Execute(theme config.Theme) error {
	_, err := exec.Run("echo", exec.WithArgs(theme.TertiaryColor.Render("> "+s.text)), exec.WithOSStdout())
	if err != nil {
		return err
	}

	_, err = exec.Run("bash", exec.WithArgs("-c", s.text), exec.WithOSStdout())
	return err
}
