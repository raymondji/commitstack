# Commitstack

CLI to facilitate [stacking workflows](https://www.stacking.dev/) in git. Support for Gitlab (!) and Github.

## Why another stacking tool?

There's a plethora of stacking tools that work great with Github (I thoroughly enjoyed using https://github.com/spacedentist/spr), but after I started using Gitlab at work, I was surprised to find almost no tools that fit the bill.

The only one I came across is the experimental `glab stack` CLI command, but it's missing a few features I personally want and has the opposite problem: it's Gitlab only. I ideally wanted a single tool once that could work with both.

Thus, Commitstack was born. Originally a simple set of bash functions/aliases, it's now morphed into a more robust Go CLI with full support for Gitlab and Github. It should also (hopefully) be easily exensible to other Git hosting providers.

## Goals

Stacking workflows in Git can be cumbersome both because of your interactions with Git itself (e.g. keeping track of stacks, pushing multiple branches), and because of your interactions with Gitlab/Github/etc (e.g. opening PRs, setting target branches). Commitstack aims to make both of these aspects easier.

In addition, Commitstack aims to provide a user experience that feels like a natural extension of Git. Git already has all the basic building blocks needed to do stacking workflows, and many people use stacking workflows without any extra tools. I wanted to avoid introducing entirely new paradigms and instead make Commitstack feel familiar if you're used to stacking with plain Git - just with the annoying parts automated.

## Installation

Go version >= 1.22 is required. To install Go on macOS:
```
brew install go 
```

Install Commitstack:
```
go install github.com/raymondji/commitstack/cmd/git-stack@latest
```

## Getting started

Commitstack needs a personal access token in order to manage MRs on Gitlab/PRs on Github for you. To set this up:
```
cd ~/your/git/repo
git stack init
```

Now you're all set up! The CLI ships with an interactive tutorial that teaches you the basics, access it here:
```
git stack learn
```

## References

`exec.go` is heavily inspired by https://github.com/aviator-co/av (MIT license)
