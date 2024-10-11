#!/bin/bash
GS_BASE_BRANCH=${GS_BASE_BRANCH:-main}

gs() {
    git-stacked "$@"
}

git-stacked() {
    if [ $# -eq 0 ]; then
        echo "Must provide command"
        return 1
    fi

    COMMAND=$1
    shift

    if [ "$COMMAND" = "help" ] || [ "$COMMAND" = "h" ]; then
        git-stacked-help "$@"
    elif [ "$COMMAND" = "push-force" ] || [ "$COMMAND" = "pf" ]; then
        git-stacked-push-force "$@"
    elif [ "$COMMAND" = "pull-rebase" ] || [ "$COMMAND" = "pr" ]; then
        git-stacked-pull-rebase "$@"
    elif [ "$COMMAND" = "rebase" ] || [ "$COMMAND" = "r" ]; then
        git-stacked-rebase "$@"
    elif [ "$COMMAND" = "branch" ] || [ "$COMMAND" = "b" ]; then
        git-stacked-branch "$@"
    elif [ "$COMMAND" = "stack" ] || [ "$COMMAND" = "s" ]; then
        git-stacked-stack "$@"
    elif [ "$COMMAND" = "log" ] || [ "$COMMAND" = "l" ]; then
        git-stacked-log "$@"
    elif [ "$COMMAND" = "reorder" ] || [ "$COMMAND" = "ro" ]; then
        git-stacked-reorder "$@"
    else
       echo "Invalid command"
    fi
}

git-stacked-help() {
    echo 'usage: git-stacked ${subcommand} ...
    alias: gs

subcommands:

push-force
    alias: pf
    push all branches in the current stack to remote

pull-rebase
    alias: pr
    update the base branch from mainstream, then rebase the current stack onto the base branch

rebase
    alias: r
    start interactive rebase of the current stack against the base branch

log
    alias: l
    git log helper

stack
    alias: s
    list all stacks

branch
    alias: b
    list all branches in the current stack

reorder
    alias: ro
    start interactive rebase to reorder branches in the current stack'
}

git-stacked-branch() {
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    BRANCHES=$(git log --pretty='format:%D' $GS_BASE_BRANCH.. --decorate-refs=refs/heads | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "No branches in the current stack"
        return 1
    fi

    echo "Branches in the current stack:"
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        # Check if this branch is the current branch
        if [ "$BRANCH" = "$CURRENT_BRANCH" ]; then
            echo "* \033[0;32m$BRANCH\033[0m (top)"
        else
            echo "  $BRANCH"
        fi
    done
}

git-stacked-stack() {
    CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    BRANCHES=$(git branch --format='%(refname:short)')
    STACKS=()
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        HAS_DESCENDENT=false

        echo "$BRANCHES" | while IFS= read -r MAYBE_DESCENDENT; do
            IS_ANCESTOR=$(git merge-base --is-ancestor $BRANCH $MAYBE_DESCENDENT^; echo $?)
            if [[ $BRANCH != $MAYBE_DESCENDENT ]] && [[ $IS_ANCESTOR == "0" ]]; then
                HAS_DESCENDENT=true
                break
            fi
        done

        if [[ $HAS_DESCENDENT == false ]]; then
            STACKS+=($BRANCH)
        fi
    done

    echo "Stacks:"
    for STACK in "${STACKS[@]}"; do
        # Check if this stack is the current stack
        if [ "$STACK" = "$CURRENT_BRANCH" ]; then
            echo -e "* \033[0;32m$STACK\033[0m" # green highlight
        else
            echo "  $STACK"
        fi
    done
}

git-stacked-log() {
    git log $QS_BASE_BRANCH..
}

git-stacked-push-force() {
    # Reverse so we push from bottom -> top
    BRANCHES=$(git log --pretty='format:%D' $QS_BASE_BRANCH.. --decorate-refs=refs/heads --reverse | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "No branches in the current stack"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo "Force pushing branch $BRANCH"
        echo "----------------------------"
        git push origin "$BRANCH":"$BRANCH" --force
        echo "" # newline
    done
}

git-stacked-pull-rebase() {
    git checkout $QS_BASE_BRANCH && \
    git pull && \
    git checkout - && \
    git rebase -i $QS_BASE_BRANCH --update-refs
}

git-stacked-rebase() {
    git rebase -i $QS_BASE_BRANCH --update-refs --keep-base
}

git-stacked-reorder() {
    echo "Please make sure to update target branches of all merge requests in this stack first"
    read -p "Acknowledge with Y to continue: " input
    if [[ "$input" == "Y" || "$input" == "y" ]]; then
       echo "Proceeding..."
    else
        echo "Exiting..."
        exit 1
    fi

    git checkout -b tmp-reorder-branch && \
    git rebase -i $QS_BASE_BRANCH --update-refs --keep-base && \
    git checkout - && \
    git branch -D tmp-reorder-branch
}
