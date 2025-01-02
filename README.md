# commitstack

CLI to facilitate [stacking workflows](https://www.stacking.dev/) in git. Support for Gitlab (!) and Github.

## Why another stacking tool?

Stacking with plain Git is totally doable but cumbersome. There's a plethora of tools that make stacking easier for Github, but after I started using Gitlab at work, I was surprised to find almost nothing that fit the bill.

The only tool I came across is the experimental `glab stack` CLI command. While it seems promising, it's missing a few features I wanted and has the opposite problem: it only supports Gitlab. Ideally I wanted a single tool that works with both.

Enter: commitstack. The project currently supports Gitlab and Github, and should (hopefully) be easily extensible to other Git hosting providers.

## Goals

Stacking workflows in Git can be cumbersome both because of your interactions with Git itself (e.g. pushing multiple branches and keeping track of your stacks), and because of your interactions with Gitlab/Github/etc. (e.g. opening PRs and setting target branches). commitstack aims to make both of these aspects easier.

In addition, commitstack tries to fit into Git as naturally as possible and lean on existing concepts/funcionality. Git already has the basic building blocks needed to do stacking workflows (e.g. with `git rebase --update-refs`), and many people stack without any extra tools. For everyone in that bucket, commitstack should (hopefully) feel like it fits right in (while removing much of the friction).

Lastly, as mentioned above, commitstack aims to provide full support for Gitlab and Github, and hopefully be easy to extend to other Git hosting providers.

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

To install commitstack:
```
go install github.com/raymondji/commitstack/cmd/git-stack@current
```

## Getting started

The commitstack binary is named `git-stack`. Git offers a handy trick allowing binaries named `git-<foo>` to be invoked as git subcommands, so commitstack can be invoked as `git stack`.

commitstack needs a Gitlab/Github personal access token in order to manage MRs/PRs for you. To set this up:
```
cd ~/your/git/repo
git stack init
```

To learn how to use commitstack, you can access an interactive tutorial built-in to the CLI:
```
git stack learn
```

## Sample usage

This sample output is taken from `git stack learn --chapter=1 --mode=exec`.

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
[myfirststack c538f6e] hello world
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
[myfirststack-pt2 c5a18db] break
 1 file changed, 1 insertion(+)
> echo 'have a kitkat' >> myfirststack.txt
> git commit -am 'kitkat'
[myfirststack-pt2 5475cd6] kitkat
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
* 5475cd6 (myfirststack-pt2) kitkat
  c5a18db break
  c538f6e (myfirststack) hello world
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can easily push all branches in the stack up  │
│ as separate PRs.                                 │
│ commitstack automatically sets the target        │
│ branches for you.                                │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack push
Pushed myfirststack-pt2: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/110
Pushed myfirststack: https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/109
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can quickly view the PRs in the stack using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show --prs
On stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/110

  myfirststack
  └── https://gitlab.com/raymondji/git-stacked-gitlab-test/-/merge_requests/109

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
[mysecondstack 0b070c1] My second stack
 1 file changed, 1 insertion(+)
 create mode 100644 mysecondstack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ To view all the stacks:                          │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack list
  myfirststack-pt2 (2 branches)
* mysecondstack (1 branches)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Nice! All done chapter 1 of the tutorial.        │
│ ...                                              │
│                                                  │
╰──────────────────────────────────────────────────╯
```

## How does it work?

When working with Git we tend to think in terms of branches as the unit of work, and Gitlab/Github both tie pull requests to branches. Thus, as much as possible, commitstack tries to frame stacks in terms of "stacks of branches".

However, branches in Git don't inherently make sense as belonging to a "stack", i.e. where one branch is stacked on top of another branch. Branches in Git are just pointers to commits, so:
- Multiple branches can point to the same commit
- Branches don't inherently have a notion of parent branches or children branches

Under the hood, commitstack therefore thinks about stacks as "stacks of commits", not "stacks of branches". Hence the name. :) Commits serve this purpose much better than branches because:
- Each commit is a unique entity
- Commits do inherently have a notion of parent commits and children commits

Merge commits still pose a problem. It's easy to reason about a linear series of commits as a stack, but merge commits have multiple parents. So, commitstack takes the simple option of being incompatible with merge commits. If it encounters a merge commit, it will print an error message and not try to interpret it as a stack.

## References

`exec.go` is heavily inspired by https://github.com/aviator-co/av (MIT license)
