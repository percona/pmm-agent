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

package tls_helpers

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/percona/pmm-agent/agents/process"
	"github.com/pkg/errors"
)

// RegisterMySQLCerts is used for register TLS config before sql.Open is called.
func RegisterMySQLCerts(files map[string]string) error {
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
			InsecureSkipVerify: false,
			RootCAs:            ca,
			Certificates:       []tls.Certificate{cert},
		})
		if err != nil {
			return errors.Wrap(err, "register MySQL CA cert failed")
		}
	}

	return nil
}

// CreateMySQLCerts generate certificates into temp folder from provided files.
func CreateMySQLCerts(process *process.Params, files map[string]string, tempDir string) ([]string, error) {
	var certFileNames []string
	for k, v := range files {
		var flag string
		switch k {
		case "tlsCert":
			flag = "mysql.ssl-cert-file"
		case "tlsKey":
			flag = "mysql.ssl-key-file"
		default:
			continue
		}

		tempFile, err := ioutil.TempFile(tempDir, fmt.Sprintf("mysql_ssl_%s_*", k))
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		defer os.Remove(tempFile.Name()) //nolint:errcheck

		if err = ioutil.WriteFile(tempFile.Name(), []byte(v), 0600); err != nil {
			return []string{}, errors.WithStack(err)
		}

		process.Args = append(certFileNames, fmt.Sprintf("--%s=%s", flag, tempFile.Name()))
		certFileNames = append(process.Args, tempFile.Name())
	}

	// TODO SSL: processParams.Args = append(processParams.Args, "--mysql.ssl-skip-verify")

	return certFileNames, nil
}
