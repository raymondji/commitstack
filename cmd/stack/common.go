package main

import (
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/config"
	"github.com/raymondji/git-stack/githost"
	"github.com/raymondji/git-stack/githost/gitlab"
	"github.com/raymondji/git-stack/libgit"
)

type deps struct {
	git     libgit.Git
	host    githost.Host
	repoCfg config.RepoConfig
	theme   config.Theme
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
	repoCfg, ok := cfg.Repositories[remote.RepoPath]
	if !ok {
		return deps{}, fmt.Errorf(
			"no config found for the current repo (%s)"+
				", please setup git stack using the `git stack init` command",
			remote.RepoPath)
	}

	var host githost.Host
	switch remote.Kind {
	case githost.Gitlab:
		host, err = gitlab.New(repoCfg.Gitlab.PersonalAccessToken)
		if err != nil {
			return deps{}, fmt.Errorf("failed to init gitlab client, err: %v", err)
		}
	default:
		return deps{}, fmt.Errorf("Unsupported git host %s", remote.Kind)
	}

	return deps{
		theme:   config.NewTheme(cfg.Theme),
		git:     git,
		host:    host,
		repoCfg: repoCfg,
	}, nil
}

func printProblems(stacks commitstack.Stacks) {
	if len(stacks.Errors) > 0 {
		fmt.Println()
		fmt.Println("Problems detected:")
		for _, err := range stacks.Errors {
			fmt.Printf("  %s\n", err.Error())
		}
	}
}
