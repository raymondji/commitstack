package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/raymondji/git-stack-cli/config"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/libgit"
	"github.com/raymondji/git-stack-cli/slices"
	"github.com/raymondji/git-stack-cli/stackparser"
	"golang.org/x/exp/maps"
)

type deps struct {
	git     libgit.Git
	host    githost.Host
	repoCfg config.RepoConfig
	theme   config.Theme
	remote  libgit.Remote
}

func initDeps() (deps, error) {
	git := libgit.New()
	benchmarkPoint("initDeps", "done initiating git")

	remote, err := git.GetRemote()
	if err != nil {
		return deps{}, err
	}
	benchmarkPoint("initDeps", "got git remote")

	cfg, err := config.Load()
	if err != nil {
		return deps{}, fmt.Errorf("failed to load config, err: %v", err.Error())
	}
	benchmarkPoint("initDeps", "loaded config")

	repoCfg, ok := cfg.Repositories[remote.URLPath]
	if !ok {
		return deps{}, fmt.Errorf(
			"no config found for the current repo (%s)"+
				", please setup git stack using the `git stack init` command",
			remote.URLPath)
	}

	host, err := githost.New(remote.Kind, repoCfg)
	if err != nil {
		return deps{}, err
	}

	out := deps{
		theme:   config.NewTheme(cfg.Theme),
		git:     git,
		host:    host,
		repoCfg: repoCfg,
		remote:  remote,
	}
	benchmarkPoint("initDeps", "done")
	return out, nil
}

// TODO: try to format this similar to git status.
/*
Unmerged paths:
  (use "git add <file>..." to mark resolution)
        both modified:   src/module1.py
        both added:      src/module2.py
*/
func printProblems(stacks []stackparser.Stack, theme config.Theme) {
	var divergentStacksMsgs []string
	for _, stack := range stacks {
		got := stack.DivergesFrom()
		if len(got) > 0 {
			divergentStacksMsgs = append(divergentStacksMsgs, fmt.Sprintf(
				"%s has diverged from %s",
				stack.Name,
				strings.Join(maps.Keys(got), ", "),
			))
		}
	}
	if len(divergentStacksMsgs) > 0 {
		fmt.Println()
		fmt.Println("Divergent stacks:")
		fmt.Println(strings.Repeat(" ", 2) + `(use "git merge" to merge one branch into another)`)
		fmt.Println(strings.Repeat(" ", 2) + `(use "git stack rebase" to rebase one stack onto another)`)
		for _, msg := range divergentStacksMsgs {
			fmt.Println(strings.Repeat(" ", 8) + theme.QuaternaryColor.Render(msg))
		}
	}

	var noTotalOrderMsgs []string
	for _, stack := range stacks {
		_, err := stack.TotalOrderedBranches()
		if err != nil {
			noTotalOrderMsgs = append(noTotalOrderMsgs, err.Error())
		}
	}
	if len(noTotalOrderMsgs) > 0 {
		fmt.Println()
		fmt.Println("Partially ordered stacks:")
		fmt.Println(strings.Repeat(" ", 2) + `(use "git reset --hard <ref>..." to undo a merge commit)`)
		fmt.Println(strings.Repeat(" ", 2) + `(use "git log --oneline --graph" to visualize the commit history)`)
		for _, msg := range noTotalOrderMsgs {
			fmt.Println(strings.Repeat(" ", 8) + theme.QuaternaryColor.Render(msg))
		}
	}
}

func printMergedBranches(branches []string, defaultBranch string, theme config.Theme) {
	branches = slices.Filter(branches, func(b string) bool {
		return b != defaultBranch
	})
	if len(branches) > 0 {
		fmt.Println()
		fmt.Printf("Excluding branches merged into %s:\n", defaultBranch)
		fmt.Println(strings.Repeat(" ", 2) + `(use "git commit ..." to add a commit on a branch for it to appear as a stack)`)
		fmt.Println(strings.Repeat(" ", 2) + `(use "git branch -D <branch>" to remove unneeded branches)`)
	}
	for _, b := range branches {
		fmt.Println(strings.Repeat(" ", 8) + theme.QuaternaryColor.Render(b))
	}
}

var (
	benchmarkCheckpoint time.Time
)

func benchmarkPoint(section string, msg string) {
	if !benchmarkFlag {
		return
	}

	elapsed := time.Since(benchmarkCheckpoint)
	benchmarkCheckpoint = time.Now()
	fmt.Printf("[benchmark:%s] %s - %v\n", section, msg, elapsed)
}
