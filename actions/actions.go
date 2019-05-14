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

package actions

import (
	"context"
)

// Action describe abstract thing that can be runned by a client and returns some output.
// Every structure that implement this interface can be an action.
type Action interface {
	// ID returns an action UUID. Used in log messages.
	ID() string
	// String representation of action name. Used in log messages.
	Name() string
	// Run runs an action and returns output and error.
	// This method can be blocking.
	Run(ctx context.Context) ([]byte, error)
	// Stop stops an action.
	// It should return "false" if action can't be stopped for some reason.
	Stop() bool
}
