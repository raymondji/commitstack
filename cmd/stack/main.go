package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raymondji/git-stacked/concurrently"
	"github.com/raymondji/git-stacked/githost"
	"github.com/raymondji/git-stacked/githost/gitlab"
	"github.com/raymondji/git-stacked/gitlib"
	"github.com/raymondji/git-stacked/stackslib"
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
		fmt.Println(err.Error())
		return
	}

	ok, err := isInstalled("glab")
	if err != nil {
		fmt.Println(err.Error())
		return
	} else if !ok {
		fmt.Println("glab CLI must be installed")
		return
	}

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
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: this should fail if in a stack but not at the tip
			branchName := args[0]
			g := gitlib.Git{}
			if err := g.CreateBranch(branchName); err != nil {
				return err
			}
			return g.CommitEmpty(branchName)
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
				var prefix, suffix string
				if s.Current() {
					prefix = "*"
				} else {
					prefix = " "
				}
				if s.Error != nil {
					suffix = ", problem!"
				}

				fmt.Printf("%s %s (%d branches, %d commits%s)\n", prefix, s.Name(), len(s.LocalBranches()), len(s.Commits), suffix)
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
			var host githost.GitHost = gitlab.Gitlab{}

			stacks, err := stackslib.Compute(g, cfg.DefaultBranch)
			if err != nil {
				return err
			}
			s, err := stacks.GetCurrent()
			if err != nil {
				return err
			}

			wantTargets := map[string]string{}
			lb := s.LocalBranches()
			for i, b := range lb {
				if i == len(lb)-1 {
					wantTargets[b.Name] = cfg.DefaultBranch
				} else {
					wantTargets[b.Name] = lb[i+1].Name
				}
			}

			fmt.Println("Pushing stack...")
			// For safety, reset the target branch on any existing MRs if they don't match.
			// If any branches have been re-ordered, Gitlab can automatically consider the MRs merged.
			_, err = concurrently.Map(lb, func(branch stackslib.Branch) (githost.PullRequest, error) {
				// TODO: I'm not sure if this scheme is 100% safe against branch reordering.
				pr, err := host.GetPullRequest(branch.Name)
				if errors.Is(err, githost.ErrDoesNotExist) {
					return githost.PullRequest{}, nil
				} else if err != nil {
					return githost.PullRequest{}, err
				}

				if pr.TargetBranch != wantTargets[branch.Name] {
					return host.UpdatePullRequest(githost.PullRequest{
						SourceBranch: branch.Name,
						TargetBranch: cfg.DefaultBranch,
						Description:  pr.Description,
					})
				}

				return githost.PullRequest{}, nil
			})
			if err != nil {
				return fmt.Errorf("failed to force push branches, errors: %v", err)
			}

			// Push all branches.
			localBranches := s.LocalBranches()
			_, err = concurrently.Map(localBranches, func(branch stackslib.Branch) (string, error) {
				return g.PushForceWithLease(branch.Name)
			})
			if err != nil {
				return fmt.Errorf("failed to force push branches, errors: %v", err.Error())
			}

			// Create MRs or update exising MRs to the right target branch.
			prs, err := concurrently.Map(localBranches, func(branch stackslib.Branch) (githost.PullRequest, error) {
				pr, err := host.GetPullRequest(branch.Name)
				if errors.Is(err, githost.ErrDoesNotExist) {
					return host.CreatePullRequest(githost.PullRequest{
						SourceBranch: branch.Name,
						TargetBranch: wantTargets[branch.Name],
						Description:  "",
					})
				} else if err != nil {
					return githost.PullRequest{}, err
				}

				return pr, nil
			})
			if err != nil {
				return err
			}

			// Update PRs with info on the stacks.
			prs, err = concurrently.Map(prs, func(pr githost.PullRequest) (githost.PullRequest, error) {
				desc := formatPullRequestDescription(pr, prs)
				pr, err := host.UpdatePullRequest(githost.PullRequest{
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
			currBranch, err := g.GetCurrentBranch()
			if err != nil {
				return err
			}
			if currBranch != stack.Name() {
				return fmt.Errorf("must be on the tip of the stack to pull, currently checked out: %s, tip: %s",
					currBranch, stack.Name())
			}

			fmt.Printf("Pulling from %s into the current stack %s\n", cfg.DefaultBranch, stack.Name())
			upstream, err := g.GetUpstream(cfg.DefaultBranch)
			if err != nil {
				return err
			}
			refspec := fmt.Sprintf("%s:%s", upstream.BranchName, cfg.DefaultBranch)
			if err := g.Fetch(upstream.Remote, refspec); err != nil {
				return err
			}
			res, err := g.Rebase(cfg.DefaultBranch)
			if err != nil {
				return err
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

			if err := g.RebaseInteractive(cfg.DefaultBranch, "--keep-base", "--autosquash"); err != nil {
				return err
			}
			return nil
		},
	}

	var fixupCmd = &cobra.Command{
		Use:   "fixup",
		Short: "Create a commit to fixup a branch in the stack",
		Args:  cobra.MaximumNArgs(1),
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

			var branchToFix string
			if len(args) == 1 {
				branchToFix = args[0]
			} else {
				var opts []huh.Option[string]
				for _, b := range stack.LocalBranches() {
					opts = append(opts, huh.NewOption(b.Name, b.Name))
				}
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("Choose which branch to fixup").
							Options(opts...).
							Filtering(true).
							Value(&branchToFix),
					),
				)
				err = form.Run()
				if err != nil {
					return err
				}
			}

			hash, err := g.GetCommitHash(branchToFix)
			if err != nil {
				return err
			}

			res, err := g.CommitFixup(hash)
			if err != nil {
				return err
			}
			fmt.Println(res)
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
					printProblems(stacks)
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

			showPRs, err := cmd.Flags().GetBool("prs")
			if err != nil {
				return err
			}

			prsBySrcBranch := map[string]githost.PullRequest{}
			if showPRs {
				var host githost.GitHost = gitlab.Gitlab{}
				prs, err := concurrently.Map(stack.LocalBranches(), func(branch stackslib.Branch) (githost.PullRequest, error) {
					pr, err := host.GetPullRequest(branch.Name)
					if errors.Is(err, githost.ErrDoesNotExist) {
						return githost.PullRequest{}, nil
					} else if err != nil {
						return githost.PullRequest{}, err
					}

					return pr, nil
				})
				if err != nil {
					return err
				}
				for _, pr := range prs {
					if pr.SourceBranch == "" {
						continue
					}
					prsBySrcBranch[pr.SourceBranch] = pr
				}
			}

			for i, b := range stack.LocalBranches() {
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
				if showPRs {
					if pr, ok := prsBySrcBranch[b.Name]; ok {
						fmt.Printf("  └── %s\n", pr.WebURL)
						fmt.Println()
					} else {
						fmt.Println()
					}
				}
			}

			printProblem(stack)
			return nil
		},
	}
	showCmd.Flags().Bool("prs", false, "Whether to show PRs for each branch")

	rootCmd.SilenceUsage = true
	rootCmd.AddCommand(versionCmd, addCmd, editCmd, fixupCmd, listCmd, showCmd, pushCmd, pullCmd)
	rootCmd.Execute()
}

