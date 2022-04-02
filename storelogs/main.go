package storelogs

import (
	"container/ring"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type LogsStore struct {
	log   *ring.Ring
	entry *logrus.Entry
}

func (l *LogsStore) SetUp(countLogs int, entry *logrus.Entry) {
	l.log = ring.New(countLogs)
	l.entry = entry
}

func (l *LogsStore) SaveLog(log string) {
	dt := time.Now()
	dt.Format("01-02-2006 15:04:05")
	log = dt.Format("01-02-2006 15:04:05") + " " + l.entry.Level.String() + ": " + log
	for _, v := range l.entry.Data {
		log = log + fmt.Sprintf("       %v", v)
	}
	l.log.Value = log
	l.log = l.log.Next()
	//l.MapLogs[exporter].Value = log
	//l.MapLogs[exporter] = l.MapLogs[exporter].Next()
}

func (l *LogsStore) GetLogs() []string {
	var logs []string
	l.log.Do(func(p interface{}) {
		log := fmt.Sprint(p)
		if p != nil {
			logs = append(logs, log)
		}
	})
	return logs
}
