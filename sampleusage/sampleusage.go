package sampleusage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	return s.cleanupBranches(remote.URLPath, "learncommitstack", "learncommitstack-pt2")
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
	segments := parseLines(
		"Welcome to commitstack!\nHere is a quick tutorial on how to use the CLI.",
		"First, let's start on the default branch:",
		shellCmd(fmt.Sprintf("git checkout %s", s.defaultBranch)),
		"Next, let's create our first branch:",
		shellCmd("git checkout -b learncommitstack"),
		shellCmd("echo 'hello world' > learncommitstack.txt"),
		shellCmd("git add ."),
		shellCmd("git commit -m 'hello world'"),
		"Now let's stack a second branch on top of our first:",
		shellCmd("git checkout -b learncommitstack-pt2"),
		shellCmd("echo 'have a break' >> learncommitstack.txt"),
		shellCmd("git commit -am 'break'"),
		shellCmd("echo 'have a kitkat' >> learncommitstack.txt"),
		shellCmd("git commit -am 'kitkat'"),
		"So far everything we've done has been normal Git. Let's see what commitstack can do for us already.",
		"Our current stack has two branches in it, which we can see with:",
		shellCmd(`git stack show`),
		"Our current stack has 3 commits in it, which we can see with:",
		shellCmd(`git stack log`),
		"We can easily push all branches in the stack up as separate PRs.\ncommitstack automatically sets the target branches for you on the PRs.",
		shellCmd(`git stack push`),
		"We can quickly view the PRs in the stack using:",
		shellCmd(`git stack show --prs`),
		"Nice! All done part 1 of the tutorial. In part 2 we'll learn how to make more changes to a stack.",
		"Once you're ready, continue the tutorial using:",
		shellCmd("git stack learn --part 2"),
	)
	return newSample(s.git, segments, s.theme)
}

type Sample struct {
	segments []segment
	theme    config.Theme
	git      libgit.Git
}

func newSample(git libgit.Git, segments []segment, theme config.Theme) Sample {
	return Sample{
		git:      git,
		segments: segments,
		theme:    theme,
	}
}

func (s Sample) Execute() error {
	if ok, err := s.git.IsRepoClean(); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborting, git repo has changes")
	}

	for _, l := range s.segments {
		if err := l.Execute(s.theme); err != nil {
			return err
		}
	}
	return nil
}

func (s Sample) String() string {
	var sb strings.Builder
	for _, l := range s.segments {
		sb.Write([]byte(l.String(s.theme) + "\n"))
	}
	return sb.String()
}

type segment interface {
	String(theme config.Theme) string
	Execute(theme config.Theme) error
}

func parseLines(segments ...any) []segment {
	var out []segment
	for _, l := range segments {
		switch t := l.(type) {
		case string:
			out = append(out, text(t))
		case shellCmdLine:
			out = append(out, t)
		default:
			panic(fmt.Sprintf("invalid segment %v", l))
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
