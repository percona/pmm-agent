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

// Package tlshelpers contains helpers for databases tls connections.
package tlshelpers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/percona/pmm-agent/utils/templates"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
)

// RegisterMySQLCerts is used for register TLS config before sql.Open is called.
func RegisterMySQLCerts(files map[string]string, tlsSkipVerify bool) error {
	if files == nil {
		return fmt.Errorf("CreateMySQLTempCerts: nothing to register")
	}

	ca := x509.NewCertPool()
	cert, err := tls.X509KeyPair([]byte(files["tlsCert"]), []byte(files["tlsKey"]))
	if err != nil {
		return errors.Wrap(err, "register MySQL client cert failed")
	}

	if ok := ca.AppendCertsFromPEM([]byte(files["tlsCa"])); ok {
		err = mysql.RegisterTLSConfig("custom", &tls.Config{
			InsecureSkipVerify: tlsSkipVerify,
			RootCAs:            ca,
			Certificates:       []tls.Certificate{cert},
		})
		if err != nil {
			return errors.Wrap(err, "register MySQL CA cert failed")
		}
	}

	return nil
}

func CreateMySQLCerts(processArgs []string, agentProcess *agentpb.SetStateRequest_AgentProcess, path, agentID string) ([]string, error) {
	tempDir := filepath.Join(path, strings.ToLower(agentProcess.Type.String()), agentID)

	tr := &templates.TemplateRenderer{
		TextFiles:          agentProcess.TextFiles,
		TemplateLeftDelim:  agentProcess.TemplateLeftDelim,
		TemplateRightDelim: agentProcess.TemplateRightDelim,
		TempDir:            tempDir,
	}

	files, err := tr.RenderFiles(make(map[string]interface{}))
	if err != nil {
		return []string{}, err
	}

	var ok bool
	var textFiles map[string]string
	if textFiles, ok = files["TextFiles"].(map[string]string); ok {
		args := []string{}
		for _, a := range processArgs {
			switch a {
			case "--mysql.ssl-cert-file=tlsCert":
				args = append(args, fmt.Sprintf("--mysql.ssl-cert-file=%s", textFiles["tlsCert"]))
			case "--mysql.ssl-key-file=tlsKey":
				args = append(args, fmt.Sprintf("--mysql.ssl-key-file=%s", textFiles["tlsKey"]))
			default:
				args = append(args, a)
			}
		}

		return args, nil
	}

	return []string{}, fmt.Errorf("textfiles are not valid")
}
