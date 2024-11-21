package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/raymondji/git-stacked/git"
	"github.com/raymondji/git-stacked/stack"
	"github.com/spf13/cobra"
)

const (
	configFileName = ".git-stacked.json"
)

type Config struct {
	GithubIntegration bool
	GitlabIntegration bool
}

var (
	gsBaseBranch        string
	gsEnableColorOutput bool
	gsEnableGitLabExt   bool
	gsEnableGitHubExt   bool
	gsEnableDebugOutput bool
)

func main() {
	gsBaseBranch = getEnv("GS_BASE_BRANCH", "main")
	// useGithubCli := isInstalled("gh")
	// useGitlabCli := isInstalled("glab")

	var rootCmd = &cobra.Command{
		Use:   "stack",
		Short: "A CLI tool for managing stacked Git branches.",
	}

	rootCmd.AddCommand(listCmd, branchCmd, pushCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stacks",
	Run: func(cmd *cobra.Command, args []string) {
		g := git.Git{}
		stacks, err := stack.GetAll(g, gsBaseBranch)
		if err != nil {
			log.Fatalf("Failed to list stacks, err: %v", err)
		}

		for _, s := range stacks {
			var prefix string
			if s.Current {
				prefix = "*"
			} else {
				prefix = " "
			}

			fmt.Println("%s %s %s", prefix, s.LocalBranches[len(s.LocalBranches)-1])
		}
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all branches in the stack",
	Run: func(cmd *cobra.Command, args []string) {
		g := git.Git{}
		stack, err := stack.GetCurrent(g, gsBaseBranch)
		if err != nil {

		}
		for _, b := range stack.LocalBranches {
			fmt.Printf("Pushing branch: %s\n", b.Name)
		}
	},
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "List all branches in the current stack",
	Run: func(cmd *cobra.Command, args []string) {
		g := git.Git{}
		stack, err := stack.GetCurrent(g, gsBaseBranch)
		if err != nil {
			log.Fatalf("Failed to list branches, err: %v", err)
		}

		for i, b := range stack.LocalBranches {
			var prefix, suffix string
			if i == 0 {
				suffix = "(bottom)"
			} else if i == len(stack.LocalBranches)-1 {
				suffix = "(top)"
			}
			if b.Current {
				prefix = "*"
			} else {
				prefix = " "
			}

			fmt.Println("%s %s %s", prefix, b.Name, suffix)
		}
	},
}

// readConfigFile reads the specified configuration file from the root of the Git repository.
func readConfigFile() (string, error) {
	g := git.Git{}
	gitRoot, err := g.GetRootDir()
	if err != nil {
		return "", err
	}

	configFilePath := filepath.Join(gitRoot, configFileName)
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	return string(content), nil
}

func isInstalled(file string) bool {
	_, err := exec.LookPath(file)
	return err != nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
