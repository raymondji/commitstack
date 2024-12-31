# Commitstack

CLI to facilitate [stacking workflows](https://www.stacking.dev/) in git. Support for Gitlab (!) and Github.

## Why another stacking tool?

There's a plethora of stacking tools that work great with Github (I thoroughly enjoyed using https://github.com/spacedentist/spr), but after I started working at a company that uses Gitlab, I was surprised to find almost no options.

The only one I came across is the experimental `glab stack` CLI command, but it's missing a few features I personally want and has the opposite problem: it's Gitlab only. I ideally wanted a single tool once that could work with both.

Thus, Commitstack was born. Originally a simple set of bash functions/aliases, it's now morphed into a more robust Go CLI with full support for Gitlab and Github. It should also (hopefully) be easily exensible to other Git hosting providers.

## Goals

Stacking workflows in Git can be cumbersome both because of your interactions with Git itself (e.g. keeping track of stacks, pushing multiple branches), and because of your interactions with Gitlab/Github/etc (e.g. opening PRs, setting target branches). Commitstack aims to make both of these aspects easier.

In addition, Commitstack aims to provide a user experience that feels like a natural extension of using Git. Git already has all the basic building blocks needed to do stacking workflows, and many people use stacking workflows without any extra tools. Rather than introduce entirely new paradigms, Commitstack aims to feel familiar if you're used to stacking with plain Git - just with the annoying parts automated.

## References

`exec.go` is heavily inspired by https://github.com/aviator-co/av (MIT license)
