package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var learnPartFlag int

func init() {
	learnCmd.Flags().IntVar(&learnPartFlag, "part", 1, "Which part of the tutorial to continue from")
}

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Prints sample commands to learn how to use commitstack",
	RunE: func(cmd *cobra.Command, args []string) error {
		deps, err := initDeps()
		if err != nil {
			return err
		}
		shellCmd := func(s string) string { return deps.theme.PrimaryColor.Render(s) }

		switch learnPartFlag {
		case 1:
			printLines(
				"Welcome to Commitstack!",
				"Here is a quick tutorial on how to use the CLI.",
				"First, let's start on the default branch:",
				shellCmd(fmt.Sprintf("git checkout %s", deps.repoCfg.DefaultBranch)),
				"",
				"Next, let's create our first branch:",
				shellCmd(`git checkout -b learncommitstack && \`),
				shellCmd(`echo 'hello world' > learncommitstack.txt && \`),
				shellCmd(`git add . && \`),
				shellCmd(`git commit -m 'hello world'`),
				"",
				"Now let's stack a second branch on top of our first:",
				shellCmd(`git checkout -b learncommitstack-pt2 && \`),
				shellCmd(`echo 'have a break' > learncommitstack.txt && \`),
				shellCmd(`git commit -am 'break' && \`),
				shellCmd(`echo 'have a kitkat' > learncommitstack.txt && \`),
				shellCmd(`git commit -am 'kitkat'`),
				"",
				"So far everything we've done has been normal Git. Let's see what Commitstack can do for us!",
				"Our current stack has two branches in it, which we can see with:",
				shellCmd(`git stack show`),
				"Our current stack has 3 commits in it, which we can see with:",
				shellCmd(`git stack log`),
				"",
				"We can easily push all branches in the stack up as separate PRs:",
				"Commitstack automatically sets the target branches for you on the PRs.",
				shellCmd(`git stack push`),
				"We can also quickly view the PRs in the stack using:",
				shellCmd(`git stack show --prs`),
				"",
				"Nice! All done part 1 of the tutorial. In part 2 we'll learn how to make more changes to a stack.",
				"Once you're ready, continue the tutorial using:",
				shellCmd("git stack learn --part 2"),
			)
		case 2:
			fmt.Println("TODO")
		default:
			return errors.New("invalid tutorial part")
		}

		return nil
	},
}

func printLines(lines ...string) {
	for _, l := range lines {
		fmt.Println(l)
	}
}
