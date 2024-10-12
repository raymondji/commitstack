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
    VARIANT=$1
    TEST_DIR="$TEST_ROOTDIR/$VARIANT"
    mkdir -p $TEST_DIR
    cd $TEST_DIR

    # Set up the git repo
    if [ "$VARIANT" = "git" ] || [ "$VARIANT" = "github" ]; then
        git clone $TEST_GITHUB_REPO .
        rm -rf .git
        git init
        git remote add origin $TEST_GITHUB_REPO
        echo "This repo is used for testing git-stacked" > README.md
        git add .
        git commit -am "First commit"
        git push --force --set-upstream origin main
    elif [ "VARIANT" = "gitlab" ]; then
        echo "Gitlab test not implemented yet"
        return 1
    fi

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
    ) > "$SOURCE_DIR/test-goldens/$VARIANT.txt" 2>&1
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
