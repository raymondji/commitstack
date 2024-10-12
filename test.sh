#!/bin/bash
set -e

# Setup
debug() {
    echo "> $@"
    "$@"
    echo ""
}
source ./stack.sh
GS_ENABLE_COLOR_OUTPUT=no

SOURCE_DIR=$(pwd)
TEST_ROOTDIR=(/tmp/git-stacked-test/$RANDOM)
TEST_GITHUB_REPO="git@github.com:raymondji/git-stacked-testing.git"
echo "TEST_ROOTDIR: $TEST_ROOTDIR"

run-test() {
    set -x
    TEST_VARIANT=$1
    TEST_DIR="$TEST_ROOTDIR/$TEST_VARIANT"
    mkdir -p $TEST_DIR
    cd $TEST_DIR

    # Set up the git repo
    if [ "$TEST_VARIANT" = "git" ] || [ "$TEST_VARIANT" = "github" ]; then
        git clone $TEST_GITHUB_REPO .
        REMOTE_URL=$(git config --get remote.origin.url)
        if [ "$REMOTE_URL" != "$TEST_GITHUB_REPO" ]; then
            # Sanity check before we do some of the subsequent destructive operations.
            echo "Error: Remote origin is not set to X. It's set to $REMOTE_URL."
            exit 1
        fi

        git init
        for branch in $(git branch -r | grep -v '\->'); do
            branch_name=$(echo $branch | sed 's/origin\///')
            git push origin --delete $branch_name
        done
        git checkout --orphan main
        git rm -rf .
        echo "This repo is used for testing git-stacked" > README.md
        git add .
        git commit -am "Reset repository to initial state"
        git push --force --set-upstream origin main
    elif [ "TEST_VARIANT" = "gitlab" ]; then
        echo "Gitlab test not implemented yet"
        return 1
    fi
    set +x

    echo "Running..."
    (
        debug git-stacked create a1
        debug git-stacked create a2
        debug git-stacked branch

        debug git checkout main
        debug git-stacked create b1
        debug git-stacked stack

        debug git checkout a2
        debug git-stacked push-force
    ) > "$SOURCE_DIR/test-goldens/$TEST_VARIANT.txt" 2>&1
}

# Run basic git test
echo ""
echo "Running core git test"
echo "====================="
GS_ENABLE_GITHUB_EXTENSION=false
GS_ENABLE_GITLAB_EXTENSION=false
run-test "git"

# TODO: Run gitlab test
# echo "Running github extension test"
# echo "====================="
# GS_ENABLE_GITHUB_EXTENSION=false
# GS_ENABLE_GITLAB_EXTENSION=true
# run-test "gitlab"

# Run github test 
echo ""
echo "Running github extension test"
echo "====================="
GS_ENABLE_GITHUB_EXTENSION=true
GS_ENABLE_GITLAB_EXTENSION=false
run-test "github"
