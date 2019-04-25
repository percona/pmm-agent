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

package profiling

import (
	"fmt"

	"github.com/percona/pmgo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type run func(session pmgo.SessionManager) error
type runDB func(session pmgo.SessionManager, dbname string) error

type Profiling struct {
	url     string
	session pmgo.SessionManager
	err     error
}

func New(url string) *Profiling {
	p := &Profiling{
		url: url,
	}
	p.session, p.err = createSession(url)
	return p
}

func (p *Profiling) Enable(dbname string) error {
	return p.Run(func(session pmgo.SessionManager) error {
		return profile(session.DB(dbname), 2)
	})
}

func (p *Profiling) Disable(dbname string) error {
	return p.Run(func(session pmgo.SessionManager) error {
		return profile(session.DB(dbname), 0)
	})
}

func (p *Profiling) Drop(dbname string) error {
	return p.Run(func(session pmgo.SessionManager) error {
		if !p.Exist(dbname) {
			return nil
		}
		return session.DB(dbname).C("system.profile").DropCollection()
	})
}

func (p *Profiling) Exist(dbnameToLook string) bool {
	found := fmt.Errorf("found db: %s", dbnameToLook)
	err := p.RunDB(func(session pmgo.SessionManager, dbname string) error {
		if dbnameToLook == dbname {
			return found
		}
		return nil
	})

	return err == found
}

func (p *Profiling) Reset(dbname string) error {
	err := p.Disable(dbname)
	if err != nil {
		return err
	}
	err = p.Drop(dbname)
	if err != nil {
		return err
	}
	err = p.Enable(dbname)
	if err != nil {
		return err
	}
	return nil
}

func (p *Profiling) EnableAll() error {
	return p.RunDB(func(session pmgo.SessionManager, dbname string) error {
		return p.Enable(dbname)
	})
}

func (p *Profiling) DisableAll() error {
	return p.RunDB(func(session pmgo.SessionManager, dbname string) error {
		return p.Disable(dbname)
	})
}

func (p *Profiling) DropAll() error {
	return p.RunDB(func(session pmgo.SessionManager, dbname string) error {
		return p.Drop(dbname)
	})
}

func (p *Profiling) ResetAll() error {
	err := p.DisableAll()
	if err != nil {
		return err
	}
	p.DropAll()
	err = p.EnableAll()
	if err != nil {
		return err
	}
	return nil
}

func (p *Profiling) Run(f run) error {
	if p.err != nil {
		return p.err
	}
	session := p.session.Copy()
	defer session.Close()

	return f(session)
}

func (p *Profiling) RunDB(f runDB) error {
	return p.Run(func(session pmgo.SessionManager) error {
		databases, err := session.DatabaseNames()
		if err != nil {
			return err
		}
		for _, dbname := range databases {
			err := f(session, dbname)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DatabaseNames returns the names of non-empty databases present in the cluster.
func (p *Profiling) DatabaseNames() ([]string, error) {
	return p.session.DatabaseNames()
}

func profile(db pmgo.DatabaseManager, v int) error {
	result := struct {
		Was       int
		Slowms    int
		Ratelimit int
	}{}
	return db.Run(
		bson.M{
			"profile": v,
		},
		&result,
	)
}

func createSession(url string) (pmgo.SessionManager, error) {
	dialInfo, err := pmgo.ParseURL(url)
	if err != nil {
		return nil, err
	}
	dialer := pmgo.NewDialer()

	// Disable automatic replicaSet detection, connect directly to specified server
	dialInfo.Direct = true
	session, err := dialer.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Eventual, true)
	return session, nil
}
