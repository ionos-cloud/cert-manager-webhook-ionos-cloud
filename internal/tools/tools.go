//go:build tools

package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/go-licenses/v2"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/norwoodj/helm-docs/cmd/helm-docs"
	_ "github.com/vektra/mockery/v2"
	_ "mvdan.cc/gofumpt"
)
