package storelogs

import (
	"container/ring"
	"fmt"
)

type LogsStore struct {
	MapLogs map[string]*ring.Ring
}

func (l *LogsStore) SetUp(exporter string, countLogs int) {
	l.MapLogs[exporter] = ring.New(countLogs)
}

func (l *LogsStore) LenLogs(exporter string) int {
	r, ok := l.MapLogs[exporter]
	if !ok {
		return 0
	}
	return r.Len()
}

func (l *LogsStore) SaveLog(exporter string, log string) {
	l.MapLogs[exporter].Value = log
	l.MapLogs[exporter] = l.MapLogs[exporter].Next()
}

func (l *LogsStore) GetLogs(exporter string) []string {
	var logs []string
	l.MapLogs[exporter].Do(func(p interface{}) {
		log := fmt.Sprint(p)
		logs = append(logs, log)
	})
	return logs
}

var Logs = new(LogsStore)
