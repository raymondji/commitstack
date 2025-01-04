#!/bin/bash
set -e

require_clean_repo() {
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: Your working directory has uncommitted changes. Please commit or stash them before running this script."
    exit 1
  fi
}

require_clean_repo
go build ./...
go test ./...
go install ./cmd/git-stack
CLI_VERSION=$(git stack version)
if git rev-parse "$CLI_VERSION" >/dev/null 2>&1; then
  echo "Tag '$CLI_VERSION' already exists."
  exit 1
fi

SAMPLE_FILE=$(mktemp)
git stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout main
git stack learn --chapter 1 --mode=clean

# Ensure generated files are valid
go run release/generate_template.go --template readme --sample-output "$SAMPLE_FILE"
require_clean_repo
git tag "$CLI_VERSION"
git push origin "$CLI_VERSION"