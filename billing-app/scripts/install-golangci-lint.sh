#!/usr/bin/env sh
set -eu

# Installs golangci-lint into GOPATH/bin. Pin to a stable version for reproducibility.
VERSION=${1:-v1.61.0}
echo "Installing golangci-lint ${VERSION}..."
GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint@${VERSION}
command -v golangci-lint >/dev/null 2>&1 || {
  echo "golangci-lint not found in PATH; ensure \n  $(go env GOPATH)/bin is in your PATH" >&2
  exit 1
}
golangci-lint version
