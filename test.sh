#!/bin/bash
set -e

# Setup
debug() {
    echo "DEBUG: $@"
    "$@"
}
source ./stack.sh

# Run
debug git-stacked create a1

debug git-stacked create a2

debug git-stacked branch

debug git checkout main
debug git-stacked create b1
debug git-stacked stack
debug git checkout a2
debug git-stacked push-force

# Cleanup
git checkout main
git branch -D a1 a2 b1