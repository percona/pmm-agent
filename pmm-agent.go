//go:generate make api

package main

import "github.com/percona/pmm-agent/cmd"

func main() {
	cmd.Execute()
}
