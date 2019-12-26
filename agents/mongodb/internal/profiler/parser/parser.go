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

package parser

import (
	"context"
	"runtime/pprof"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"

	"github.com/percona/pmm-agent/agents/mongodb/internal/profiler/aggregator"
)

func New(docsChan <-chan proto.SystemProfile, aggregator *aggregator.Aggregator, logger *logrus.Entry) *Parser {
	return &Parser{
		docsChan:   docsChan,
		aggregator: aggregator,
		logger:     logger,
	}
}

type Parser struct {
	// dependencies
	docsChan   <-chan proto.SystemProfile
	aggregator *aggregator.Aggregator

	logger *logrus.Entry

	// state
	sync.RWMutex                 // Lock() to protect internal consistency of the service
	running      bool            // Is this service running?
	doneChan     chan struct{}   // close(doneChan) to notify goroutines that they should shutdown
	wg           *sync.WaitGroup // Wait() for goroutines to stop after being notified they should shutdown
}

// Start starts but doesn't wait until it exits
func (self *Parser) Start() error {
	self.Lock()
	defer self.Unlock()
	if self.running {
		return nil
	}

	// create new channels over which we will communicate to...
	// ... inside goroutine to close it
	self.doneChan = make(chan struct{})

	// start a goroutine and Add() it to WaitGroup
	// so we could later Wait() for it to finish
	self.wg = &sync.WaitGroup{}
	self.wg.Add(1)

	ctx := context.Background()
	labels := pprof.Labels("component", "mongodb.monitor")
	pprof.Do(ctx, labels, func(ctx context.Context) {
		go start(
			self.wg,
			self.docsChan,
			self.aggregator,
			self.doneChan,
			self.logger,
		)
	})

	self.running = true
	return nil
}

// Stop stops running
func (self *Parser) Stop() {
	self.Lock()
	defer self.Unlock()
	if !self.running {
		return
	}
	self.running = false

	// notify goroutine to close
	close(self.doneChan)

	// wait for goroutines to exit
	self.wg.Wait()
	return
}

func (self *Parser) Name() string {
	return "parser"
}

func start(wg *sync.WaitGroup, docsChan <-chan proto.SystemProfile, aggregator *aggregator.Aggregator, doneChan <-chan struct{}, logger *logrus.Entry) {
	// signal WaitGroup when goroutine finished
	defer wg.Done()

	// update stats
	for {
		// check if we should shutdown
		select {
		case <-doneChan:
			return
		default:
			// just continue if not
		}

		// aggregate documents and create report
		select {
		case doc, ok := <-docsChan:
			// if channel got closed we should exit as there is nothing we can listen to
			if !ok {
				return
			}

			// aggregate the doc
			err := aggregator.Add(doc)
			if err != nil {
				logger.Warn("couldn't add document to aggregator")
			}
		case <-doneChan:
			// doneChan needs to be repeated in this select as docsChan can block
			// doneChan needs to be also in separate select statement
			// as docsChan could be always picked since select picks channels pseudo randomly
			return
		}
	}
}
