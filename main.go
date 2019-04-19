// pmm-agent
// Copyright (C) 2018 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"github.com/percona/pmm/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/percona/pmm-agent/commands"
	"github.com/percona/pmm-agent/config"
)

func main() {
	// empty version breaks much of pmm-managed logic
	if version.Version == "" {
		panic("pmm-agent version is not set during build.")
	}

	// check that flags and environment variables are correct, parse command,
	// ignore config file and actual configuration
	app, _ := config.Application(new(config.Config))
	setupCmd := app.Command("setup", "")
	runCmd := app.Command("run", "Run agent. Default command.").Default()
	kingpin.CommandLine = app
	kingpin.HelpFlag = app.HelpFlag
	kingpin.HelpCommand = app.HelpCommand
	kingpin.VersionFlag = app.VersionFlag

	switch cmd := kingpin.Parse(); cmd {
	case setupCmd.FullCommand():
		commands.Setup()
	case runCmd.FullCommand():
		commands.Run()
	default:
		// not reachable due to default kingpin's termination handler; keep it just in case
		kingpin.Fatalf("Unexpected command %q.", cmd)
	}
}
