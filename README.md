# Commitstack

CLI to facilitate [stacking pull requests](https://www.stacking.dev/) in git. Integrates with Gitlab and Github.

## What is stacking?

https://graphite.dev/guides/stacked-diffs has a good overview on what it is and why you might want to do it.

## Why use an additional tool for stacking?

Stacking natively with Git is totally doable, but cumbersome.
- While modern Git has made updating stacked branches much easier with [`--update-refs`](https://andrewlock.net/working-with-stacked-branches-in-git-is-easier-with-update-refs/), other tasks like keeping track of your stacks or pushing all branches in a stack are left to the user.
- Moreover, stacking also typically involves additional manual steps on Gitlab/Github/etc, such as setting the correct target branch on each pull request.

## How does Commitstack compare to `<other stacking tool>`?

All the tools offer different trade-offs.

One reason you might want to use Commitstack is if you need Gitlab support. While Github users are [spoiled](https://github.com/aviator-co/av) [for](https://graphite.dev/) [choice](https://github.com/spacedentist/spr), most of the popular options don't support Gitlab.
- If you want a stacking tool that integrates with both Gitlab and Github, Commitstack is the one of the few options I'm aware of. Some others include [git-town](https://github.com/git-town/git-town) and [git-spice](https://github.com/abhinav/git-spice).
- If you need Gitlab support only, the most prominent option seems to be the [`glab stack`](https://docs.gitlab.com/ee/user/project/merge_requests/stacked_diffs.html) CLI command.
- These options are all pretty different in terms of the user experience and feature set they provide; see which one fits your needs the best. 

Another reason is that if you already stack branches natively using Git and/or know your way around `git rebase`, you might really like the user experience of Commitstack. It's designed to lean on existing Git concepts and functionality as much as possible, and provide a minimal set of abstractions on top. The goal is to make Commitstack easy to learn and integrate into your existing Git workflow.

On the flipside, here's one reason Commitstack might not be right for you. Commitstack depends heavily on `git rebase`, so if you don't enjoy `git rebase` workflows, there probably exists a better tool for you.

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

To install Commitstack:
```
go install github.com/raymondji/commitstack/cmd/git-stack@latest
```

## Getting started

The Commitstack binary is named `git-stack`. Git offers a handy trick allowing binaries named `git-<foo>` to be invoked as git subcommands, so Commitstack can be invoked as `git stack`.

Commitstack needs a Gitlab/Github personal access token in order to manage MRs/PRs for you. To set this up:
```
cd ~/your/git/repo
git stack init
```

To learn how to use Commitstack, you can access an interactive tutorial built-in to the CLI:
```
git stack learn
```

## Sample usage

This sample output is taken from `git stack learn --chapter=1 --mode=exec`.

```
╭──────────────────────────────────────────────────╮
│                                                  │
│ Welcome to Commitstack!                          │
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
│ Git. Let's see what Commitstack can do for us    │
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
│ Commitstack automatically sets the target        │
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

When working with Git we often think in terms of branches as the unit of work, and Gitlab/Github both tie pull requests to branches. Thus, as much as possible, Commitstack tries to frame stacks in terms of "stacks of branches".

However, branches in Git don't inherently make sense as belonging to a "stack", i.e. where one branch is stacked on top of another branch. Branches in Git are just pointers to commits, so:
- Multiple branches can point to the same commit
- Branches don't inherently have a notion of parent branches or children branches

Under the hood, Commitstack therefore thinks about stacks as "stacks of commits", not "stacks of branches". Hence the name. :) Commits serve this purpose much better than branches because:
- Each commit is a unique entity
- Commits do inherently have a notion of parent commits and children commits

Merge commits still pose a problem. It's easy to reason about a linear series of commits as a stack, but merge commits have multiple parents. So, Commitstack takes the simple option of being incompatible with merge commits. If it encounters a merge commit, it will print an error message and otherwise ignore the commit.

Commitstack doesn't persist any additional state to keep track of your stacks - it relies purely on parsing the structure of your commits to infer which commits form a stack (and in turn which branches form a stack). If Commitstack encounters a structure it can't parse (e.g. merge commits), it tries to print a helpful error and otherwise ignores the incompatible commit(s).

## Attribution

Some code is adapted from sections of https://github.com/aviator-co/av (MIT license). A copy of av's license is included at `attribution/aviator-co/av/LICENSE`.
- `exec.go` is adapted from https://github.com/aviator-co/av/blob/fbcb5bfc0f19c8a7924e309cb1e86678a9761daa/internal/git/git.go#L178
