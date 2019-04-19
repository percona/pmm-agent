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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/percona/pmm/version"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

// Server represents PMM Server configration.
type Server struct {
	Address     string `yaml:"address"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	InsecureTLS bool   `yaml:"insecure-tls"`
}

// Paths represents binaries paths configuration.
type Paths struct {
	NodeExporter     string `yaml:"node_exporter"`
	MySQLdExporter   string `yaml:"mysqld_exporter"`
	MongoDBExporter  string `yaml:"mongodb_exporter"`
	PostgresExporter string `yaml:"postgres_exporter"`
	TempDir          string `yaml:"tempdir"`
}

// Lookup replaces paths with absolute paths.
func (p *Paths) Lookup() {
	p.NodeExporter, _ = exec.LookPath(p.NodeExporter)
	p.MySQLdExporter, _ = exec.LookPath(p.MySQLdExporter)
	p.MongoDBExporter, _ = exec.LookPath(p.MongoDBExporter)
	p.PostgresExporter, _ = exec.LookPath(p.PostgresExporter)
}

// Ports represents ports configuration.
type Ports struct {
	Min uint16 `yaml:"min"`
	Max uint16 `yaml:"max"`
}

// Config represents pmm-agent's static configuration.
//nolint:maligned
type Config struct {
	ID         string `yaml:"id"`
	ListenPort uint16 `yaml:"listen-port"`

	Server Server `yaml:"server"`
	Paths  Paths  `yaml:"paths"`
	Ports  Ports  `yaml:"ports"`

	Debug bool `yaml:"debug"`
	Trace bool `yaml:"trace"`
}

func Get(l *logrus.Entry) (*Config, string, error) {
	return get(os.Args[1:], l)
}

func get(args []string, l *logrus.Entry) (*Config, string, error) {
	// parse flags and environment variables
	cfg := new(Config)
	app, configFileF := Application(cfg)
	_, err := app.Parse(args)
	if err != nil {
		return nil, "", err
	}

	// if config file is given (it must exist), read and parse it, then re-parse flags into this configuration
	if *configFileF != "" {
		l.Debugf("Loading configuration file %s.", *configFileF)
		if cfg, err = LoadFromFile(*configFileF); err != nil {
			return nil, "", err
		}
		if cfg == nil {
			return nil, "", fmt.Errorf("configuration file %q does not exist", *configFileF)
		}
		app, _ = Application(cfg)
		if _, err = app.Parse(args); err != nil {
			return nil, "", err
		}
	}

	cfg.Paths.Lookup()
	return cfg, *configFileF, nil
}

// Application returns kingpin application that parses all flags and environment variables into cfg
// except --config-file that is returned separately.
func Application(cfg *Config) (*kingpin.Application, *string) {
	app := kingpin.New("pmm-agent", fmt.Sprintf("Version %s.", version.Version))
	app.HelpFlag.Short('h')
	app.Version(version.FullInfo())

	// this flags has to be optional and has empty default value for `pmm-agent setup`
	configFileF := app.Flag("config-file", "Configuration file path. [PMM_AGENT_CONFIG_FILE]").
		Envar("PMM_AGENT_CONFIG_FILE").PlaceHolder("</path/to/pmm-agent.yaml>").String()

	app.Flag("id", "ID of this pmm-agent. [PMM_AGENT_ID]").
		Envar("PMM_AGENT_ID").PlaceHolder("</agent_id/...>").StringVar(&cfg.ID)
	app.Flag("listen-port", "Agent local API port. [PMM_AGENT_LISTEN_PORT]").
		Envar("PMM_AGENT_LISTEN_PORT").Default("7777").Uint16Var(&cfg.ListenPort)

	app.Flag("server-address", "PMM Server address. [PMM_AGENT_SERVER_ADDRESS]").
		Envar("PMM_AGENT_SERVER_ADDRESS").PlaceHolder("<host:port>").StringVar(&cfg.Server.Address)
	app.Flag("server-username", "HTTP BasicAuth username to connect to PMM Server. [PMM_AGENT_SERVER_USERNAME]").
		Envar("PMM_AGENT_SERVER_USERNAME").StringVar(&cfg.Server.Username)
	app.Flag("server-password", "HTTP BasicAuth password to connect to PMM Server. [PMM_AGENT_SERVER_PASSWORD]").
		Envar("PMM_AGENT_SERVER_PASSWORD").StringVar(&cfg.Server.Password)
	app.Flag("server-insecure-tls", "Skip PMM Server TLS certificate validation. [PMM_AGENT_SERVER_INSECURE_TLS]").
		Envar("PMM_AGENT_SERVER_INSECURE_TLS").BoolVar(&cfg.Server.InsecureTLS)

	app.Flag("paths-node_exporter", "Path to node_exporter to use. [PMM_AGENT_PATHS_NODE_EXPORTER]").
		Envar("PMM_AGENT_PATHS_NODE_EXPORTER").Default("node_exporter").StringVar(&cfg.Paths.NodeExporter)
	app.Flag("paths-mysqld_exporter", "Path to mysqld_exporter to use. [PMM_AGENT_PATHS_MYSQLD_EXPORTER]").
		Envar("PMM_AGENT_PATHS_MYSQLD_EXPORTER").Default("mysqld_exporter").StringVar(&cfg.Paths.MySQLdExporter)
	app.Flag("paths-mongodb_exporter", "Path to mongodb_exporter to use. [PMM_AGENT_PATHS_MONGODB_EXPORTER]").
		Envar("PMM_AGENT_PATHS_MONGODB_EXPORTER").Default("mongodb_exporter").StringVar(&cfg.Paths.MongoDBExporter)
	app.Flag("paths-postgres_exporter", "Path to postgres_exporter to use. [PMM_AGENT_PATHS_POSTGRES_EXPORTER]").
		Envar("PMM_AGENT_PATHS_POSTGRES_EXPORTER").Default("postgres_exporter").StringVar(&cfg.Paths.PostgresExporter)
	app.Flag("paths-tempdir", "Temporary directory for exporters. [PMM_AGENT_PATHS_TEMPDIR]").
		Envar("PMM_AGENT_PATHS_TEMPDIR").Default(os.TempDir()).StringVar(&cfg.Paths.TempDir)

	// TODO read defaults from /proc/sys/net/ipv4/ip_local_port_range ?
	app.Flag("ports-min", "Minimal allowed port number for listening sockets. [PMM_AGENT_PORTS_MIN]").
		Envar("PMM_AGENT_PORTS_MIN").Default("32768").Uint16Var(&cfg.Ports.Min)
	app.Flag("ports-max", "Maximal allowed port number for listening sockets. [PMM_AGENT_PORTS_MAX]").
		Envar("PMM_AGENT_PORTS_MAX").Default("60999").Uint16Var(&cfg.Ports.Max)

	app.Flag("debug", "Enable debug output. [PMM_AGENT_DEBUG]").
		Envar("PMM_AGENT_DEBUG").BoolVar(&cfg.Debug)
	app.Flag("trace", "Enable trace output (implies debug). [PMM_AGENT_TRACE]").
		Envar("PMM_AGENT_TRACE").BoolVar(&cfg.Trace)

	return app, configFileF
}

// LoadFromFile loads configuration from file.
// As a special case, if file does not exist, it returns (nil, nil).
// Error is returned if file exists, but configuration can't be loaded due to permission problems,
// YAML parsing problems, etc.
func LoadFromFile(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	b, err := ioutil.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err = yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveToFile saves configuration to file.
// No special cases.
func SaveToFile(path string, cfg *Config) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0640)
}
