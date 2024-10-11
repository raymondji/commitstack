# Git stacked

A simple tool to facilitate [stacking workflows](https://www.stacking.dev/) in git.

## Sample usage

[View](https://github.com/raymondji/git-stacked/blob/main/test-output-golden.txt)

## Design goals

- Works alongside the `git` CLI (does not try to replace it)
- Stateless (only uses what's already tracked in git)
- Lean on standard git commands and built-in functionality (like `--update-refs`)
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

source ~/dev/git-stacked/stack.sh
```

## Recommended `~/.gitconfig` settings

These settings are not required to use `git-stacked`, but are helpful for git operations you do outside of the `git-stacked` commands.

```
[rebase]
    updateRefs = true
```

`updateRefs = true` means that rebasing will automatically update refs, without having to specify `--update-refs`.

## Comparison with other tools

Good overview of the available tools: https://stacking.dev. Find one that speaks to you!

https://github.com/spacedentist/spr
- API is based on phabricator (one commit per PR)
- only works with Github
- requires Github personal access token

https://graphite.dev/
- paid SaaS
