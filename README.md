# git stack

A minimal Git CLI subcommand for managing stacked branches/pull requests. Works with Gitlab and Github.

Core commands:
- `git stack list`: list all stacks
- `git stack show`: view your current stack
- `git stack push`: push branches in the current stack and open MRs/PRs

## What is stacking?

https://graphite.dev/guides/stacked-diffs has a good overview on what it is and why you might want to do it.

## Where does native Git fall short with stacking?

Stacking branches natively with Git is completely doable, but cumbersome.
- While modern Git has made updating stacked branches much easier with [`--update-refs`](https://andrewlock.net/working-with-stacked-branches-in-git-is-easier-with-update-refs/), other tasks like keeping track of your stacks or pushing all branches in a stack are left to the user.
- Moreover, stacking also typically involves additional manual steps on Gitlab/Github/etc, such as setting the correct target branch on each pull request.

## How does `git stack` compare to `<other stacking tool>`?

There are two main areas where `git stack` differs from most existing tools:
- `git stack` is designed to feel like a minimal addition to the Git CLI. It works with existing Git concepts and functionality (like `--update-refs`), and aims to unintrusively fill in the gaps. Unlike most stacking tools, it's also stateless, so there's no state to keep in sync between Git and `git stack`. Instead, it works by automatically inferring stacks from the structure of your commits.
- `git stack` integrates with Gitlab (and Github). I was surprised to find most of the [popular](https://graphite.dev/) [stacking](https://github.com/aviator-co/av) [tools](https://github.com/gitbutlerapp/gitbutler) only support Github. Besides `git stack`, some other projects I've found with Gitlab support include [git-town](https://github.com/git-town/git-town), [git-spice](https://github.com/abhinav/git-spice) and the new [`glab stack`](https://docs.gitlab.com/ee/user/project/merge_requests/stacked_diffs.html) CLI command. They all work pretty differently and have different feature sets.

## Limitations

- `git stack` requires maintaining linear commit histories in your feature branches to be able to infer stacks. Thus it's effectively tied to using `git rebase`, which seemed reasonable given that `git rebase --update-refs` is the native way of updating stacked branches in Git. However, this means `git stack` is not compatible with `git merge` workflows (at least within feature branches, merging into `main` is no problem).

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

To install `git stack`:
```
go install github.com/raymondji/git-stack-cli/cmd/git-stack@0.19.0
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
│ Let&#39;s start things off on the default branch:    │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git checkout main
Your branch is ahead of &#39;origin/main&#39; by 5 commits.
  (use &#34;git push&#34; to publish your local commits)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Next, let&#39;s create our first branch:             │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git checkout -b myfirststack
&gt; echo &#39;hello world&#39; &gt; myfirststack.txt
&gt; git add .
&gt; git commit -m &#39;hello world&#39;
[myfirststack 67e46e0] hello world
 1 file changed, 1 insertion(&#43;)
 create mode 100644 myfirststack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ Now let&#39;s stack a second branch on top of our    │
│ first:                                           │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git checkout -b myfirststack-pt2
&gt; echo &#39;have a break&#39; &gt;&gt; myfirststack.txt
&gt; git commit -am &#39;break&#39;
[myfirststack-pt2 befdc2b] break
 1 file changed, 1 insertion(&#43;)
&gt; echo &#39;have a kitkat&#39; &gt;&gt; myfirststack.txt
&gt; git commit -am &#39;kitkat&#39;
[myfirststack-pt2 8e97c25] kitkat
 1 file changed, 1 insertion(&#43;)
╭──────────────────────────────────────────────────╮
│                                                  │
│ So far we&#39;ve only used standard Git commands.    │
│ Let&#39;s see what git stack can do for us already.  │
│                                                  │
│ Our current stack has two branches in it, which  │
│ we can see with:                                 │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git stack show
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
&gt; git stack show --log
In stack myfirststack-pt2
Commits in stack:
* 8e97c25 (HEAD -&gt; myfirststack-pt2) kitkat (top)
  befdc2b break      
  67e46e0 (myfirststack) hello world      
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can easily push all branches in the stack up  │
│ as separate PRs.                                 │
│ git stack automatically sets the target branches │
│ for you.                                         │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git stack push
Pushed myfirststack-pt2: https://github.com/raymondji/git-stack-cli/pull/82
Pushed myfirststack: https://github.com/raymondji/git-stack-cli/pull/81
╭──────────────────────────────────────────────────╮
│                                                  │
│ We can quickly view the PRs in the stack using:  │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git stack show --prs
In stack myfirststack-pt2
Branches in stack:
* myfirststack-pt2 (top)
  └── https://github.com/raymondji/git-stack-cli/pull/82

  myfirststack
  └── https://github.com/raymondji/git-stack-cli/pull/81
╭──────────────────────────────────────────────────╮
│                                                  │
│ To sync the latest changes from the default      │
│ branch into the stack, you can run:              │
│ git rebase main --update-refs                    │
│ Or to avoid having to remember --update-refs,    │
│ you can do:                                      │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git stack rebase main
Successfully rebased myfirststack-pt2 on main
╭──────────────────────────────────────────────────╮
│                                                  │
│ Great, we&#39;ve got the basics down for one stack.  │
│ How do we deal with multiple stacks?             │
│ Let&#39;s head back to our default branch and create │
│ a second stack.                                  │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git checkout main
Your branch is ahead of &#39;origin/main&#39; by 5 commits.
  (use &#34;git push&#34; to publish your local commits)
&gt; git checkout -b mysecondstack
&gt; echo &#39;buy one get one free&#39; &gt; mysecondstack.txt
&gt; git add .
&gt; git commit -m &#39;My second stack&#39;
[mysecondstack 7e4c183] My second stack
 1 file changed, 1 insertion(&#43;)
 create mode 100644 mysecondstack.txt
╭──────────────────────────────────────────────────╮
│                                                  │
│ To view all the stacks:                          │
│                                                  │
╰──────────────────────────────────────────────────╯
&gt; git stack list
  dev (1 branch)
  myfirststack-pt2 (2 branches)
* mysecondstack (1 branch)
╭──────────────────────────────────────────────────╮
│                                                  │
│ Nice! All done chapter 1 of the tutorial.        │
│                                                  │
│ In chapter 2 we&#39;ll see how to make changes to    │
│ earlier branches in the stack.                   │
│ Once you&#39;re ready, continue the tutorial using:  │
│ git stack learn --chapter 2                      │
│                                                  │
│ To cleanup all the branches/PRs that were        │
│ created, run:                                    │
│ git stack learn --chapter 1 --cleanup            │
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
- `exec.go` is adapted from https://github.com/aviator-co/av/blob/fbcb5bfc0f19c8a7924e309cb1e86678a9761daa/internal/git/git.go#L178
