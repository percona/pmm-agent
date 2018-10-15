//go:generate make api

package main

import (
	"fmt"
	"os"

	"github.com/percona/pmm-agent/app"
	"github.com/percona/pmm-agent/cmd"
)

func main() {
	app := &app.App{}
	if err := app.Config.Read(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cmd.New(app).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
