package main

import (
	"fmt"

	"github.com/raymondji/git-stack/commitstack"
	"github.com/raymondji/git-stack/libgit"
	"github.com/spf13/cobra"
)

func newPullCmd(git libgit.Git, defaultBranch string) *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pulls the latest changes from the default branch into the stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			stacks, err := commitstack.ComputeAll(git, defaultBranch)
			if err != nil {
				return err
			}
			stack, err := stacks.GetCurrent()
			if err != nil {
				return err
			}
			currBranch, err := git.GetCurrentBranch()
			if err != nil {
				return err
			}
			if currBranch != stack.Name() {
				return fmt.Errorf("must be on the tip of the stack to pull, currently checked out: %s, tip: %s",
					currBranch, stack.Name())
			}

			fmt.Printf("Pulling from %s into the current stack %s\n", defaultBranch, stack.Name())
			upstream, err := git.GetUpstream(defaultBranch)
			if err != nil {
				return err
			}
			refspec := fmt.Sprintf("%s:%s", upstream.BranchName, defaultBranch)
			if err := git.Fetch(upstream.Remote, refspec); err != nil {
				return err
			}
			res, err := git.Rebase(defaultBranch, libgit.RebaseOpts{})
			if err != nil {
				return err
			}
			fmt.Println(res)
			return nil
		},
	}
}
