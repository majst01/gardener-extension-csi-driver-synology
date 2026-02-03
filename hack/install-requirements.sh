#!/bin/bash

set -e

echo "Installing Go dependencies..."
go mod download
go mod tidy

echo "Installing development tools..."
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
go install github.com/golang/mock/mockgen@latest

echo "Requirements installed successfully!"
