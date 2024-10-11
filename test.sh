#!/bin/bash
set -e

# Setup
debug() {
    echo "> $@"
    "$@"
    echo ""
}
source ./stack.sh
git update-index --assume-unchanged test-output-golden.txt
GS_COLOR_OUTPUT=no

# Run
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
) > test-output-golden.txt 2>&1

# Cleanup
git checkout main
git branch -D a1 a2 b1
git update-index --no-assume-unchanged test-output-golden.txt
