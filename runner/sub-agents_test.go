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

package runner

import (
	"testing"
)

func TestSubAgent_args(t *testing.T) {
	type fields struct {
		params *AgentParams
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			"No args test",
			fields{&AgentParams{
				Args: []string{},
				Port: 1234,
			}},
			[]string{},
			false,
		},
		{
			"Simple test",
			fields{&AgentParams{
				Args: []string{"-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				Port: 1234,
			}},
			[]string{"-web.listen-address=127.0.0.1:1234"},
			false,
		},
		{
			"Multiple args test",
			fields{&AgentParams{
				Args: []string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:{{ .ListenPort }}"},
				Port: 9175,
			}},
			[]string{"-collect.binlog_size", "-web.listen-address=127.0.0.1:9175"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSubAgent(tt.fields.params)
			got, err := m.args()
			if (err != nil) != tt.wantErr {
				t.Errorf("SubAgent.args() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("SubAgent.args() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SubAgent.args() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}
