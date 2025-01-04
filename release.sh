#!/bin/bash
set -e

go build ./...
go test ./...
go install ./cmd/git-stack

CLI_VERSION=$(git stack version)
if git rev-parse "$CLI_VERSION" >/dev/null 2>&1; then
  echo "Tag '$CLI_VERSION' already exists."
  exit 1
fi

readme/generate.sh
git add .
git commit -am "Release $CLI_VERSION"
git push

git tag "$CLI_VERSION"
git push origin "$CLI_VERSION"
go install "github.com/raymondji/git-stack-cli/cmd/git-stack@$CLI_VERSION"
