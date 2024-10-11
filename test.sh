#!/bin/bash
set -e
set -x

git checkout main
git-stacked create a1
git-stacked create a2
git-stacked branch
git checkout main
git-stacked create b1
git-stacked stack
git checkout a2
git-stacked push