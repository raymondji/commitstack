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
gitstack-create logging frontend
git add .
git commit -m "Add frontend logging"

gitstack-branch backend
git add .
git commit -m "Add backend logging"

gitstack-push

gitstack-rebase
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
