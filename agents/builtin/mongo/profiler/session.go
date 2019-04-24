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

package profiler

import (
	"github.com/percona/pmgo"
	"gopkg.in/mgo.v2"
)

func createSession(dialInfo *pmgo.DialInfo, dialer pmgo.Dialer) (pmgo.SessionManager, error) {
	dialInfo.Timeout = MgoTimeoutDialInfo
	// Disable automatic replicaSet detection, connect directly to specified server
	dialInfo.Direct = true
	session, err := dialer.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Eventual, true)
	session.SetSyncTimeout(MgoTimeoutSessionSync)
	session.SetSocketTimeout(MgoTimeoutSessionSocket)

	return session, nil
}
