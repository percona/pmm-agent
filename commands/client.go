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

package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	httptransport "github.com/go-openapi/runtime/client"
	agentlocal "github.com/percona/pmm/api/agentlocalpb/json/client"
	"github.com/sirupsen/logrus"
)

// setTransport configures transport for accessing local pmm-agent API.
//
// This method is not thread-safe.
func setTransport(port uint16, l *logrus.Entry) {
	// use JSON APIs over HTTP/1.1
	transport := httptransport.New(fmt.Sprintf("127.0.0.1:%d", port), "/", []string{"http"})
	transport.SetLogger(l)
	transport.SetDebug(l.Logger.GetLevel() >= logrus.DebugLevel)
	transport.Context = context.Background()
	// disable HTTP/2
	transport.Transport.(*http.Transport).TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}

	agentlocal.Default.SetTransport(transport)
}

type statusResult struct {
	ConfigFilePath string
}

// status returns locally running pmm-agent status.
// Error is returned if pmm-agent is not running.
//
// This method is not thread-safe.
func status() (*statusResult, error) {
	res, err := agentlocal.Default.AgentLocal.Status(nil)
	if err != nil {
		return nil, err
	}

	return &statusResult{
		ConfigFilePath: res.Payload.ConfigFilePath,
	}, nil
}

func reload() error {
	_, err := agentlocal.Default.AgentLocal.Reload(nil)
	return err
}
