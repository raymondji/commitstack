package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/raymondji/git-stack-cli/config"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/inference"
	"github.com/raymondji/git-stack-cli/libgit"
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
func printProblems(stacks []inference.Stack) {
	var invalidStackMsgs []string
	for _, stack := range stacks {
		got := stack.DivergesFrom()
		if len(got) > 0 {
			invalidStackMsgs = append(invalidStackMsgs, fmt.Sprintf(
				"%s diverges from %s",
				stack.Name,
				strings.Join(maps.Keys(got), ", "),
			))
		}
	}
	for _, stack := range stacks {
		_, err := stack.TotalOrderedBranches()
		if err != nil {
			invalidStackMsgs = append(invalidStackMsgs, err.Error())
		}
	}
	if len(invalidStackMsgs) > 0 {
		fmt.Println()
		fmt.Println("Invalid stacks:")
		fmt.Println(strings.Repeat(" ", 2) + "(use `git merge` or `git rebase` to resolve)")
		for _, msg := range invalidStackMsgs {
			fmt.Println(strings.Repeat(" ", 8) + msg)
		}
	}
}

func printMergedBranches(branches []string) {
	if len(branches) > 0 {
		fmt.Println()
		fmt.Println("Stack inference is ignoring branches merged into main:")
	}
	for _, b := range branches {
		fmt.Println(strings.Repeat(" ", 8) + b)
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
