#!/bin/bash
set -e

go build -o git-stack ./cmd/stack/
sudo mv git-stack /usr/local/bin