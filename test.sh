#!/bin/bash
set -e

source ./qstack.sh

# Delete all branches except main
git branch | grep -v "main" | xargs git branch -D 

qstack create logging frontend
qstack branch backend
qstack push
qstack create helm prometheus
qstack list
qstack switch logging
qstack list-branches
qstack rebase