# commitstack

CLI to facilitate [stacking workflows](https://www.stacking.dev/) in git. Support for Gitlab (!) and Github.

## Why another stacking tool?

Stacking with plain Git is totally doable but cumbersome. There's a plethora of tools that make stacking easier for Github, but after I started using Gitlab at work, I was surprised to find almost nothing that fit the bill.

The only tool I came across is the experimental `glab stack` CLI command, but it's missing a few features I wanted and has the opposite problem: it's Gitlab only. I ideally wanted a single tool I could use with both.

Enter: commitstack. The project currently supports Gitlab and Github, and should (hopefully) be easily extensible to other Git hosting providers.

## Goals

Stacking workflows in Git can be cumbersome both because of your interactions with Git itself (e.g. pushing multiple branches and keeping track of your stacks), and because of your interactions with Gitlab/Github/etc. (e.g. opening PRs and setting target branches). commitstack aims to make both of these aspects easier.

In addition, commitstack tries to fit into Git as naturally as possible and avoid introducing whole new paradigms. Git already has the basic building blocks needed to do stacking workflows (e.g. with `git rebase --update-refs`), and many people stack without any extra tools. For everyone in that bucket, commitstack should (hopefully) feel like it fits right in (while removing much of the friction).

## Sample usage (output from `git stack learn`)

This sample output is generated entirely from `git stack learn --mode=exec`, an interactive tutorial built-in to the CLI.

```
╭──────────────────────────────────────────────────╮
│                                                  │
│ Welcome to commitstack!                          │
│ Here is a quick tutorial on how to use the CLI.  │
│                                                  │
╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────╮
│                                                  │
│ Let's start things off on the default branch:    │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout main
Your branch is up to date with 'origin/main'.
╭──────────────────────────────────────────────────╮
│                                                  │
│ Next, let's create our first branch:             │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout -b myfirststack
> echo 'hello world' > myfirststack.txt
> git add .
> git commit -m 'hello world'
[myfirststack 37d2115] hello world
 1 file changed, 1 insertion(+)
 create mode 100644 myfirststack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ Now let's stack a second branch on top of our    │
│ first:                                           │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout -b myfirststack-pt2
> echo 'have a break' >> myfirststack.txt
> git commit -am 'break'
[myfirststack-pt2 31525c9] break
 1 file changed, 1 insertion(+)
> echo 'have a kitkat' >> myfirststack.txt
> git commit -am 'kitkat'
[myfirststack-pt2 fd78a1f] kitkat
 1 file changed, 1 insertion(+)
╭──────────────────────────────────────────────────╮
│                                                  │
│ So far everything we've done has been normal     │
│ Git. Let's see what commitstack can do for us    │
│ already.                                         │
│                                                  │
│ Our current stack has two branches in it, which  │
│ we can see with:                                 │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show
On stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  myfirststack
╭──────────────────────────────────────────────────╮
│                                                  │
│ Our current stack has 3 commits in it, which we  │
│ can see with:                                    │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack log
* fd78a1f (myfirststack-pt2) kitkat
  31525c9 break
  37d2115 (myfirststack) hello world
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can easily push all branches in the stack up  │
│ as separate PRs.                                 │
│ commitstack automatically sets the target        │
│ branches for you.                                │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack push
Pushed myfirststack-pt2: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/100
Pushed myfirststack: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/99
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can quickly view the PRs in the stack using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show --prs
On stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/100

  myfirststack
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/99

╭──────────────────────────────────────────────────╮
│                                                  │
│ To pull in the latest changes from the default   │
│ branch into the stack, run:                      │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack pull
Pulling from main into the current stack myfirststack-pt2
Current branch myfirststack-pt2 is up to date.
╭──────────────────────────────────────────────────╮
│                                                  │
│ Great, we're getting the hang of this!           │
│ One stack is nice, but how do we deal with       │
│ multiple stacks?                                 │
│ Let's head back to our default branch and create │
│ a second stack.                                  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout main
Your branch is up to date with 'origin/main'.
> git checkout -b mysecondstack
> echo 'buy one get one free' > mysecondstack.txt
> git add .
> git commit -m 'My second stack'
[mysecondstack 363f597] My second stack
 1 file changed, 1 insertion(+)
 create mode 100644 mysecondstack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ To view all the stacks:                          │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack list
  B1 (1 branches)
  foo (1 branches)
  myfirststack-pt2 (2 branches)
* mysecondstack (1 branches)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Nice! All done part 1 of the tutorial.           │
│                                                  │
│ In part 2 we'll see how to make changes to       │
│ earlier branches in the stack.                   │
│ Once you're ready, continue the tutorial using:  │
│ git stack learn --part 2                         │
│                                                  │
│ To cleanup all the branches/PRs that were        │
│ created, run:                                    │
│ git stack learn --cleanup                        │
│                                                  │
╰──────────────────────────────────────────────────╯
```

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

## References

`exec.go` is heavily inspired by https://github.com/aviator-co/av (MIT license)
