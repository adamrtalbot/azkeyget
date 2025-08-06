//go:build tools

// Package tools declares development dependencies for Go tooling.
// These tools are used by pre-commit hooks and development workflows.
// The build tag 'tools' ensures they are not included in the main build.
package tools

import (
	// Development tools for linting and code quality
	_ "github.com/fzipp/gocyclo/cmd/gocyclo"                // Cyclomatic complexity analyzer
	_ "github.com/go-critic/go-critic/cmd/gocritic"         // Go source code checker
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Comprehensive Go linter
	_ "github.com/mgechev/revive"                           // Fast Go linter
	_ "golang.org/x/tools/cmd/goimports"                    // Tool to update Go import lines
)
