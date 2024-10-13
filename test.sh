#!/bin/bash
set -e

debug() {
    echo "> $*"
    "$@"
    echo ""
}
# shellcheck disable=SC1091
source ./stack.sh
export GS_ENABLE_COLOR_OUTPUT=no
SOURCE_DIR=$(pwd)
TEST_ROOTDIR="/tmp/git-stacked-test/$RANDOM"
TEST_GITHUB_REPO="git@github.com:raymondji/git-stacked-testing.git"
echo "TEST_ROOTDIR: $TEST_ROOTDIR"

run-test() {
    TEST_EXTENSION=$1
    echo ""
    echo "Running test variant: $TEST_EXTENSION"
    echo "==================================="

    TEST_DIR="$TEST_ROOTDIR/$TEST_EXTENSION"
    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    if [ "$TEST_EXTENSION" = "gitlab" ]; then
        export GS_ENABLE_GITLAB_EXTENSION=true
    elif [ "$TEST_EXTENSION" = "github" ]; then
        export GS_ENABLE_GITHUB_EXTENSION=true
    else
        export GS_ENABLE_GITLAB_EXTENSION=false
        export GS_ENABLE_GITHUB_EXTENSION=false
    fi

    # Set up the git repo
    if [ "$TEST_EXTENSION" = "none" ] || [ "$TEST_EXTENSION" = "github" ]; then
        git clone $TEST_GITHUB_REPO .
        REMOTE_URL=$(git config --get remote.origin.url)
        if [ "$REMOTE_URL" != "$TEST_GITHUB_REPO" ]; then
            # Sanity check before we do some of the subsequent destructive operations.
            echo "Error: Remote origin is not set to X. It's set to $REMOTE_URL."
            exit 1
        fi

        git init
        git checkout --orphan new-main
        git rm -rf .
        echo "This repo is used for testing git-stacked" > README.md
        git add .
        git commit -am "Reset repository to initial state"
        git branch -d main
        git branch -m main
        git push --force --set-upstream origin main
        
        # Delete all remote branches except the new main branch
        # List branches and remove each branch
        for BRANCH in $(git branch -r |  grep -v 'origin/main'); do
            BRANCH_NAME=$(echo "$BRANCH" | sed 's/origin\///')
            git push origin --delete "$BRANCH_NAME"
        done
    elif [ "$TEST_EXTENSION" = "gitlab" ]; then
        echo "Gitlab test not implemented yet"
        return 0
    fi
    

    echo "Running..."
    (
        debug git-stacked stack a1
        debug git-stacked stack a2
        debug git-stacked branch

        debug git checkout main
        debug git-stacked stack b1
        debug git-stacked all

        debug git checkout a2
        debug git-stacked push

        debug git checkout main
        debug git commit --allow-empty -m "New changes in main"
        debug git checkout a2
        debug git-stacked pull
    ) > "$SOURCE_DIR/test-goldens/$TEST_EXTENSION.txt" 2>&1
}

run-test "none"
run-test "github"
run-test "gitlab"