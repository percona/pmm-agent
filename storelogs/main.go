package storelogs

import (
	"bytes"
	"container/ring"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type LogsStore struct {
	log   *ring.Ring
	Entry *logrus.Entry
	count int
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
	l.log.Value = b.String()
	l.log = l.log.Next()
}

func (l *LogsStore) GetLogs() (logs []string) {
	if l != nil {
		l.log.Do(func(p interface{}) {
			log := fmt.Sprint(p)
			if p != nil {
				logs = append(logs, log)
			}
		})
	}
	return logs
}

func (l *LogsStore) Warnf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Warnf(format, v)
}

func (l *LogsStore) Infof(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Infof(format, v)
}

func (l *LogsStore) Debugf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Debugf(format, v)
}

func (l *LogsStore) Tracef(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Tracef(format, v)
}

func (l *LogsStore) Errorf(format string, v ...interface{}) {
	l.SaveLog(fmt.Sprintf(format, v...))
	l.Entry.Errorf(format, v)
}

func (l *LogsStore) Trace(message string) {
	l.SaveLog(message)
	l.Entry.Trace(message, nil)
}
