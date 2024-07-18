# gitstack

## Setup

Clone this repo somewhere, e.g. `~/dev/gitstack`

In your `~/.zshrc` or `~/.bashrc`:
```
# Optional:
# export GS_BASE_BRANCH="..."
# export GS_BRANCH_PREFIX="..."

source ~/dev/gitstack/gitstack.sh
```

## Usage

Sample usage:
```
> gitstack-create logging frontend
Switched to a new branch 'user/logging/frontend/TIP'

> gitstack-branch backend
Renamed branch 'user/logging/frontend/TIP' -> 'user/logging/frontend'
Switched to a new branch 'user/logging/backend/TIP'

> gitstack-push
Pushing branch: user/logging/backend/TIP
...

Pushing branch: user/logging/frontend
...

> gitstack-rebase

> gitstack-create helm prometheus
Switched to a new branch 'user/helm/prometheus/TIP'

> gitstack-list
helm
logging

> gitstack-checkout logging
Switched to branch 'user/logging/backend/TIP'

> gitstack-list-branches
user/logging/backed/TIP
user/logging/frontend
```

## Recommended `~/.gitconfig` settings

These settings are not required to use gitstack, but will be helpful for git operations you do outside of the gitstack commands.


```
[push]
    autoSetupRemote = true
    default = upstream

[rebase]
    updateRefs = true
```

`gitstack` makes significant use of renaming local branches, so `default = upstream` avoids errors when you try to re-push after renaming a local branch.

`updateRefs = true` means that rebasing will automatically update refs, without having to specify `--update-refs`.
