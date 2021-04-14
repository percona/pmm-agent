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
	"path"

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
	for k, _ := range files {
		var flag string
		switch k {
		case "tlsCert":
			flag = "mysql.ssl-cert-file"
		case "tlsKey":
			flag = "mysql.ssl-key-file"
		default:
			continue
		}

		// tempFile, err := ioutil.TempFile(tempDir, fmt.Sprintf("mysql_ssl_%s_*", k))
		// if err != nil {
		// 	return []string{}, errors.WithStack(err)
		// }
		// // TODO REMOVE
		// //defer os.Remove(tempFile.Name()) //nolint:errcheck

		// if err = ioutil.WriteFile(tempFile.Name(), []byte(v), 0600); err != nil {
		// 	return []string{}, errors.WithStack(err)
		// }
		path := path.Join(tempDir, k)
		process.Args = append(process.Args, fmt.Sprintf("--%s=%s", flag, path))
		certFileNames = append(certFileNames, path)
	}

	// TODO REMOVE
	process.Args = append(process.Args, "--mysql.ssl-skip-verify")

	return certFileNames, nil
}
