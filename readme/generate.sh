#!/bin/bash

if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: Your working directory has uncommitted changes. Please commit or stash them before running this script."
    exit 1
fi

SAMPLE_FILE=$(mktemp)
go run ./cmd/git-stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout main
go run ./cmd/git-stack learn --chapter 1 --mode=clean
go run ./release/generate_template.go --template readme --sample-output "$SAMPLE_FILE"