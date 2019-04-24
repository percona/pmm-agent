package profiler

import (
	"fmt"
	"sync"

	"github.com/percona/pmgo"

	"github.com/percona/pmm-agent/agents/builtin/mongo/profiler/aggregator"
	"github.com/percona/pmm-agent/agents/builtin/mongo/profiler/collector"
	"github.com/percona/pmm-agent/agents/builtin/mongo/profiler/parser"
)

func NewMonitor(
	session pmgo.SessionManager,
	dbName string,
	aggregator *aggregator.Aggregator,
) *monitor {
	return &monitor{
		session:    session,
		dbName:     dbName,
		aggregator: aggregator,
	}
}

type monitor struct {
	// dependencies
	session    pmgo.SessionManager
	dbName     string
	aggregator *aggregator.Aggregator

	// internal services
	services []services

	// state
	sync.RWMutex      // Lock() to protect internal consistency of the service
	running      bool // Is this service running?
}

func (m *monitor) Start() error {
	m.Lock()
	defer m.Unlock()

	if m.running {
		return nil
	}

	defer func() {
		// if we failed to start
		if !m.running {
			// be sure that any started internal service is shutdown
			for _, s := range m.services {
				s.Stop()
			}
			m.services = nil
		}
	}()

	// create collector and start it
	c := collector.New(m.session, m.dbName)
	docsChan, err := c.Start()
	if err != nil {
		return err
	}
	m.services = append(m.services, c)

	// create parser and start it
	p := parser.New(docsChan, m.aggregator)
	err = p.Start()
	if err != nil {
		return err
	}
	m.services = append(m.services, p)

	m.running = true
	return nil
}

func (m *monitor) Stop() {
	m.Lock()
	defer m.Unlock()

	if !m.running {
		return
	}

	// stop internal services
	for _, s := range m.services {
		s.Stop()
	}

	m.running = false
}

// Status returns list of statuses
func (m *monitor) Status() map[string]string {
	m.RLock()
	defer m.RUnlock()

	statuses := &sync.Map{}

	wg := &sync.WaitGroup{}
	wg.Add(len(m.services))
	for _, s := range m.services {
		go func(s services) {
			defer wg.Done()
			for k, v := range s.Status() {
				key := fmt.Sprintf("%s-%s", s.Name(), k)
				statuses.Store(key, v)
			}
		}(s)
	}
	wg.Wait()

	statusesMap := map[string]string{}
	statuses.Range(func(key, value interface{}) bool {
		statusesMap[key.(string)] = value.(string)
		return true
	})

	return statusesMap
}

type services interface {
	Status() map[string]string
	Stop()
	Name() string
}
