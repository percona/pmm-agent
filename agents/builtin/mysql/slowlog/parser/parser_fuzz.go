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

// +build gofuzz

// See https://github.com/dvyukov/go-fuzz

package parser

import (
	"bufio"
	"bytes"

	"github.com/percona/go-mysql/log"
)

func Fuzz(data []byte) int {
	r := bufio.NewReader(bytes.NewReader(data))
	p := NewSlowLogParser(r, log.Options{})

	done := make(chan error)
	go func() {
		done <- p.Start()
	}()

	for p.Parse() != nil {
	}

	err := <-done
	if err == nil {
		return 1
	}
	return 0
}
