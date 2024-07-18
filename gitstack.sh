#!/bin/bash
GS_BASE_BRANCH=${GS_BASE_BRANCH:-main}
GS_BRANCH_PREFIX=${GS_BRANCH_PREFIX:-$(whoami)}
GS_TIP_OF_STACK="TIP"

gitstack-create() {
    if [ -z "$1" ]; then
        echo "Must specify new stack name"
        return 1
    fi
    if [ -z "$2" ]; then
        echo "Must specify a name for the first branch in the stack"
        return 1
    fi

    NEW_BRANCH="$GS_BRANCH_PREFIX.$1.$2.$GS_TIP_OF_STACK"
    git checkout $GS_BASE_BRANCH && \
    git pull && \
    git checkout -b $NEW_BRANCH && \
    git commit --allow-empty -m "$GS_BRANCH_PREFIX.$1.$2 start"
}

gitstack-branch() {
    if [ -z "$1" ]; then
        echo "Must specify new branch name"
        return 1
    fi

    CURRENT_BRANCH=$(git branch --show-current)
    if [[ ! "$CURRENT_BRANCH" == *"$GS_TIP_OF_STACK" ]]; then
        echo "You must be on the tip of the stack to add a new branch"
        return 1
    fi
    RENAMED_CURRENT_BRANCH=${CURRENT_BRANCH%".$GS_TIP_OF_STACK"}
    STACK=$(echo $CURRENT_BRANCH | cut -d'.' -f2)
    NEW_BRANCH="$GS_BRANCH_PREFIX.$STACK.$1.$GS_TIP_OF_STACK"

    git branch -m $RENAMED_CURRENT_BRANCH && \
    echo "Renamed branch $CURRENT_BRANCH -> $RENAMED_CURRENT_BRANCH" && \
    git checkout -b $NEW_BRANCH && \
    git commit --allow-empty -m "$GS_BRANCH_PREFIX.$STACK.$1 start"
}

gitstack-push() {
    CURRENT_BRANCH=$(git branch --show-current)
    STACK=$(echo $CURRENT_BRANCH | cut -d'.' -f2)
    if [ -z "$STACK" ]; then
        echo "Not within a stack"
        return 1
    fi

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${GS_BRANCH_PREFIX}.$STACK.**.*")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        REMOTE_BRANCH=${BRANCH%".$GS_TIP_OF_STACK"}
        echo "Pushing branch $BRANCH -> $REMOTE_BRANCH"
        git push origin --set-upstream "$BRANCH":"$REMOTE_BRANCH" --force
        echo "" # newline
    done
}

gitstack-list() {
    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${GS_BRANCH_PREFIX}.**.*.$GS_TIP_OF_STACK")
    if [ -z "$BRANCHES" ]; then
        echo "No stacks found"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        STACK=$(echo $BRANCH | cut -d'.' -f2)
        echo $STACK
    done
}

gitstack-list-branches() {
    CURRENT_BRANCH=$(git branch --show-current)
    STACK=$(echo $CURRENT_BRANCH | cut -d'.' -f2)
    if [ -z "$STACK" ]; then
        echo "Not within a stack"
        return 1
    fi

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/$GS_BRANCH_PREFIX.$STACK.**.*")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found"
        return 1
    fi

    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        echo $BRANCH
    done
}

gitstack-checkout() {
    if [ -z "$1" ]; then
        echo "Must specify stack name"
        return 1
    fi 

    BRANCHES=$(git for-each-ref --format='%(refname:short)' "refs/heads/${GS_BRANCH_PREFIX}.${1}.*.$GS_TIP_OF_STACK")
    if [ -z "$BRANCHES" ]; then
        echo "No branches found for stack '${1}'"
        return 1
    fi
    
    echo "$BRANCHES" | while IFS= read -r BRANCH; do
        git checkout "$BRANCH"
        return 0
    done
}

gitstack-rebase() {
    CURRENT_BRANCH=$(git branch --show-current)
    if [[ ! "$CURRENT_BRANCH" == *"$GS_TIP_OF_STACK" ]]; then
        echo "You must be on the tip of the stack to rebase the stack"
        return 1
    fi

    git checkout $GS_BASE_BRANCH && \
    git pull && \
    git checkout - && \
    git rebase -i $GS_BASE_BRANCH --update-refs
}
