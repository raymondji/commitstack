#!/bin/bash
set -e

source ./stack.sh

other_branches=$(git branch --list | grep -v 'main')
if [ -n "$other_branches" ]; then
  echo "Error: please delete all other branches before running the test script"
  exit 1
fi

git checkout -b foo/1
git-stacked stack
git-stacked branch
# TODO: more steps

# Cleanup
git checkout main
git branch | grep -v "main" | xargs git branch -D 