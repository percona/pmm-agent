// +build tools

package tools

import (
	// Tools.
	_ "github.com/AlekSi/gocoverutil"
	_ "golang.org/x/tools/cmd/goimports"

	// Tools for `go generate`.
	_ "github.com/golang/protobuf/protoc-gen-go"
	// Test requirements.
	_ "github.com/percona/mysqld_exporter"
)
