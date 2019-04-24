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
