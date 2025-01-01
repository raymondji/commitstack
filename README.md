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

The commitstack binary is named `git-stack`. Git offers a handy built-in trick allowing binaries named `git-<foo>` to be invoked as git subcommands, so commitstack can be invoked as `git stack`.

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
