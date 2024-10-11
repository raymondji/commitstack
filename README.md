# Git stacked

A simple tool to facilitate [stacking workflows](https://www.stacking.dev/) in git.

## Design goals

- Works alongside the `git` CLI (does not try to replace it)
- Stateless (only uses what's already tracked in git)
- Lean on standard git commands and built-in functionality (like `--update-refs`)
- Easy to understand what `git-stacked` is doing (no magic)
- Simple code is a priority
- Core functionality works with any git service provider
- Optional extensions improve the experience for gitlab and github
- Minimal dependencies to run

## Setup

Install the dependencies you want to use:
- Required: bash and git
- Gitlab extension: glab and jq
- Github extension: gh and jq

Clone this repo somewhere, e.g. `~/dev/git-stacked`

In your `~/.zshrc` or `~/.bashrc`:
```
# Optional:
# export GS_BASE_BRANCH="..."

source ~/dev/git-stacked/git-stacked.sh
```

## Recommended `~/.gitconfig` settings

These settings are not required to use `git-stacked`, but will be helpful for git operations you do outside of the `git-stacked` commands.

```
[push]
    autoSetupRemote = true
    default = upstream

[rebase]
    updateRefs = true
```

`git-stacked` makes significant use of renaming local branches, so `default = upstream` avoids errors when you try to re-push after renaming a local branch.

`updateRefs = true` means that rebasing will automatically update refs, without having to specify `--update-refs`.

## Comparison with other tools

Good overview of the available tools: https://stacking.dev. Find one that speaks to you!

https://github.com/spacedentist/spr
- API is based on phabricator (one commit per PR)
- only works with Github
- requires Github personal access token

https://graphite.dev/
- paid SaaS
