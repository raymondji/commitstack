package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/raymondji/git-stack-cli/commitstack"
	"github.com/raymondji/git-stack-cli/config"
	"github.com/raymondji/git-stack-cli/githost"
	"github.com/raymondji/git-stack-cli/libgit"
)

type deps struct {
	git     libgit.Git
	host    githost.Host
	repoCfg config.RepoConfig
	theme   config.Theme
	remote  libgit.Remote
}

func initDeps() (deps, error) {
	git, err := libgit.New()
	if err != nil {
		return deps{}, err
	}
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
func printProblems(inference commitstack.InferenceResult) {
	validationErrMessages := []string{}
	for _, s := range inference.InferredStacks {
		for _, err := range s.ValidationErrors {
			validationErrMessages = append(validationErrMessages, err.Error())
		}
	}
	sort.Strings(validationErrMessages)
	if len(validationErrMessages) > 0 {
		fmt.Println()
		fmt.Println("Invalid stacks:")
		for _, msg := range validationErrMessages {
			fmt.Printf("  %s\n", msg)
		}
	}

	if len(inference.InferenceErrors) > 0 {
		fmt.Println()
		fmt.Println("Stack inference is ignoring incompatible commits:")
		for _, err := range inference.InferenceErrors {
			fmt.Printf("  %s\n", err.Error())
		}
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
