# Git stack

A simple tool to facilitate [stacking workflows](https://www.stacking.dev/) in git.

[![CI](https://github.com/raymondji/git-stack/actions/workflows/ci.yml/badge.svg)](https://github.com/raymondji/git-stack/actions/workflows/ci.yml)

## Goals

1. Make stacking PRs in Git easier.
2. Provide first-class support for Gitlab (which most stacking tools don't support) and make it easy to extend to other Git providers.
3. Feel like a natural extension of using Git directly (no black magic, easy to understand how it interacts with other Git commands, tells you clearly when something else you've done in Git is incompatible with the tool, no additional state)

# How to use

1. Linear commit history in your local feature branches

2. Single branch per commit

## References

https://github.com/aviator-co/av

## Setup

Install the dependencies:
- Required: `git` (>= 2.38)
- (Optional) Gitlab extension: `glab`
- (Optional) Github extension: `gh`

Clone this repo somewhere, e.g. `~/dev/git-stack`
