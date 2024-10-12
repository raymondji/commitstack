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

TEST_ROOTDIR=/tmp/git-stacked-test
mkdir -p $TEST_ROOTDIR
rm -rf "$TEST_ROOTDIR/*"

run-test() {
    VARIANT=$1
    TEST_DIR="$TEST_ROOTDIR/$VARIANT"
    cd $TEST_DIR
    git init

    if [ "$VARIANT" = "github" ]; then
        git remote add origin git@github.com:raymondji/git-stacked-testing.git
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
    ) > "test-goldens/$VARIANT.txt" 2>&1
}

# Run basic git test
GS_ENABLE_GITHUB_EXTENSION=false
GS_ENABLE_GITLAB_EXTENSION=false
run-test "git"

# Run gitlab test
TEST_DIR="$TEST_ROOTDIR/gitlab"
GS_ENABLE_GITHUB_EXTENSION=false
GS_ENABLE_GITLAB_EXTENSION=true
run-test "gitlab"

# Run github test
TEST_DIR="$TEST_ROOTDIR/github"
GS_ENABLE_GITHUB_EXTENSION=true
GS_ENABLE_GITLAB_EXTENSION=false
run-test "github"
