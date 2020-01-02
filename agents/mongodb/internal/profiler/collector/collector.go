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

package collector

import (
	"context"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MgoTimeoutTail = 1 * time.Second
)

var (
	cursorTimeout = 3 * time.Second
)

// New creates new Collector.
func New(client *mongo.Client, dbName string, logger *logrus.Entry) *Collector {
	return &Collector{
		client: client,
		dbName: dbName,
		logger: logger,
	}
}

type Collector struct {
	// dependencies
	client *mongo.Client
	dbName string
	logger *logrus.Entry

	// provides
	docsChan chan proto.SystemProfile

	// state
	m        sync.Mutex      // Lock() to protect internal consistency of the service
	running  bool            // Is this service running?
	doneChan chan struct{}   // close(doneChan) to notify goroutines that they should shutdown
	wg       *sync.WaitGroup // Wait() for goroutines to stop after being notified they should shutdown
}

// Start starts but doesn't wait until it exits
func (c *Collector) Start() (<-chan proto.SystemProfile, error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.running {
		return nil, nil
	}

	// create new channels over which we will communicate to...
	// ... outside world by sending collected docs
	c.docsChan = make(chan proto.SystemProfile, 100)
	// ... inside goroutine to close it
	c.doneChan = make(chan struct{})

	// start a goroutine and Add() it to WaitGroup
	// so we could later Wait() for it to finish
	c.wg = &sync.WaitGroup{}
	c.wg.Add(1)

	// create ready sync.Cond so we could know when goroutine actually started getting data from db
	ready := sync.NewCond(&sync.Mutex{})
	ready.L.Lock()
	defer ready.L.Unlock()

	ctx := context.Background()
	labels := pprof.Labels("component", "mongodb.aggregator")
	pprof.Do(ctx, labels, func(ctx context.Context) {
		go start(
			ctx,
			c.wg,
			c.client,
			c.dbName,
			c.docsChan,
			c.doneChan,
			ready,
			c.logger,
		)
	})

	// wait until we actually fetch data from db
	ready.Wait()

	c.running = true

	return c.docsChan, nil
}

// Stop stops running
func (c *Collector) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	if !c.running {
		return
	}

	c.running = false
	close(c.doneChan) // notify goroutine to close
	c.wg.Wait()       // wait for goroutines to exit
	close(c.docsChan) // we can now safely close channels goroutines write to as goroutine is stopped
}

func (c *Collector) Name() string {
	return "collector"
}

func start(ctx context.Context, wg *sync.WaitGroup, client *mongo.Client, dbName string,
	docsChan chan<- proto.SystemProfile, doneChan <-chan struct{}, ready *sync.Cond, logger *logrus.Entry) {
	// signal WaitGroup when goroutine finished
	defer wg.Done()

	lastCollectTime := time.Now()
	firstTry := true
	for {
		now := time.Now()
		// make a connection and collect data
		connectAndCollect(
			ctx,
			client,
			dbName,
			docsChan,
			doneChan,
			ready,
			logger,
			lastCollectTime,
			now,
		)
		lastCollectTime = now

		select {
		// check if we should shutdown
		case <-ctx.Done():
			return
		case <-doneChan:
			return
		// wait some time before reconnecting
		case <-time.After(1 * time.Second):
		}

		// After first failure in connection we signal that we are ready anyway
		// this way service starts, and will automatically connect when db is available.
		if firstTry {
			signalReady(ready)
			firstTry = false
		}
	}
}

func connectAndCollect(ctx context.Context, client *mongo.Client, dbName string, docsChan chan<- proto.SystemProfile, doneChan <-chan struct{}, ready *sync.Cond, logger *logrus.Entry, startTime, endTime time.Time) { //nolint: lll
	logger.Traceln("connect and collect is called")
	collection := client.Database(dbName).Collection("system.profile")
	query := createQuery(dbName, startTime, endTime)

	timeoutCtx, cancel := context.WithTimeout(context.TODO(), cursorTimeout)
	defer cancel()
	cursor, err := createIterator(timeoutCtx, collection, query)
	if err != nil {
		logger.Errorf("couldn't create system.profile iterator, reason: %v", err)
		return
	}
	// do not cancel cursor closing when ctx is canceled
	defer cursor.Close(context.Background()) //nolint:errcheck

	// we got iterator, we are ready
	signalReady(ready)

	// check if we should shutdown
	select {
	case <-ctx.Done():
		return
	case <-doneChan:
		return
	default:
		// just continue if not
	}

	for cursor.TryNext(timeoutCtx) {
		doc := proto.SystemProfile{}
		e := cursor.Decode(&doc)
		if e != nil {
			logger.Error(e)
			continue
		}

		// check if we should shutdown
		select {
		case <-ctx.Done():
			return
		case <-doneChan:
			return
		default:
			// just continue if not
		}

		// try to push doc
		select {
		case docsChan <- doc:
		// or exit if we can't push the doc and we should shutdown
		// note that if we can push the doc then exiting is not guaranteed
		// that's why we have separate `select <-doneChan` above
		case <-doneChan:
			return
		}
	}
	if err := cursor.Err(); err != nil {
		logger.Warnln("couldn't retrieve data from cursor", err)
	}
}

func createQuery(dbName string, startTime, endTime time.Time) bson.M {
	return bson.M{
		"ns": bson.M{"$ne": dbName + ".system.profile"},
		"ts": bson.M{"$gt": startTime, "$lt": endTime},
	}
}

func createIterator(ctx context.Context, collection *mongo.Collection, query bson.M) (*mongo.Cursor, error) {
	opts := options.Find().SetSort(bson.M{"$natural": 1}).SetCursorType(options.NonTailable)
	return collection.Find(ctx, query, opts)
}

func signalReady(ready *sync.Cond) {
	ready.L.Lock()
	defer ready.L.Unlock()
	ready.Broadcast()
}
