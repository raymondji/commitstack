#!/bin/bash
set -e

ORIGINAL_BRANCH=$(git rev-parse --abbrev-ref HEAD)
SAMPLE_FILE=$(mktemp)
go install ./cmd/git-stack
git stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout "$ORIGINAL_BRANCH"
git stack learn --chapter 1 --mode=clean
go run ./readme/generate.go --sample-output "$SAMPLE_FILE"
git commit -am "Update README"
