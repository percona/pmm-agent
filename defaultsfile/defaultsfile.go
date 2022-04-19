// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package defaultsfile provides managing of defaults file.
package defaultsfile

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

// DefaultsFile is a struct with database specs fetched from file.
type DefaultsFile struct {
	username string
	password string
	host     string
	port     uint32
}

// New creates new DefaultsFile.
func New() *DefaultsFile {
	return &DefaultsFile{}
}

// ParseDefaultsFile parses given defaultsFile. It returns the database specs.
func (d *DefaultsFile) ParseDefaultsFile(req *agentpb.ParseDefaultsFileRequest) *agentpb.ParseDefaultsFileResponse {
	var res agentpb.ParseDefaultsFileResponse
	defaultsFile, err := parseDefaultsFile(req.ConfigPath, req.ServiceType)
	if err != nil {
		res.Error = err.Error()
		return &res
	}

	res.Username = defaultsFile.username
	res.Password = defaultsFile.password
	res.Host = defaultsFile.host
	res.Port = defaultsFile.port

	return &res
}

func parseDefaultsFile(configPath string, serviceType inventorypb.ServiceType) (*DefaultsFile, error) {
	if len(configPath) == 0 {
		return nil, errors.New("configPath for DefaultsFile is empty")
	}

	switch serviceType {
	case inventorypb.ServiceType_MYSQL_SERVICE:
		return parseMySQLDefaultsFile(configPath)
	case inventorypb.ServiceType_EXTERNAL_SERVICE:
	case inventorypb.ServiceType_HAPROXY_SERVICE:
	case inventorypb.ServiceType_MONGODB_SERVICE:
	case inventorypb.ServiceType_POSTGRESQL_SERVICE:
	case inventorypb.ServiceType_PROXYSQL_SERVICE:
	case inventorypb.ServiceType_SERVICE_TYPE_INVALID:
		return nil, errors.Errorf("unimplemented service type %s", serviceType)
	}

	return nil, errors.Errorf("unimplemented service type %s", serviceType)
}

func parseMySQLDefaultsFile(configPath string) (*DefaultsFile, error) {
	configPath, err := expandPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("fail to normalize path: %w", err)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("fail to read config file: %w", err)
	}

	port, _ := cfg.Section("client").Key("port").Uint()
	return &DefaultsFile{
		username: cfg.Section("client").Key("user").String(),
		password: cfg.Section("client").Key("password").String(),
		host:     cfg.Section("client").Key("host").String(),
		port:     uint32(port),
	}, nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("failed to expand path: %w", err)
		}
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}
	return path, nil
}
