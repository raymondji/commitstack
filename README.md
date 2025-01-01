# commitstack

CLI to facilitate [stacking workflows](https://www.stacking.dev/) in git. Support for Gitlab (!) and Github.

## Why another stacking tool?

Stacking with plain Git is totally doable but cumbersome. There's a plethora of tools that make things easier for Github, but after I started using Gitlab at work, I was surprised to find almost nothing that fit the bill.

The only tool I came across is the experimental `glab stack` CLI command, but it's missing a few features I wanted and has the opposite problem: it's Gitlab only. I ideally wanted a single tool I could use with both.

Thus, commitstack was born. It was initially a set of bash aliases and functions, and turned into a Go binary as I kept adding to it. The project currently supports Gitlab and Github, and should (hopefully) be easily extensible to other Git hosting providers.

## Goals

Stacking workflows in Git can be cumbersome both because of your interactions with Git itself (e.g. pushing multiple branches and keeping track of your stacks), and because of your interactions with Gitlab/Github/etc. (e.g. opening PRs and setting target branches). commitstack aims to make both of these aspects easier.

In addition, commitstack tries to fit into Git as naturally as possible and avoid introducing whole new paradigms. Git already has the basic building blocks needed to do stacking workflows (e.g. with `git rebase --update-refs`), and many people stack without any extra tools. For everyone in that bucket, commitstack should (hopefully) feel like it fits right in (while removing much of the friction).

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

To install commitstack:
```
go install github.com/raymondji/commitstack/cmd/git-stack@latest
```

## Getting started

The commitstack binary is named `git-stack`. Git offers a handy trick allowing binaries named `git-<foo>` to be invoked as git subcommands, so commitstack can be invoked as `git stack`.

commitstack needs a Gitlab/Github personal access token in order to manage MRs/PRs for you. To set this up:
```
cd ~/your/git/repo
git stack init
```

All set! To learn how to use commitstack, you can access an interactive tutorial built-in to the CLI:
```
git stack learn
```

## Sample usage (output from `git stack learn`)

```
╭──────────────────────────────────────────────────╮
│                                                  │
│ Welcome to commitstack!                          │
│ Here is a quick tutorial on how to use the CLI.  │
│                                                  │
╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────╮
│                                                  │
│ First, let's start on the default branch:        │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout main
Your branch is up to date with 'origin/main'.
╭──────────────────────────────────────────────────╮
│                                                  │
│ Next, let's create our first branch:             │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout -b learncommitstack
> echo 'hello world' > learncommitstack.txt
> git add .
> git commit -m 'hello world'
[learncommitstack 400ab52] hello world
 1 file changed, 1 insertion(+)
 create mode 100644 learncommitstack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ Now let's stack a second branch on top of our    │
│ first:                                           │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout -b learncommitstack-pt2
> echo 'have a break' >> learncommitstack.txt
> git commit -am 'break'
[learncommitstack-pt2 b78bbdb] break
 1 file changed, 1 insertion(+)
> echo 'have a kitkat' >> learncommitstack.txt
> git commit -am 'kitkat'
[learncommitstack-pt2 60b3b2c] kitkat
 1 file changed, 1 insertion(+)
╭──────────────────────────────────────────────────╮
│                                                  │
│ So far everything we've done has been normal     │
│ Git. Let's see what commitstack can do for us    │
│ already.                                         │
│                                                  │
╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────╮
│                                                  │
│ Our current stack has two branches in it, which  │
│ we can see with:                                 │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show
On stack learncommitstack-pt2
Branches in stack:
* learncommitstack-pt2 (top)
  learncommitstack
╭──────────────────────────────────────────────────╮
│                                                  │
│ Our current stack has 3 commits in it, which we  │
│ can see with:                                    │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack log
* 60b3b2c (learncommitstack-pt2) kitkat
  b78bbdb break
  400ab52 (learncommitstack) hello world
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can easily push all branches in the stack up  │
│ as separate PRs.                                 │
│ commitstack automatically sets the target        │
│ branches for you on the PRs.                     │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack push
Pushed learncommitstack-pt2: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/57
Pushed learncommitstack: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/58
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can quickly view the PRs in the stack using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show --prs
On stack learncommitstack-pt2
Branches in stack:
* learncommitstack-pt2 (top)
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/57

  learncommitstack
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/58

╭──────────────────────────────────────────────────╮
│                                                  │
│ Nice! All done part 1 of the tutorial. In part 2 │
│ we'll learn how to make more changes to a stack. │
│                                                  │
╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────╮
│                                                  │
│ Once you're ready, continue the tutorial using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack learn --part 2
TODO
```

## References

`exec.go` is heavily inspired by https://github.com/aviator-co/av (MIT license)