func formatPullRequestDescription(
	currPR githost.PullRequest, prs []githost.PullRequest,
) string {
	var newStackDescParts []string
	currIndex := slices.IndexFunc(prs, func(pr githost.PullRequest) bool {
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
	newStackDesc := "Pull request stack:\n" + strings.Join(newStackDescParts, "\n")

	beginMarker := "<!-- DO NOT EDIT: generated by git stack push (start)-->"
	endMarker := "<!-- DO NOT EDIT: generated by git stack push (end) -->"
	newSection := fmt.Sprintf("%s\n%s\n%s", beginMarker, newStackDesc, endMarker)
	sectionPattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(beginMarker) + `.*?` + regexp.QuoteMeta(endMarker))

	if sectionPattern.MatchString(currPR.Description) {
		return sectionPattern.ReplaceAllString(currPR.Description, newSection)
	} else {
		return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(currPR.Description), newSection)
	}
}

func printProblem(stack stackslib.Stack) {
	if stack.Error != nil {
		fmt.Println()
		fmt.Println("Problems:")
		fmt.Printf("  %s\n", stack.Error.Error())
	}
}

func printProblems(stacks stackslib.Stacks) {
	if len(stacks.Errors) > 0 {
		fmt.Println()
		fmt.Println("Problems:")
		for _, err := range stacks.Errors {
			fmt.Printf("  %s\n", err.Error())
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

func isInstalled(file string) (bool, error) {
	_, err := exec.LookPath(file)
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		// Generally returned when file is not a executable
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking if %s is installed, err: %v", file, err)
	}
	return true, nil
}
