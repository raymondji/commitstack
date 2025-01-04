#!/bin/bash
set -e

go install ./cmd/git-stack
SAMPLE_FILE=$(mktemp)
git stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout main
git stack learn --chapter 1 --mode=clean
go run ./readme/generate.go --sample-output "$SAMPLE_FILE"