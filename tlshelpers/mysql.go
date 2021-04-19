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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

	"github.com/percona/pmm-agent/agents/process"
)

// RegisterMySQLCerts is used for register TLS config before sql.Open is called.
func RegisterMySQLCerts(files map[string]string) error {
	if files == nil {
		return fmt.Errorf("CreateMySQLTempCerts: nothing to register")
	}

	var tlsCfg tls.Config
	caBundle := x509.NewCertPool()

	// CA is not mandatory. It is OK if we only have ssl-cert and ssl-key.
	if files["tlsCa"] != "" {
		pemCA, err := ioutil.ReadFile(filepath.Clean(files["tlsCa"]))
		if err != nil {
			return err
		}
		if ok := caBundle.AppendCertsFromPEM(pemCA); ok {
			tlsCfg.RootCAs = caBundle
		} else {
			return errors.Wrapf(err, "failed parse pem-encoded CA certificates from %s", files["tlsCa"])
		}
	}

	if files["tlsCert"] != "" && files["tlsKey"] != "" {
		certPairs := make([]tls.Certificate, 0, 1)
		keypair, err := tls.LoadX509KeyPair(files["tlsCert"], files["tlsKey"])
		if err != nil {
			return errors.Wrapf(err, "failed to parse pem-encoded SSL cert %s or SSL key %s", files["tlsCert"], files["tlsKey"])
		}

		certPairs = append(certPairs, keypair)
		tlsCfg.Certificates = certPairs
		tlsCfg.InsecureSkipVerify = true
	}

	err := mysql.RegisterTLSConfig("custom", &tlsCfg)
	if err != nil {
		return errors.Wrap(err, "register MySQL CA cert failed")
	}

	return nil
}

// ProcessMySQLCertsArgs generate right args for given certificates.
func ProcessMySQLCertsArgs(process *process.Params, files map[string]string, tempDir string) func() {
	certFileNames := []string{}
	for k := range files {
		path := path.Join(tempDir, k)
		certFileNames = append(certFileNames, path)

		switch k {
		case "tlsCert":
			process.Args = append(process.Args, fmt.Sprintf("--%s=%s", "mysql.ssl-cert-file", path))
		case "tlsKey":
			process.Args = append(process.Args, fmt.Sprintf("--%s=%s", "mysql.ssl-key-file", path))
		default:
			continue
		}
	}

	cleanCerts := func() {
		for _, cert := range certFileNames {
			if _, err := os.Stat(cert); os.IsExist(err) {
				e := os.Remove(cert)
				if e != nil {
					log.Error(e)
					return
				}
			}
		}
	}

	return cleanCerts
}
