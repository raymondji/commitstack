package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	gsBaseBranch        string
	gsEnableColorOutput bool
	gsEnableGitLabExt   bool
	gsEnableGitHubExt   bool
	gsEnableDebugOutput bool
)

func main() {
	gsBaseBranch = getEnv("GS_BASE_BRANCH", "main")
	gsEnableColorOutput = getEnv("GS_ENABLE_COLOR_OUTPUT", "true") == "true"
	gsEnableGitLabExt = getEnv("GS_ENABLE_GITLAB_EXTENSION", "false") == "true"
	gsEnableGitHubExt = getEnv("GS_ENABLE_GITHUB_EXTENSION", "false") == "true"
	gsEnableDebugOutput = getEnv("GS_ENABLE_DEBUG_OUTPUT", "false") == "true"

	var rootCmd = &cobra.Command{
		Use:   "stack",
		Short: "A CLI tool for managing stacked Git branches.",
	}

	rootCmd.AddCommand(helpCmd, stackCmd, allCmd, pushCmd, pullCmd, branchCmd, logCmd, rebaseCmd, reorderCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Displays help information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("usage: git-stacked <subcommand> ...\n" +
			"alias: gs\n\n" +
			"subcommands:\n\n" +
			"stack - Create a new branch stacked on top of the current one\n" +
			"all   - List all stacks\n" +
			"push  - Push all branches in the current stack\n" +
			"pull  - Update base branch and rebase stack\n" +
			"branch- List all branches in the current stack\n" +
			"log   - Show logs for the current stack\n" +
			"rebase- Start interactive rebase of the current stack\n" +
			"reorder- Reorder branches interactively")
	},
}

var stackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Stack a new branch on top of the current one",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Error: Branch name is required")
			return
		}
		branch := args[0]
		currentBranch := gitCommand("rev-parse", "--abbrev-ref", "HEAD")
		if !isTopOfStack() && currentBranch != gsBaseBranch {
			fmt.Println("Can only run from the base branch or top of stack")
			return
		}
		gitCommand("checkout", "-b", branch)
		gitCommand("commit", "--allow-empty", "-m", fmt.Sprintf("Start of %s", branch))
	},
}

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "List all stacks",
	Run: func(cmd *cobra.Command, args []string) {
		branches := gitCommand("branch", "--format=%(refname:short)")
		currentBranch := gitCommand("rev-parse", "--abbrev-ref", "HEAD")
		stackedBranches := parseBranches(branches)

		for _, branch := range stackedBranches {
			if branch == currentBranch {
				fmt.Printf("* \033[0;32m%s\033[0m\n", branch) // Green for active branch
			} else {
				fmt.Println("  ", branch)
			}
		}
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all branches in the stack",
	Run: func(cmd *cobra.Command, args []string) {
		if !isTopOfStack() {
			fmt.Println("Error: Must be at the top of the stack")
			return
		}

		branches := gitCommand("log", "--pretty=format:%D", fmt.Sprintf("%s..", gsBaseBranch), "--decorate-refs=refs/heads")
		for _, branch := range strings.Split(branches, "\n") {
			branchName := strings.TrimSpace(branch)
			fmt.Printf("Pushing branch: %s\n", branchName)
			gitCommand("push", "origin", fmt.Sprintf("%s:%s", branchName, branchName), "--force")
		}
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Update base branch and rebase stack",
	Run: func(cmd *cobra.Command, args []string) {
		if !isTopOfStack() {
			fmt.Println("Error: Must be at the top of the stack")
			return
		}
		gitCommand("checkout", gsBaseBranch)
		gitCommand("pull")
		gitCommand("checkout", "-")
		gitCommand("rebase", gsBaseBranch, "--update-refs")
	},
}

// Other commands like branchCmd, logCmd, rebaseCmd, and reorderCmd would follow the same pattern.

func gitCommand(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running git %s: %s\n", strings.Join(args, " "), err.Error())
		os.Exit(1)
	}
	return strings.TrimSpace(string(output))
}

func isTopOfStack() bool {
	currentBranch := gitCommand("rev-parse", "--abbrev-ref", "HEAD")
	descendentCount := gitCommand("branch", "--contains", currentBranch, "|", "wc", "-l")
	return descendentCount == "1"
}

func parseBranches(branches string) []string {
	var result []string
	for _, branch := range strings.Split(branches, "\n") {
		if strings.TrimSpace(branch) != gsBaseBranch {
			result = append(result, strings.TrimSpace(branch))
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "List all branches in the current stack",
	Run: func(cmd *cobra.Command, args []string) {
		if !isTopOfStack() {
			fmt.Println("Error: Must be at the top of the stack")
			return
		}
		if !isStackValid() {
			fmt.Println("Error: Stack is invalid")
			return
		}

		currentBranch := gitCommand("rev-parse", "--abbrev-ref", "HEAD")
		branches := gitCommand("log", "--pretty=format:%D", fmt.Sprintf("%s..", gsBaseBranch), "--decorate-refs=refs/heads")

		for _, branch := range strings.Split(branches, "\n") {
			branch = strings.TrimSpace(branch)
			if branch == currentBranch {
				if gsEnableColorOutput {
					fmt.Printf("* \033[0;32m%s\033[0m (top)\n", branch)
				} else {
					fmt.Printf("* %s (top)\n", branch)
				}
			} else {
				fmt.Printf("  %s\n", branch)
			}
		}
	},
}
