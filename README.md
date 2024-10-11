# Git stacked

A simple tool to facilitate [stacking workflows](https://www.stacking.dev/) in git.

## Sample usage

```
> git-stacked create a1

> git-stacked create a2

> git-stacked branch
* a2 (top)
  a1
  main

> git-stacked log
commit ere821123 (HEAD -> a2)
Author: ...
Date:   ...

    Start of a2

commit jpo12323 (a1)
Author: Raymond Ji <34181040+raymondji@users.noreply.github.com>
Date:   Fri Oct 11 05:42:52 2024 -0700

    Start of a1

commit ka123129 (main)
Author: ...
Date:   ...

    Hello world

> git checkout main

> git-stacked create b1

> git-stacked stack
* b1
  a2

> git checkout a2

> git-stacked push
Push branch: a1
----------------
...

Push branch: a2
----------------
....
```

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
