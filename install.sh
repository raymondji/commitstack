#!/bin/bash

go build -o git-stack ./cmd/stack/main.go
sudo mv git-stack /usr/local/bin