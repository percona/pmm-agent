package app

import (
	"github.com/percona/pmm-agent/app/client"
	"github.com/percona/pmm-agent/app/config"
	"github.com/percona/pmm-agent/app/format"
	"github.com/percona/pmm-agent/app/server"
)

type App struct {
	Client client.Client
	Server server.Server
	Config config.Config
	Format format.Format
}
