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

package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	agentlocalpb "github.com/percona/pmm/api/agentlocalpb/json/client"
	managementpb "github.com/percona/pmm/api/managementpb/json/client"
	"github.com/percona/pmm/api/managementpb/json/client/node"
	"github.com/percona/pmm/utils/tlsconfig"
	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/config"
)

// setLocalTransport configures transport for accessing local pmm-agent API.
//
// This method is not thread-safe.
func setLocalTransport(host string, port uint16, l *logrus.Entry) {
	// use JSON APIs over HTTP/1.1
	address := net.JoinHostPort(host, strconv.Itoa(int(port)))
	transport := httptransport.New(address, "/", []string{"http"})
	transport.SetLogger(l)
	transport.SetDebug(l.Logger.GetLevel() >= logrus.DebugLevel)
	transport.Context = context.Background()

	// disable HTTP/2
	httpTransport := transport.Transport.(*http.Transport)
	httpTransport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}

	agentlocalpb.Default.SetTransport(transport)
}

type statusResult struct {
	ConfigFilepath string
}

// localStatus returns locally running pmm-agent status.
// Error is returned if pmm-agent is not running.
//
// This method is not thread-safe.
func localStatus() (*statusResult, error) {
	res, err := agentlocalpb.Default.AgentLocal.Status(nil)
	if err != nil {
		return nil, err
	}

	return &statusResult{
		ConfigFilepath: res.Payload.ConfigFilepath,
	}, nil
}

// localReload reloads locally running pmm-agent.
//
// This method is not thread-safe.
func localReload() error {
	_, err := agentlocalpb.Default.AgentLocal.Reload(nil)
	return err
}

type errFromNginx string

func (e errFromNginx) Error() string {
	return "response from nginx: " + string(e)
}

func (e errFromNginx) GoString() string {
	return fmt.Sprintf("errFromNginx(%q)", string(e))
}

// setServerTransport configures transport for accessing PMM Server API.
//
// This method is not thread-safe.
func setServerTransport(u *url.URL, insecureTLS bool, l *logrus.Entry) {
	// use JSON APIs over HTTP/1.1
	transport := httptransport.New(u.Host, u.Path, []string{u.Scheme})
	if u.User != nil {
		password, _ := u.User.Password()
		transport.DefaultAuthentication = httptransport.BasicAuth(u.User.Username(), password)
	}
	transport.SetLogger(l)
	transport.SetDebug(l.Logger.GetLevel() >= logrus.DebugLevel)
	transport.Context = context.Background()

	// set error handlers for nginx responses if pmm-managed is down
	errorConsumer := runtime.ConsumerFunc(func(reader io.Reader, data interface{}) error {
		b, _ := ioutil.ReadAll(reader)
		return errFromNginx(string(b))
	})
	transport.Consumers = map[string]runtime.Consumer{
		runtime.JSONMime:    runtime.JSONConsumer(),
		runtime.HTMLMime:    errorConsumer,
		runtime.TextMime:    errorConsumer,
		runtime.DefaultMime: errorConsumer,
	}

	// disable HTTP/2, set TLS config
	httpTransport := transport.Transport.(*http.Transport)
	httpTransport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	if u.Scheme == "https" {
		httpTransport.TLSClientConfig = tlsconfig.Get()
		httpTransport.TLSClientConfig.ServerName = u.Hostname()
		httpTransport.TLSClientConfig.InsecureSkipVerify = insecureTLS
	}

	managementpb.Default.SetTransport(transport)
}

// serverRegister registers Node on PMM Server.
//
// This method is not thread-safe.
func serverRegister(cfgSetup *config.Setup) (string, error) {
	nodeTypes := map[string]string{
		"generic":   node.RegisterNodeBodyNodeTypeGENERICNODE,
		"container": node.RegisterNodeBodyNodeTypeCONTAINERNODE,
	}

	res, err := managementpb.Default.Node.RegisterNode(&node.RegisterNodeParams{
		Body: node.RegisterNodeBody{
			NodeType:      pointer.ToString(nodeTypes[cfgSetup.NodeType]),
			NodeName:      cfgSetup.NodeName,
			MachineID:     cfgSetup.MachineID,
			Distro:        cfgSetup.Distro,
			ContainerID:   cfgSetup.ContainerID,
			ContainerName: cfgSetup.ContainerName,
			NodeModel:     cfgSetup.NodeModel,
			Region:        cfgSetup.Region,
			Az:            cfgSetup.Az,
			// TODO CustomLabels:  customLabels,
			Address: cfgSetup.Address,

			Reregister:  cfgSetup.Force,
			MetricsMode: pointer.ToString(strings.ToUpper(cfgSetup.MetricsMode)),
		},
		Context: context.Background(),
	})
	if err != nil {
		return "", err
	}
	return res.Payload.PMMAgent.AgentID, nil
}

// check interfaces
var (
	_ error          = errFromNginx("")
	_ fmt.GoStringer = errFromNginx("")
)
