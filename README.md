# gitstack

## Setup

## Recommended `~/.gitconfig` settings

```
[push]
    autoSetupRemote = true
    default = upstream
[rebase]
    updateRefs = true
```

`gitstack` makes significant use of renaming local branches, so `default = upstream` avoids errors when you try to re-push after renaming the local branch.

`updateRefs = true` means that rebasing will automatically update refs, without having to specify `--update-refs`.
