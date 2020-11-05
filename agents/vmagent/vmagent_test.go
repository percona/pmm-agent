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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/percona/pmm-agent/config"
)

func TestVMAgent_args(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want []string
	}{
		{
			name: "init ok with base cfg",
			cfg: &config.Config{
				Paths: config.Paths{TempDir: "/tmp"},
				Ports: config.Ports{VMAgent: 8429},
				Server: config.Server{
					Address:  "127.0.0.1:443",
					Password: "admin",
					Username: "admin",
				},
			},
			want: []string{
				"-remoteWrite.url=https://127.0.0.1:443/victoriametrics/api/v1/write",
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
			vma, err := NewVMAgent(tt.cfg)
			assert.NoError(t, err)
			got := vma.args()
			assert.Equal(t, tt.want, got)
		})
	}
}
