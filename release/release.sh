#!/bin/bash
# Exit immediately if any command fails
set -e

CLI_VERSION=$1

if [ -z "$CLI_VERSION" ]; then
  echo "Error: CLI_VERSION argument is required."
  echo "Usage: $0 <CLI_VERSION>"
  exit 1
fi

echo "CLI_VERSION: $CLI_VERSION"
go run release/generate_template.go --template version --version "$CLI_VERSION"
go install ./cmd/git-stack
git stash save # learn command requires a clean repo

SAMPLE_FILE=$(mktemp)
git stack learn --chapter 1 --mode=exec > "$SAMPLE_FILE"
git stash pop
go run release/generate_template.go --template readme --version "$CLI_VERSION" --sample-output "$SAMPLE_FILE"
echo "Release generation complete!"
