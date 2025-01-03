package main

import (
	"fmt"

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

	remote, err := git.GetRemote()
	if err != nil {
		return deps{}, err
	}

	cfg, err := config.Load()
	if err != nil {
		return deps{}, fmt.Errorf("failed to load config, err: %v", err.Error())
	}
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

	return deps{
		theme:   config.NewTheme(cfg.Theme),
		git:     git,
		host:    host,
		repoCfg: repoCfg,
		remote:  remote,
	}, nil
}

// TODO: try to format this similar to git status.
/*
Unmerged paths:
  (use "git add <file>..." to mark resolution)
        both modified:   src/module1.py
        both added:      src/module2.py
*/
func printProblems(stacks commitstack.Stacks) {
	if len(stacks.Errors) > 0 {
		fmt.Println()
		fmt.Println("Unable to infer all stacks:")
		for _, err := range stacks.Errors {
			fmt.Printf("  %s\n", err.Error())
		}
	}
}
