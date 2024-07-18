#!/bin/bash
QS_BASE_BRANCH=${QS_BASE_BRANCH:-main}
QS_BRANCH_PREFIX=${QS_BRANCH_PREFIX:-$(whoami)}
QS_TIP_OF_STACK="TIP"

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

    if [ "$COMMAND" = "create" ] || [ "$COMMAND" = "c" ]; then
        qstack-create "$@"
    elif [ "$COMMAND" = "branch" ] || [ "$COMMAND" = "b" ]; then
        qstack-branch "$@"
    elif [ "$COMMAND" = "push" ] || [ "$COMMAND" = "p" ]; then
        qstack-push "$@"
    elif [ "$COMMAND" = "list" ] || [ "$COMMAND" = "l" ]; then
        qstack-list "$@"
    elif [ "$COMMAND" = "list-branches" ] || [ "$COMMAND" = "lb" ]; then
        qstack-list-branches "$@"
    elif [ "$COMMAND" = "switch" ] || [ "$COMMAND" = "s" ]; then
        qstack-switch "$@"
    elif [ "$COMMAND" = "rebase" ] || [ "$COMMAND" = "r" ]; then
        qstack-rebase "$@"
    else
       echo "Invalid command"
    fi
}

qstack-create() {
    if [ -z "$1" ]; then
        echo "Must specify new stack name"
        return 1
    fi
    if [[ "$1" == *"/"* ]]; then
        echo "Stack name cannot contain /"
        return 1
    fi
    if [ -z "$2" ]; then
        echo "Must specify a name for the first branch in the stack"
        return 1
    fi
    if [[ "$2" == *"/"* ]]; then
        echo "Branch name cannot contain /"
        return 1
    fi

    NEW_BRANCH="$QS_BRANCH_PREFIX/$1/$2/$QS_TIP_OF_STACK"
    git checkout $QS_BASE_BRANCH && \
    git pull && \
    git checkout -b $NEW_BRANCH && \
    git commit --allow-empty -m "$QS_BRANCH_PREFIX/$1/$2 start"
}

qstack-branch() {
    if [ -z "$1" ]; then
        echo "Must specify new branch name"
        return 1
    fi
    if [[ "$1" == *"/"* ]]; then
        echo "Branch name cannot contain /"
        return 1
    fi

    CURRENT_BRANCH=$(git branch --show-current)
    if [[ ! "$CURRENT_BRANCH" == *"$QS_TIP_OF_STACK" ]]; then
        echo "You must be on the tip of the stack to add a new branch"
        return 1
    fi
    RENAMED_CURRENT_BRANCH=${CURRENT_BRANCH%"/$QS_TIP_OF_STACK"}
    STACK=$(echo $CURRENT_BRANCH | cut -d'/' -f2)
    NEW_BRANCH="$QS_BRANCH_PREFIX/$STACK/$1/$QS_TIP_OF_STACK"

    git branch -m $RENAMED_CURRENT_BRANCH && \
    echo "Renamed branch $CURRENT_BRANCH -> $RENAMED_CURRENT_BRANCH" && \
    git checkout -b $NEW_BRANCH && \
    git commit --allow-empty -m "$QS_BRANCH_PREFIX/$STACK/$1 start"
}

qstack-push() {
    CURRENT_BRANCH=$(git branch --show-current)
    STACK=$(echo $CURRENT_BRANCH | cut -d'/' -f2)
    if [ -z "$STACK" ]; then
        echo "Not within a stack"
        return 1
    fi

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${QS_BRANCH_PREFIX}/$STACK/**/*")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        EXISING_REMOTE_BRANCH=$(git for-each-ref --format='%(upstream:short)' "refs/heads/$BRANCH")
        if [ -z "$EXISING_REMOTE_BRANCH" ]; then
            NEW_REMOTE_BRANCH=${BRANCH%"/$QS_TIP_OF_STACK"}
            echo "Pushing to new remote branch $BRANCH -> $NEW_REMOTE_BRANCH"
            git push origin --set-upstream "$BRANCH":"$NEW_REMOTE_BRANCH" --force
        else
            echo "Pushing to existing remote branch $BRANCH -> $EXISING_REMOTE_BRANCH"
            git push origin "$BRANCH":"$EXISING_REMOTE_BRANCH" --force
        fi

        echo "" # newline
    done
}

qstack-list() {
    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${QS_BRANCH_PREFIX}/**/*/$QS_TIP_OF_STACK")
    if [ -z "$BRANCHES" ]; then
        echo "No stacks found"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        STACK=$(echo $BRANCH | cut -d'/' -f2)
        echo $STACK
    done
}

qstack-list-branches() {
    CURRENT_BRANCH=$(git branch --show-current)
    STACK=$(echo $CURRENT_BRANCH | cut -d'/' -f2)
    if [ -z "$STACK" ]; then
        echo "Not within a stack"
        return 1
    fi

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/$QS_BRANCH_PREFIX/$STACK/**/*")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo $BRANCH
    done
}

qstack-switch() {
    if [ -z "$1" ]; then
        echo "Must specify stack name"
        return 1
    fi 

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${QS_BRANCH_PREFIX}/${1}/*/$QS_TIP_OF_STACK")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        git checkout "$BRANCH"
        return 0
    done
}

qstack-rebase() {
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ ! "$CURRENT_BRANCH" == *"$QS_TIP_OF_STACK" ]]; then
        echo "You must be on the tip of the stack to rebase the stack"
        return 1
    fi

    git checkout $QS_BASE_BRANCH && \
    git pull && \
    git checkout - && \
    git rebase -i $QS_BASE_BRANCH --update-refs
}
