#!/bin/bash
set -e
set -x

# Setup
function debug_on() {
    set +x
}
function debug_off() {
    set -x
}
trap 'debug_on' DEBUG
trap 'debug_off' RETURN
source ./stack.sh

# Run
git-stacked create a1
git-stacked create a2
git-stacked branch
git checkout main
git-stacked create b1
git-stacked stack
git checkout a2
git-stacked push-force

set +x
# Cleanup
git branch -D a1 a2 b1