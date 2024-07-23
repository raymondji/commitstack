#!/bin/bash
QS_BASE_BRANCH=${QS_BASE_BRANCH:-main}

qs() {
    qstack "$@"
}

qstack() {
    if [ $# -eq 0 ]; then
        echo "Must provide command"
        return 1
    fi

    COMMAND=$1
    shift

    if [ "$COMMAND" = "help" ] || [ "$COMMAND" = "h" ]; then
        qstack-help "$@"
    elif [ "$COMMAND" = "push" ]; then
        qstack-push "$@"
    elif [ "$COMMAND" = "pull" ]; then
        qstack-pull "$@"
    elif [ "$COMMAND" = "branch" ] || [ "$COMMAND" = "b" ]; then
        qstack-branch "$@"
    elif [ "$COMMAND" = "list" ] || [ "$COMMAND" = "li" ]; then
        qstack-list "$@"
    elif [ "$COMMAND" = "log" ] || [ "$COMMAND" = "l" ]; then
        qstack-log "$@"
    elif [ "$COMMAND" = "rebase" ] || [ "$COMMAND" = "rb" ]; then
        qstack-rebase "$@"
    elif [ "$COMMAND" = "reorder" ] || [ "$COMMAND" = "ro" ]; then
        qstack-reorder "$@"
    else
       echo "Invalid command"
    fi
}

qstack-help() {
    echo 'usage: qstack ${subcommand} ...
    alias: qs

subcommands:

push
    push all branches in the current stack to remote

pull
    update the base branch from mainstream, then rebase the current stack onto the base branch

log
    alias: l
    git log helper

list
    alias: li
    list all stacks

branch
    alias: b
    list all branches in thestacks

rebase
    alias: rb
    start interactive rebase of the current stack against the base branch

reoder
    alias: ro
    start interactive rebase to reorder branches in the current stack'
}

qstack-branch() {
    BRANCHES=$(git log --pretty='format:%D' $QS_BASE_BRANCH.. --decorate-refs=refs/heads | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "No branches in the current stack"
        return 1
    fi

    echo "----Top------"
    echo $BRANCHES
    echo "---Bottom----"
}

qstack-push() {
    # Reverse so we push from bottom -> top
    BRANCHES=$(git log --pretty='format:%D' $QS_BASE_BRANCH.. --decorate-refs=refs/heads --reverse | grep -v '^$')
    if [ -z "$BRANCHES" ]; then
        echo "No branches in the current stack"
        return 1
    fi
    
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        EXISING_REMOTE_BRANCH=$(git for-each-ref --format='%(upstream:lstrip=3)' "refs/heads/$BRANCH")
        if [ -z "$EXISING_REMOTE_BRANCH" ]; then
            NEW_REMOTE_BRANCH=${BRANCH%"/$QS_TIP_OF_STACK"}
            git push origin --set-upstream "$BRANCH":"$NEW_REMOTE_BRANCH" --force
        else
            git push origin "$BRANCH":"$EXISING_REMOTE_BRANCH" --force
        fi

        echo "" # newline
    done
}

qstack-pull() {
    git checkout $QS_BASE_BRANCH && \
    git pull && \
    git checkout - && \
    git rebase -i $QS_BASE_BRANCH --update-refs
}

qstack-list() {
    BRANCHES=$(git branch --format='%(refname:short)')
    LEAVES=()
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
            LEAVES+=($BRANCH)
        fi
    done
    git branch --list $LEAVES[@]
}

qstack-log() {
    git log $QS_BASE_BRANCH..
}

qstack-rebase() {
    git rebase -i $QS_BASE_BRANCH --update-refs --keep-base
}

qstack-reorder() {
    git checkout -b tmp-reorder-branch && \
    git rebase -i $QS_BASE_BRANCH --update-refs --keep-base && \
    git checkout - && \
    git branch -D tmp-reorder-branch
}
