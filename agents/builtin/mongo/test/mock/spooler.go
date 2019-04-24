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

package mock

import (
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
)

type Spooler struct {
	FilesOut      []string          // test provides
	DataOut       map[string][]byte // test provides
	DataIn        []interface{}
	dataChan      chan *qan.Report
	RejectedFiles []string
}

func NewSpooler(dataChan chan *qan.Report) *Spooler {
	s := &Spooler{
		dataChan:      dataChan,
		DataIn:        []interface{}{},
		RejectedFiles: []string{},
	}
	return s
}

func (s *Spooler) Write(data *qan.Report) error {
	if s.dataChan != nil {
		s.dataChan <- data
	} else {
		s.DataIn = append(s.DataIn, data)
	}
	return nil
}
