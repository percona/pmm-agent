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

// Package config provides access to pmm-agent configuration.
package config

import (
	"os/exec"

	"gopkg.in/alecthomas/kingpin.v2"
)

// Paths represents binaries paths configuration.
type Paths struct {
	NodeExporter   string
	MySQLdExporter string
}

// Lookup replaces paths with absolute paths.
func (p *Paths) Lookup() {
	p.NodeExporter, _ = exec.LookPath(p.NodeExporter)
	p.MySQLdExporter, _ = exec.LookPath(p.MySQLdExporter)
}

// Ports represents ports configuration.
type Ports struct {
	Min uint16
	Max uint16
}

// Config represents pmm-agent's static configuration.
type Config struct {
	ID      string
	Address string

	Debug       bool
	InsecureTLS bool

	Paths Paths
	Ports Ports
}

func Application(cfg *Config, version string) *kingpin.Application {
	app := kingpin.New("pmm-agent", "Version "+version+".")
	app.HelpFlag.Short('h')
	app.Version(version)
	app.Flag("id", "ID of this pmm-agent.").Envar("PMM_AGENT_ID").StringVar(&cfg.ID)
	app.Flag("address", "PMM Server address (host:port).").Envar("PMM_AGENT_ADDRESS").StringVar(&cfg.Address)

	app.Flag("debug", "Enable debug output.").Envar("PMM_AGENT_DEBUG").BoolVar(&cfg.Debug)
	app.Flag("insecure-tls", "Skip TLS certificate validation.").Envar("PMM_AGENT_INSECURE_TLS").BoolVar(&cfg.InsecureTLS)

	app.Flag("node_exporter", "Path to node_exporter to use.").Envar("PMM_NODE_EXPORTER").Default("node_exporter").StringVar(&cfg.Paths.NodeExporter)
	app.Flag("mysqld_exporter", "Path to mysqld_exporter to use.").Envar("PMM_MYSQLD_EXPORTER").Default("mysqld_exporter").StringVar(&cfg.Paths.MySQLdExporter)

	// TODO load configuration from file with kingpin.ExpandArgsFromFile
	// TODO show environment variables in help

	// TODO use [32768,60999] range for ports by default
	//      or try to read /proc/sys/net/ipv4/ip_local_port_range ?

	return app
}
