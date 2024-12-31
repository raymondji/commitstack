package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/raymondji/git-stack/config"
	"github.com/raymondji/git-stack/githost"
	"github.com/raymondji/git-stack/githost/github"
	"github.com/raymondji/git-stack/githost/gitlab"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config for the current git repo",
	RunE: func(cmd *cobra.Command, args []string) error {
		git, err := libgit.New()
		if err != nil {
			return err
		}

		remote, err := git.GetRemote()
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config, err: %v", err.Error())
		}
		if len(cfg.Repositories) == 0 {
			cfg.Repositories = map[string]config.RepoConfig{}
		}

		_, ok := cfg.Repositories[remote.RepoPath]
		if ok {
			fmt.Printf("Repository '%s' already exists in the config. Do you want to overwrite it? (y/n): ", remote.RepoPath)
			overwrite, err := promptUserConfirmation()
			if err != nil {
				return err
			}
			if !overwrite {
				fmt.Println("Cancelling.")
				return nil
			}
		}

		switch remote.Kind {
		case githost.Gitlab:
			fmt.Print("Enter your GitLab personal access token: ")
			personalAccessToken, err := promptUserInput()
			if err != nil {
				return err
			}

			host, err := gitlab.New(personalAccessToken)
			if err != nil {
				return err
			}
			repo, err := host.GetRepo(remote.RepoPath)
			if err != nil {
				return err
			}

			cfg.Repositories[remote.RepoPath] = config.RepoConfig{
				Gitlab: config.GitlabConfig{
					PersonalAccessToken: personalAccessToken,
				},
				DefaultBranch: repo.DefaultBranch,
			}
		case githost.Github:
			fmt.Println("Git stack requires a Github personal access token in order to manage pull requests on your behalf.")
			fmt.Println("You can create a personal access token at https://github.com/settings/personal-access-tokens/new.")
			fmt.Println("At minimum, the following fine grained permissions are required:")
			fmt.Println("- Repository permissions (Pull Requests): Read and write ")
			fmt.Println("- Repository permissions (Contents): Read-only")
			fmt.Println("- Repository permissions (Metadata): Read-only")
			fmt.Print("To continue, enter your Github personal access token: ")
			personalAccessToken, err := promptUserInput()
			if err != nil {
				return err
			}

			fmt.Print("Enter your Github username:")
			username, err := promptUserInput()
			if err != nil {
				return err
			}

			host, err := github.New(username, personalAccessToken)
			if err != nil {
				return err
			}

			repo, err := host.GetRepo(remote.RepoPath)
			if err != nil {
				return err
			}

			cfg.Repositories[remote.RepoPath] = config.RepoConfig{
				Github: config.GithubConfig{
					Username:            username,
					PersonalAccessToken: personalAccessToken,
				},
				DefaultBranch: repo.DefaultBranch,
			}
		default:
			return fmt.Errorf("Unsupported git host %s", remote.Kind)
		}

		cfgPath, err := config.Save(cfg)
		if err != nil {
			return err
		}
		fmt.Println("Config saved to", cfgPath)
		return nil
	},
}

func promptUserConfirmation() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid input: %s", input)
	}
}

func promptUserInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
