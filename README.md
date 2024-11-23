# Git stacked

A simple tool to facilitate [stacking workflows](https://www.stacking.dev/) in git.

## Goals

- Make stacking in Git easy
- Provide first-class support for Gitlab (which most stacking tools don't suppor) and make it easy to extend to other Git providers.
- Feel like a natural extension of using Git directly (no black magic, easy to understand how it interacts with other Git commands, tells you clearly when something else you've done in Git is incompatible with the tool)

## Sample usage

[View sample usage](https://github.com/raymondji/git-stacked/blob/main/test-goldens/none.txt)

[View sample usage with the github extension](https://github.com/raymondji/git-stacked/blob/main/test-goldens/github.txt)

[View sample usage with the gitlab extension](https://github.com/raymondji/git-stacked/blob/main/test-goldens/gitlab.txt)

# Requirements

1. Linear commit history in your local feature branches

2. Single branch per commit

# Concepts

Status: WIP

A stack is a sequence of branches from $GS_BASE_BRANCH, excluding $GS_BASE_BRANCH itself (default: `main`), where the topmost branch contains 1+ commits not reachable from any oher branch.

## Design goals

The key guiding principle: start with how someone would go about stacking using just Git. Then improve the UX by automating the aspects that are cumbersome or tricky, without enforcing a whole new way of doing things.

- Works alongside the `git` CLI (does not try to replace it)
- Lean on standard git commands and built-in functionality (like `--update-refs`)
- Stateless (only uses what's already tracked in git)
- Easy to understand what `git-stacked` is doing (no magic)
- Simple code is a priority
- Core functionality works with any git service provider
- Optional, enhanced integration with gitlab and github
- Minimal dependencies to run
  
## Setup

Install the dependencies:
- Required: `bash` and `git` (>= 2.38)
- (Optional) Gitlab extension: `glab` and `jq`
- (Optional) Github extension: `gh` and `jq`

Clone this repo somewhere, e.g. `~/dev/git-stacked`

Configure your `~/.zshrc` or `~/.bashrc`:
```
# Optional:
# export GS_BASE_BRANCH="..."
# export GS_ENABLE_GITLAB_EXTENSION=true
# export GS_ENABLE_GITHUB_EXTENSION=true

source ~/dev/git-stacked/stack.sh
```

## Comparison with other tools

Good overview of the available tools: https://stacking.dev. Find one that speaks to you!

https://github.com/spacedentist/spr
- API is based on phabricator (one commit per PR)
- only works with Github
- requires Github personal access token

https://graphite.dev/
- paid SaaS
