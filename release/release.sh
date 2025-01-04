#!/bin/bash
set -e

CLI_VERSION=$1

if [ -z "$CLI_VERSION" ]; then
  echo "Error: CLI_VERSION argument is required."
  echo "Usage: $0 <CLI_VERSION>"
  exit 1
fi

if git rev-parse "$CLI_VERSION" >/dev/null 2>&1; then
  echo "Tag '$CLI_VERSION' already exists."
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: Your working directory has uncommitted changes. Please commit or stash them before running this script."
  exit 1
fi

go run release/generate_template.go --template version --version "$CLI_VERSION"
go install ./cmd/git-stack
git stash save # learn command requires a clean repo
SAMPLE_FILE=$(mktemp)
git stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git checkout main
git stack learn --chapter 1 --mode=clean
git stash pop || true
go run release/generate_template.go --template readme --version "$CLI_VERSION" --sample-output "$SAMPLE_FILE"

git add .
git commit -m "Release $CLI_VERSION"
git push
git tag "$CLI_VERSION"
git push origin "$CLI_VERSION"