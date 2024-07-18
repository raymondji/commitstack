#!/bin/bash

BASE_BRANCH=${BASE_BRANCH:-main}
BRANCH_PREFIX=${BRANCH_PREFIX:-$(whoami)}
BRANCH_SUFFIX_TIP_OF_STACK="TIP"

git-stack-create() {
    if [ -z "$1" ]; then
        echo "Must specify new stack name"
        return 1
    fi
    if [ -z "$2" ]; then
        echo "Must specify a name for the first branch in the stack"
        return 1
    fi

    git checkout $BASE_BRANCH
    git pull
    git checkout -b "$BRANCH_PREFIX/$1/$2/$TIP_OF_STACK"
}

git-stack-branch() {
    if [ -z "$1" ]; then
        echo "Must specify new branch name"
        return 1
    fi

    CURRENT_BRANCH=$(git branch --show-current)
    if [[ ! "$CURRENT_BRANCH" == *"$BRANCH_SUFFIX_TIP_OF_STACK" ]]; then
        echo "You must be on the tip of the stack to add a new branch"
        return 1
    fi
    RENAMED_CURRENT_BRANCH=${CURRENT_BRANCH%"/$BRANCH_SUFFIX_TIP_OF_STACK"}
    STACK=$(echo $CURRENT_BRANCH | cut -d'/' -f2)
    NEW_BRANCH="$BRANCH_PREFIX/$STACK/$1/$BRANCH_SUFFIX_TIP_OF_STACK"
    git branch -m $RENAMED_CURRENT_BRANCH
    git checkout -b $NEW_BRANCH
}

git-stack-push() {
    if [ -z "$1" ]; then
        echo "Must specify stack name"
        return 1
    fi

    branches=$(git for-each-ref --format='%(refnameshort)' "refs/heads/${BRANCH_PREFIX}/${1}/**/*")
    if [ -z "$branches" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$branches" | while IFS= read -r branch; do
        echo "Pushing branch: $branch"
        git push origin "$branch" --force-with-lease
    done
}

git-stack-list() {
    if [ -z "$1" ]; then
        echo "Must specify stack name"
        return 1
    fi

    git for-each-ref --format='%(refnameshort)' "refs/heads/${BRANCH_PREFIX}/${1}/**/$BRANCH_SUFFIX_TIP_OF_STACK"
}

git-stack-checkout() {
    if [ -z "$1" ]; then
        echo "Must specify stack name"
        return 1
    fi 

    branches=$(git for-each-ref --format='%(refnameshort)' "refs/heads/${BRANCH_PREFIX}/${1}/**/$BRANCH_SUFFIX_TIP_OF_STACK")
    if [ -z "$branches" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$branches" | while IFS= read -r branch; do
        git checkout "$branch"
        return 0
    done
}

