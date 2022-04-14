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

// Package storelogs help to store logs
package storelogs

import (
	"bytes"
	"container/ring"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type LogsStore struct {
	log   *ring.Ring
	Entry *logrus.Entry
	count int
	m     sync.RWMutex
}

func (l *LogsStore) SetUp(entry *logrus.Entry) {
	if l.count == 0 {
		l.count = 10
	}
	l.log = ring.New(l.count)
	l.Entry = entry
}

func (l *LogsStore) SetCountLogs(countLogs int) {
	l.count = countLogs
}

func (l *LogsStore) SaveLog(log string) {
	dt := time.Now()
	var b bytes.Buffer
	b.WriteString(l.Entry.Level.String())
	b.WriteString(" [")
	b.WriteString(dt.Format("01-02-2006 15:04:05"))
	b.WriteString("] ")
	b.WriteString(log)
	for _, v := range l.Entry.Data {
		b.WriteString(fmt.Sprintf("  %v", v))
	}
	l.m.Lock()
	l.log.Value = b.String()
	l.m.Unlock()
	l.log = l.log.Next()
}

func (l *LogsStore) GetLogs() (logs []string) {
	if l != nil {
		l.m.Lock()
		l.log.Do(func(p interface{}) {
			log := fmt.Sprint(p)
			if p != nil {
				logs = append(logs, log)
			}
		})
		l.m.Unlock()
	}
	return logs
}

func (l *LogsStore) Warnf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Warnf(format, v...)
}

func (l *LogsStore) Infof(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Infof(format, v...)
}

func (l *LogsStore) Debugf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Debugf(format, v...)
}

func (l *LogsStore) Tracef(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Tracef(format, v...)
}

func (l *LogsStore) Errorf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Errorf(format, v...)
}

func (l *LogsStore) Trace(message string) {
	l.SaveLog(message)
	l.Entry.Trace(message, nil)
}
