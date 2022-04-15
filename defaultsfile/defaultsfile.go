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

package defaultsfile

import (
	"fmt"
	"github.com/percona/pmm/api/agentpb"
	"github.com/percona/pmm/api/inventorypb"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

type DefaultsFile struct {
	username string
	password string
}

func New() *DefaultsFile {
	return &DefaultsFile{}
}

func (d *DefaultsFile) ParseDefaultsFile(req *agentpb.ParseDefaultsFileRequest) *agentpb.ParseDefaultsFileResponse {
	var res agentpb.ParseDefaultsFileResponse
	defaultsFile, err := parseDefaultsFile(req.ConfigPath, req.ServiceType)
	if err != nil {
		res.Error = err.Error()
		return &res
	}

	res.Username = defaultsFile.username
	res.Password = defaultsFile.password

	return &res
}

func parseDefaultsFile(configPath string, serviceType inventorypb.ServiceType) (*DefaultsFile, error) {
	if len(configPath) == 0 {
		return nil, errors.New("configPath for DefaultsFile is empty")
	}

	switch serviceType {
	case inventorypb.ServiceType_MYSQL_SERVICE:
		return parseMySqlDefaultsFile(configPath)
	default:
		return nil, errors.Errorf("unimplemented service type %s", serviceType)
	}

}

func parseMySqlDefaultsFile(configPath string) (*DefaultsFile, error) {
	configPath, err := expandPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("fail to normalize path: %v", err)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("fail to read config file: %v", err)
	}

	return &DefaultsFile{
		username: cfg.Section("client").Key("user").String(),
		password: cfg.Section("client").Key("password").String(),
	}, nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}
	return path, nil
}
