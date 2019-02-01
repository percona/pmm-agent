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

//+build ignore

// Run it with:
//   go run -tags child process_child.go

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/percona/pmm-agent/supervisor"
)

func main() {
	flag.Parse()
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	l := logrus.NewEntry(logger)

	process := supervisor.NewProcess(context.Background(), supervisor.NewProcessParams("sleep", []string{"100500"}), l)

	<-process.Changes()
	<-process.Changes()

	cmd := supervisor.ExportCmd(process)

	fmt.Println(cmd.Process.Pid)
	time.Sleep(30 * time.Second)
}
