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

// +build child

// Export some identifiers just for process_child.go.

package supervisor

import "os/exec"

func NewProcessParams(path string, args []string) *processParams {
	return &processParams{
		path: path,
		args: args,
	}
}

var NewProcess = newProcess

func GetCmd(pcs *process) *exec.Cmd {
	return pcs.cmd
}
