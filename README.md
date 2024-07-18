# qstack: simple stacked diffs in git

The simplest* tool for stacking git branches.

*In my opinion :)

## Design goals

- Works alongside the git commands you already know, does not try to replace the whole git CLI
- All state is tracked within git and easy for humans to understand and edit manually when needed
- Lean on built-in git functionality (like `--update-refs`) rather than implementing complex custom logic
- No dependencies besides bash and git
- Works with github, gitlab, or any other service provider

## Sample usage

```
> qstack create logging frontend
Switched to a new branch 'user/logging/frontend/TIP'

> qstack branch backend
Renamed branch 'user/logging/frontend/TIP' -> 'user/logging/frontend'
Switched to a new branch 'user/logging/backend/TIP'

> qstack push
Pushing branch: user/logging/backend/TIP -> user/logging/backend
...

Pushing branch: user/logging/frontend -> user/logging/frontend
...

> qstack rebase
# starts interactive rebase

> qstack create helm prometheus
Switched to a new branch 'user/helm/prometheus/TIP'

> qstack list
helm
logging

> qstack switch logging
Switched to branch 'user/logging/backend/TIP'

> qstack list-branches
user/logging/backed/TIP
user/logging/frontend
```

## Setup

Clone this repo somewhere, e.g. `~/dev/qstack`

In your `~/.zshrc` or `~/.bashrc`:
```
# Optional:
# export GS_BASE_BRANCH="..."
# export GS_BRANCH_PREFIX="..."

source ~/dev/qstack/qstack.sh
``` 

## Recommended `~/.gitconfig` settings

These settings are not required to use `qstack`, but will be helpful for git operations you do outside of the `qstack` commands.

```
[push]
    autoSetupRemote = true
    default = upstream

[rebase]
    updateRefs = true
```

`qstack` makes significant use of renaming local branches, so `default = upstream` avoids errors when you try to re-push after renaming a local branch.

`updateRefs = true` means that rebasing will automatically update refs, without having to specify `--update-refs`.

## Naming

Inspired by one of my favourite League of Legends characters, Nasus :)

## Comparison with other tools

https://github.com/spacedentist/spr
- only works with Github
- requires Github personal access token
- API is based on phabricator (strictly one commit per PR)

https://graphite.dev/
- paid SaaS
