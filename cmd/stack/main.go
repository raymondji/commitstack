package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/raymondji/git-stacked/concurrently"
	"github.com/raymondji/git-stacked/gitlib"
	"github.com/raymondji/git-stacked/gitplatform"
	"github.com/raymondji/git-stacked/gitplatform/gitlab"
	"github.com/raymondji/git-stacked/stackslib"
	"github.com/spf13/cobra"
)

// TODO: commands that are relative to the current stack should work
// even if some other stack is in an invalid state.
// TODO: list should still show the valid stacks even if some are invalid,
// but provide some notice about those invalid stacks and how to debug.
// This could happen easily if a repo uses multiple long lived branches with merge workflows,
// e.g. release branches.

const (
	configFileName = ".git-stacked.json"
)

type config struct {
	DefaultBranch string `json:"defaultBranch"`
}

var (
	defaultCfg = config{
		DefaultBranch: "main",
	}
)

func main() {
	var err error
	cfg, err := readConfigFile()
	if err != nil {
		log.Print(err.Error())
		return
	}

	// useGithubCli := isInstalled("gh")
	// useGitlabCli := isInstalled("glab")

	var rootCmd = &cobra.Command{
		Use:   "stack",
		Short: "A CLI tool for managing stacked Git branches.",
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the current version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("0.0.1")
		},
	}

	var addCmd = &cobra.Command{
		Use:   "add [branch_name]",
		Short: "Start a new stack or add a new branch onto the current stack",
		Long:  "Start a new stack if not currently in one, or add a new branch onto the current stack",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: this should fail if in a stack but not at the tip
			branchName := args[0]
			g := gitlib.Git{}
			if err := g.CreateBranch(branchName); err != nil {
				log.Fatal(err.Error())
			}
			if err := g.CommitEmpty(branchName); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all stacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gitlib.Git{}
			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}

			for _, s := range stacks.Entries {
				var prefix string
				if s.Current() {
					prefix = "*"
				} else {
					prefix = " "
				}

				fmt.Printf("%s %s (%d branches, %d commits)\n", prefix, s.Name(), len(s.LocalBranches), len(s.Commits))
			}

			printProblems(stacks)
			return nil
		},
	}

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push the stack to the remote and create merge requests.",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gitlib.Git{}
			var platform gitplatform.GitPlatform = gitlab.Gitlab{}

			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}
			s, err := stacks.GetCurrent()
			if err != nil {
				return err
			}

			wantTargets := map[string]string{}
			for i, b := range s.LocalBranches {
				if i == len(s.LocalBranches)-1 {
					wantTargets[b.Name] = cfg.DefaultBranch
				} else {
					wantTargets[b.Name] = s.LocalBranches[i+1].Name
				}
			}

			fmt.Println("Pushing stack...")
			// For safety, reset the target branch on any existing MRs if they don't match.
			// If any branches have been re-ordered, Gitlab can automatically consider the MRs merged.
			_, err = concurrently.ForEach(s.LocalBranches, func(branch stackslib.Branch) (gitplatform.PullRequest, error) {
				// TODO: I'm not sure if this scheme is 100% safe against branch reordering.
				pr, err := platform.GetPullRequest(branch.Name)
				if errors.Is(err, gitplatform.ErrDoesNotExist) {
					return gitplatform.PullRequest{}, nil
				} else if err != nil {
					return gitplatform.PullRequest{}, err
				}

				if pr.TargetBranch != wantTargets[branch.Name] {
					return platform.UpdatePullRequest(gitplatform.PullRequest{
						SourceBranch: branch.Name,
						TargetBranch: cfg.DefaultBranch,
						Description:  pr.Description,
					})
				}

				return gitplatform.PullRequest{}, nil
			})
			if err != nil {
				return fmt.Errorf("failed to force push branches, errors: %v", err)
			}

			// Push all branches.
			_, err = concurrently.ForEach(s.LocalBranches, func(branch stackslib.Branch) (string, error) {
				return g.ForcePush(branch.Name)
			})
			if err != nil {
				return fmt.Errorf("failed to force push branches, errors: %v", err.Error())
			}

			// Create MRs or update exising MRs to the right target branch.
			prs, err := concurrently.ForEach(s.LocalBranches, func(branch stackslib.Branch) (gitplatform.PullRequest, error) {
				pr, err := platform.GetPullRequest(branch.Name)
				if errors.Is(err, gitplatform.ErrDoesNotExist) {
					return platform.CreatePullRequest(gitplatform.PullRequest{
						SourceBranch: branch.Name,
						TargetBranch: wantTargets[branch.Name],
						Description:  "",
					})
				} else if err != nil {
					return gitplatform.PullRequest{}, err
				}

				return pr, nil
			})
			if err != nil {
				log.Fatalf("failed to create some pull requests, errors: %v", err.Error())
			}

			// Update PRs with info on the stacks.
			prs, err = concurrently.ForEach(prs, func(pr gitplatform.PullRequest) (gitplatform.PullRequest, error) {
				desc := formatPullRequestDescription(pr, prs)
				pr, err := platform.UpdatePullRequest(gitplatform.PullRequest{
					SourceBranch: pr.SourceBranch,
					TargetBranch: wantTargets[pr.SourceBranch],
					Description:  desc,
				})
				return pr, err
			})
			if err != nil {
				return err
			}

			for _, pr := range prs {
				fmt.Printf("Pushed %s: %s\n", pr.SourceBranch, pr.WebURL)
			}
			fmt.Println("Done")
			return nil
		},
	}

	var pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pulls the latest changes from the default branch into the stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gitlib.Git{}
			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}
			stack, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			fmt.Printf("Pulling from %s into the current stack %s\n", cfg.DefaultBranch, stack.Name())
			// TODO: this should fail if not at the tip of the stack

			res, err := g.Fetch()
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(res)

			res, err = g.Rebase(cfg.DefaultBranch)
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(res)
			return nil
		},
	}

	var editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit the stack using interactive rebase",
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gitlib.Git{}
			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}
			stack, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			fmt.Printf("Pulling from %s into the current stack %s\n", cfg.DefaultBranch, stack.Name())

			if err := g.RebaseInteractiveKeepBase(cfg.DefaultBranch); err != nil {
				log.Fatal(err.Error())
			}
			return nil
		},
	}

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show all branches in the current stack",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			g := gitlib.Git{}
			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}

			var stack stackslib.Stack
			if len(args) == 0 {
				stack, err = stacks.GetCurrent()
				if err != nil {
					return err
				}
			} else {
				wantStack := args[0]
				var found bool
				for _, s := range stacks.Entries {
					if s.Name() == wantStack {
						stack = s
						found = true
					}
				}
				if !found {
					return fmt.Errorf("no stack named: %s", wantStack)
				}
			}

			for i, b := range stack.LocalBranches {
				var prefix, suffix string
				if i == 0 {
					suffix = "(top)"
				}
				if b.Current {
					prefix = "*"
				} else {
					prefix = " "
				}

				fmt.Printf("%s %s %s\n", prefix, b.Name, suffix)
			}

			if isSharingHistory(stacks, stack.Name()) {
				printProblems(stacks)
			}
			return nil
		},
	}

	rootCmd.AddCommand(versionCmd, addCmd, editCmd, listCmd, showCmd, pushCmd, pullCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func formatPullRequestDescription(
	currPR gitplatform.PullRequest, prs []gitplatform.PullRequest,
) string {
	var newStackDescParts []string
	currIndex := slices.IndexFunc(prs, func(pr gitplatform.PullRequest) bool {
		return pr.SourceBranch == currPR.SourceBranch
	})
	for i, pr := range prs {
		var prefix string
		if i == currIndex {
			prefix = "Current: "
		} else if i == currIndex-1 {
			prefix = "Next: "
		} else if i == currIndex+1 {
			prefix = "Prev: "
		}
		newStackDescParts = append(newStackDescParts, fmt.Sprintf("- %s%s", prefix, pr.MarkdownWebURL))
	}
	newStackDesc := strings.Join(newStackDescParts, "\n")

	beginMarker := "Pull Request Stack:"
	endMarker := "<!-- End Pull Request Stack -->"
	newSection := fmt.Sprintf("%s\n%s\n\n%s", beginMarker, newStackDesc, endMarker)
	sectionPattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker))

	if sectionPattern.MatchString(currPR.Description) {
		return sectionPattern.ReplaceAllString(currPR.Description, newSection)
	} else {
		return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(currPR.Description), newSection)
	}
}

func isSharingHistory(stacks stackslib.Stacks, stackName string) bool {
	for _, grp := range stacks.SharingHistory {
		for _, s := range grp {
			if s == stackName {
				return true
			}
		}
	}
	return false
}

func printProblems(stacks stackslib.Stacks) {
	if len(stacks.SharingHistory) > 0 {
		fmt.Println()
		fmt.Println("Problems:")
		for _, grp := range stacks.SharingHistory {
			fmt.Printf("  %s have diverged, please reconcile (e.g. by rebasing one stack onto another)\n", strings.Join(grp, ", "))
		}
	}
}

// readConfigFile reads the specified configuration file from the root of the Git repository.
func readConfigFile() (config, error) {
	g := gitlib.Git{}
	dir, err := g.GetRootDir()
	if err != nil {
		return config{}, err
	}

	configFilePath := filepath.Join(dir, configFileName)
	content, err := os.ReadFile(configFilePath)
	// Return a default configuration if the file doesn't exist
	if errors.Is(err, os.ErrNotExist) {
		return defaultCfg, nil
	} else if err != nil {
		return config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}

func isInstalled(file string) bool {
	_, err := exec.LookPath(file)
	return err != nil
}
