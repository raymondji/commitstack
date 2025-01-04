#!/bin/bash

SAMPLE_FILE=$(mktemp)
go run ./cmd/git-stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout main
go run ./cmd/git-stack learn --chapter 1 --mode=clean
go run ./readme/generate.go --sample-output "$SAMPLE_FILE"