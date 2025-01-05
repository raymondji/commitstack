<!--- DO NOT EDIT: this file is generated --->

# git stack

A minimal Git CLI subcommand for managing stacked branches/MRs/PRs. Works with Gitlab and Github.

Core commands:
- `git stack list`: list all stacks
- `git stack show`: view your current stack
- `git stack push`: push branches in the current stack and open MRs/PRs

## What is stacking?

https://graphite.dev/blog/stacked-prs has a good overview on what it is and why you might want to do it.

## Where does native Git fall short with stacking?

Stacking branches natively with Git is completely doable, but cumbersome.
- While modern Git has made updating stacked branches much easier with [`--update-refs`](https://andrewlock.net/working-with-stacked-branches-in-git-is-easier-with-update-refs/), other tasks like keeping track of your stacks or pushing all branches in a stack are left to the user.
- Moreover, stacking also typically involves additional manual steps on Gitlab/Github/etc, such as setting the correct target branch on each pull request.

## How does `git stack` compare to `<other stacking tool>`?

There are two main areas where `git stack` differs from most existing tools:
- `git stack` functions as a minimal Git CLI extension and is designed to feel as native as possible. It works with existing Git concepts and functionality (like `--update-refs`), and unlike most stacking tools, it's stateless. That means there's no state to keep in sync between Git and `git stack`. Instead, `git stack` works by inferring stacks automatically from the structure of your commits.
- `git stack` integrates with both Gitlab and Github. I was surprised to find most of the [popular](https://graphite.dev/) [stacking](https://github.com/aviator-co/av) [tools](https://github.com/gitbutlerapp/gitbutler) only support Github. Besides `git stack`, other options with Gitlab support include [git-town](https://github.com/git-town/git-town), [git-spice](https://github.com/abhinav/git-spice) and the new [`glab stack`](https://docs.gitlab.com/ee/user/project/merge_requests/stacked_diffs.html) CLI command.

## Limitations

- `git stack` requires linear commit histories in feature branches in order to infer stacks, making it effectively tied to `git rebase`. `git rebase --update-refs` is the native way of updating stacked branches in Git, so this approach seems well aligned. However, this means `git stack` is not compatible with `git merge` workflows within feature branches (note: merging into `main` is no problem).

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

To install `git stack`:
```
go install github.com/raymondji/git-stack-cli/cmd/git-stack@0.28.0
```

## Getting started

The `git stack` binary is named `git-stack`. Git offers a handy trick allowing binaries named `git-<foo>` to be invoked as git subcommands, so `git stack` can be invoked as `git stack`.

`git stack` needs a Gitlab/Github personal access token in order to manage MRs/PRs for you. To set this up:
```
cd ~/your/git/repo
git stack init
```

To learn how to use `git stack`, you can access an interactive tutorial built-in to the CLI:
```
git stack learn
```

## Sample usage

This sample output is taken from `git stack learn --chapter=1 --mode=exec`.

```
╭──────────────────────────────────────────────────╮
│                                                  │
│ Welcome to git stack!                            │
│ Here is a quick tutorial on how to use the CLI.  │
│                                                  │
╰──────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────╮
│                                                  │
│ Let's start things off on the default branch:    │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout main
Your branch is ahead of 'origin/main' by 1 commit.
  (use "git push" to publish your local commits)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Next, let's create our first branch:             │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout -b myfirststack
> echo 'hello world' > myfirststack.txt
> git add .
> git commit -m 'hello world'
[myfirststack ef6a33f] hello world
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
[myfirststack-pt2 6da4354] break
 1 file changed, 1 insertion(+)
> echo 'have a kitkat' >> myfirststack.txt
> git commit -am 'kitkat'
[myfirststack-pt2 efadf0b] kitkat
 1 file changed, 1 insertion(+)
╭──────────────────────────────────────────────────╮
│                                                  │
│ So far we've only used standard Git commands.    │
│ Let's see what git stack can do for us already.  │
│                                                  │
│ Our current stack has two branches in it, which  │
│ we can see with:                                 │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show
In stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  myfirststack
╭──────────────────────────────────────────────────╮
│                                                  │
│ Our current stack has 3 commits in it, which we  │
│ can see with:                                    │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show --log
In stack myfirststack-pt2
Commits in stack:
* efadf0b (HEAD -> myfirststack-pt2) kitkat (top)
  6da4354 break      
  ef6a33f (myfirststack) hello world      
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can easily push all branches in the stack up  │
│ as separate PRs.                                 │
│ git stack automatically sets the target branches │
│ for you.                                         │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack push
Pushed myfirststack-pt2: https://github.com/raymondji/git-stack-cli/pull/126
Pushed myfirststack: https://github.com/raymondji/git-stack-cli/pull/127
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can quickly view the PRs in the stack using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack show --prs
In stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  └── https://github.com/raymondji/git-stack-cli/pull/126

  myfirststack
  └── https://github.com/raymondji/git-stack-cli/pull/127
╭──────────────────────────────────────────────────╮
│                                                  │
│ To sync the latest changes from the default      │
│ branch into the stack, you can run:              │
│ git rebase main --update-refs                    │
│ Or to avoid having to remember --update-refs,    │
│ you can do:                                      │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack rebase main
Successfully rebased myfirststack-pt2 on main
╭──────────────────────────────────────────────────╮
│                                                  │
│ Great, we've got the basics down for one stack.  │
│ How do we deal with multiple stacks?             │
│ Let's head back to our default branch and create │
│ a second stack.                                  │
│                                                  │
╰──────────────────────────────────────────────────╯
> git checkout main
Your branch is ahead of 'origin/main' by 1 commit.
  (use "git push" to publish your local commits)
> git checkout -b mysecondstack
> echo 'buy one get one free' > mysecondstack.txt
> git add .
> git commit -m 'My second stack'
[mysecondstack a2db832] My second stack
 1 file changed, 1 insertion(+)
 create mode 100644 mysecondstack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ To view all the stacks:                          │
│                                                  │
╰──────────────────────────────────────────────────╯
> git stack list
  dev (1 branch)
  myfirststack-pt2 (2 branches)
* mysecondstack (1 branch)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Nice! All done chapter 1 of the tutorial.        │
│                                                  │
│ In chapter 2 we'll see how to make changes to    │
│ earlier branches in the stack.                   │
│ Once you're ready, continue the tutorial using:  │
│ git stack learn --chapter 2                      │
│                                                  │
│ To cleanup all the branches/PRs that were        │
│ created, run:                                    │
│ git stack learn --chapter 1 --mode=clean         │
│                                                  │
╰──────────────────────────────────────────────────╯

```

## How does it work?

When working with Git we often think in terms of branches as the unit of work, and Gitlab/Github both tie pull requests to branches. Thus, as much as possible, `git stack` tries to frame stacks in terms of "stacks of branches".

However, branches in Git don't inherently make sense as belonging to a "stack", i.e. where one branch is stacked on top of another branch. Branches in Git are just pointers to commits, so:
- Multiple branches can point to the same commit
- Branches don't inherently have a notion of parent branches or children branches

Under the hood, `git stack` therefore thinks about stacks as "stacks of commits", not "stacks of branches". Commits serve this purpose much better than branches because:
- Each commit is a unique entity
- Commits do inherently have a notion of parent commits and children commits

Merge commits still pose a problem. It's easy to reason about a linear series of commits as a stack, but merge commits have multiple parents. So, `git stack` takes the simple option of being incompatible with merge commits. If it encounters a merge commit, it will print an error message and otherwise ignore the commit.

`git stack` doesn't persist any additional state to keep track of your stacks - it relies purely on parsing the structure of your commits to infer which commits form a stack (and in turn which branches form a stack). If `git stack` encounters a structure it can't parse (e.g. merge commits), it tries to print a helpful error and otherwise ignores the incompatible commit(s).

## Attribution

Some code is adapted from sections of https://github.com/aviator-co/av (MIT license). A copy of av's license is included at `attribution/aviator-co/av/LICENSE`.
- `exec.go` is adapted from [aviator-co/av/internal/git/git.go](https://github.com/aviator-co/av/blob/fbcb5bfc0f19c8a7924e309cb1e86678a9761daa/internal/git/git.go#L178)
