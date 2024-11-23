package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/raymondji/git-stacked/concurrently"
	"github.com/raymondji/git-stacked/gitlab"
	"github.com/raymondji/git-stacked/gitlib"
	"github.com/raymondji/git-stacked/stack"
	"github.com/spf13/cobra"
)

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

	var addCmd = &cobra.Command{
		Use:   "add [branch_name]",
		Short: "Start a new stack or add a new branch onto the current stack",
		Long:  "Start a new stack if not currently in one, or add a new branch onto the current stack",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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
		Run: func(cmd *cobra.Command, args []string) {
			g := gitlib.Git{}
			stacks, err := stack.GetAll(g, cfg.DefaultBranch)
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

				fmt.Printf("%s %s\n", prefix, s.Name)
			}
		},
	}

	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Force push all branches in the stack",
		Run: func(cmd *cobra.Command, args []string) {
			g := gitlib.Git{}
			glab := gitlab.Gitlab{}
			s, err := stack.GetCurrent(g, cfg.DefaultBranch)
			if err != nil {
				log.Fatalf("Failed to get current stack, err: %v", err)
			}

			wantTargets := map[string]string{}
			for i, b := range s.LocalBranches {
				if i == len(s.LocalBranches)-1 {
					wantTargets[b.Name] = cfg.DefaultBranch
				} else {
					wantTargets[b.Name] = s.LocalBranches[i+1].Name
				}
			}

			// For safety, reset the target branch on any existing MRs if they don't match.
			// If any branches have been re-ordered, Gitlab can automatically consider the MRs merged.
			res, err := concurrently.ForEach(s.LocalBranches, func(branch stack.Branch) (string, error) {
				// TODO: I'm not sure if this scheme is 100% safe against branch reordering.
				currTarget, err := glab.GetMRTargetBranch(branch.Name)
				if errors.Is(err, gitlab.ErrNoMRForBranch) {
					return "", nil
				} else if err != nil {
					return "", err
				}

				if currTarget != wantTargets[branch.Name] {
					return glab.SetMRTargetBranch(branch.Name, cfg.DefaultBranch)
				}

				return "", nil
			})
			if err != nil {
				log.Fatalf("failed to force push branches, errors: %v", err.Error())
			}
			for _, r := range res {
				fmt.Println(r)
			}
			fmt.Println("Done resetting existing MRs")

			// Push all branches.
			res, err = concurrently.ForEach(s.LocalBranches, func(branch stack.Branch) (string, error) {
				return g.ForcePush(branch.Name)
			})
			if err != nil {
				log.Fatalf("failed to force push branches, errors: %v", err.Error())
			}
			for _, r := range res {
				fmt.Println(r)
			}
			fmt.Println("Done pushing branches")

			// Create MRs or update exising MRs to the right target branch.
			res, err = concurrently.ForEach(s.LocalBranches, func(branch stack.Branch) (string, error) {
				currTarget, err := glab.GetMRTargetBranch(branch.Name)
				if errors.Is(err, gitlab.ErrNoMRForBranch) {
					return glab.CreateMR(branch.Name, wantTargets[branch.Name])
				} else if err != nil {
					return "", err
				}

				if currTarget != wantTargets[branch.Name] {
					return glab.SetMRTargetBranch(branch.Name, wantTargets[branch.Name])
				}

				return "", nil
			})
			if err != nil {
				log.Fatalf("failed to create merge requests, errors: %v", err.Error())
			}
			for _, r := range res {
				fmt.Println(r)
			}
		},
	}

	var pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pulls the latest changes from the defaultBranch into the stack",
		Run: func(cmd *cobra.Command, args []string) {
			g := gitlib.Git{}
			stack, err := stack.GetCurrent(g, cfg.DefaultBranch)
			if err != nil {
				log.Fatalf("Failed to list branches, err: %v", err)
			}
			fmt.Printf("Pulling from %s into the current stack %s\n", cfg.DefaultBranch, stack.Name)

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
		},
	}

	var rebaseCmd = &cobra.Command{
		Use:   "rebase",
		Short: "Rebase",
		Run: func(cmd *cobra.Command, args []string) {
			g := gitlib.Git{}
			stack, err := stack.GetCurrent(g, cfg.DefaultBranch)
			if err != nil {
				log.Fatalf("Failed to list branches, err: %v", err)
			}
			fmt.Printf("Pulling from %s into the current stack %s\n", cfg.DefaultBranch, stack.Name)

			if err := g.RebaseInteractiveKeepBase(cfg.DefaultBranch); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	var branchCmd = &cobra.Command{
		Use:   "branch",
		Short: "List all branches in the current stack",
		Run: func(cmd *cobra.Command, args []string) {
			g := gitlib.Git{}
			stack, err := stack.GetCurrent(g, cfg.DefaultBranch)
			if err != nil {
				log.Fatalf("Failed to list branches, err: %v", err)
			}

			for i, b := range stack.LocalBranches {
				var prefix, suffix string
				if i == 0 {
					suffix = "(top)"
				} else if i == len(stack.LocalBranches)-1 {
					suffix = "(bot)"
				}
				if b.Current {
					prefix = "*"
				} else {
					prefix = " "
				}

				fmt.Printf("%s %s %s\n", prefix, b.Name, suffix)
			}
		},
	}

	rootCmd.AddCommand(addCmd, rebaseCmd, listCmd, branchCmd, pushCmd, pullCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
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

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
