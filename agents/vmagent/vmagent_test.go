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

package vmagent

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

/*

func TestVMAgent_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *config.Config
		want []string
	}{
		{
			name: "ok config",
			cfg: &config.Config{
				Ports: config.Ports{VMAgent: 8429},
				Server: config.Server{
					Address:  "127.0.0.1:8433",
					Username: "admin",
					Password: "admin",
				},
				Paths: config.Paths{
					TempDir: "/tmp/",
				},
			},
			want: []string{
				"--remoteWrite.url=https://admin:admin@127.0.0.1:8433/victoriametrics/api/v1/write",
				"-remoteWrite.tmpDataPath=/tmp/vmagent-tmp-dir",
				"-promscrape.config=/tmp/vmagent-scrape-config.yaml",
				"-remoteWrite.maxDiskUsagePerURL=1073741824",
				"-memory.allowedPercent=10",
				"-httpListenAddr=127.0.0.1:8429",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vma := newVMAgent(&tt.cfg.Paths, &tt.cfg.Server, tt.cfg.Ports.VMAgent)
			args, err := vma.init(nil, nil)
			require.NoError(t, err)
			require.Equal(t, tt.want, args)
		})
	}
}
*/

func TestVMAgent_Init(t *testing.T) {
	type fields struct {
		remoteInsecure      bool
		remoteWriteUserName string
		remoteWritePassword string
		client              *http.Client
		remoteURL           *url.URL
		listenURL           *url.URL
		scrapeConfigPath    string
		lastConfig          []byte
		tmpDir              string
		l                   *logrus.Entry
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "init ok",
			fields: fields{
				scrapeConfigPath:    "/tmp/vmagent-scrape-config.yaml",
				tmpDir:              "/tmp/vmagent-tmp-dir",
				l:                   logrus.WithField("vmanget", "test"),
				listenURL:           &url.URL{Host: "127.0.0.1:8429"},
				remoteWriteUserName: "admin",
				remoteWritePassword: "admin",
				remoteURL: &url.URL{
					Scheme: "https",
					Host:   "127.0.0.1:8443",
					Path:   "/victoriametrics/api/v1/write",
				},
			},
			want: []string{
				"-remoteWrite.url=https://127.0.0.1:8443/victoriametrics/api/v1/write",
				"-remoteWrite.basicAuth.username=admin",
				"-remoteWrite.basicAuth.password=admin",
				"-remoteWrite.tmpDataPath=/tmp/vmagent-tmp-dir",
				"-promscrape.config=/tmp/vmagent-scrape-config.yaml",
				"-remoteWrite.maxDiskUsagePerURL=1073741824",
				"-loggerLevel=WARN",
				"-httpListenAddr=127.0.0.1:8429",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vma := &VMAgent{
				remoteInsecure:   tt.fields.remoteInsecure,
				remoteUserName:   tt.fields.remoteWriteUserName,
				remotePassword:   tt.fields.remoteWritePassword,
				client:           tt.fields.client,
				remoteWriteURL:   tt.fields.remoteURL,
				listenURL:        tt.fields.listenURL,
				scrapeConfigPath: tt.fields.scrapeConfigPath,
				lastConfig:       tt.fields.lastConfig,
				tmpDir:           tt.fields.tmpDir,
				l:                tt.fields.l,
			}
			if got := vma.buildArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("vmagent Init() result not match, \ngot: %v, \nwant %v", got, tt.want)
			}
		})
	}
}
