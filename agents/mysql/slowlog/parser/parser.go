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

// Package parser implements a MySQL slow log parser.
package parser

import (
	stdlog "log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/percona/go-mysql/log"
)

// Regular expressions to match important lines in slow log.
var (
	timeRe    = regexp.MustCompile(`Time: (\S+\s{1,2}\S+)`)
	timeNewRe = regexp.MustCompile(`Time:\s+(\d{4}-\d{2}-\d{2}\S+)`)
	userRe    = regexp.MustCompile(`User@Host: ([^\[]+|\[[^[]+\]).*?@ (\S*) \[(.*)\]`)
	schema    = regexp.MustCompile(`Schema: +(.*?) +Last_errno:`)
	headerRe  = regexp.MustCompile(`^#\s+[A-Z]`)
	metricsRe = regexp.MustCompile(`(\w+): (\S+|\z)`)
	adminRe   = regexp.MustCompile(`command: (.+)`)
	setRe     = regexp.MustCompile(`^SET (?:last_insert_id|insert_id|timestamp)`)
	useRe     = regexp.MustCompile(`^(?i)use `)
)

// ParsedEvent represents a parsed block of MySQL slow log.
type ParsedEvent struct {
	Block []string
	Event *log.Event
}

// A SlowLogParser parses a MySQL slow log.
type SlowLogParser struct {
	r    Reader
	opts log.Options

	stopErr        error
	parsedEventsCh chan *ParsedEvent
	inHeader       bool
	inQuery        bool
	headerLines    uint
	queryLines     uint64
	bytesRead      uint64
	lineOffset     uint64
	endOffset      uint64
	currentBlock   []string
	currentEvent   *log.Event
}

// NewSlowLogParser returns a new SlowLogParser that reads from the given reader.
func NewSlowLogParser(r Reader, opts log.Options) *SlowLogParser {
	if opts.StartOffset != 0 {
		panic("StartOffset is not supported")
	}

	if opts.DefaultLocation == nil {
		// Old MySQL format assumes time is taken from SYSTEM.
		opts.DefaultLocation = time.Local
	}
	p := &SlowLogParser{
		r:    r,
		opts: opts,

		parsedEventsCh: make(chan *ParsedEvent),
		inHeader:       false,
		inQuery:        false,
		headerLines:    0,
		queryLines:     0,
		lineOffset:     0,
		bytesRead:      0,
		currentEvent:   log.NewEvent(),
	}
	return p
}

// logf logs with configured logger.
func (p *SlowLogParser) logf(format string, v ...interface{}) {
	if !p.opts.Debug {
		return
	}
	if p.opts.Debugf != nil {
		p.opts.Debugf(format, v...)
		return
	}
	stdlog.Printf(format, v...)
}

// Parse returns next parsed event, or nil, when parsing is done.
func (p *SlowLogParser) Parse() *ParsedEvent {
	return <-p.parsedEventsCh
}

// Err returns a reason why parsing stop.
// It must be called only after Parse() returned nil.
func (p *SlowLogParser) Err() error {
	return p.stopErr
}

// Run parses events until readNextBlock function returns error.
// Caller should call Parse() until nil is returned, then inspect Err().
func (p *SlowLogParser) Run() {
	defer func() {
		if p.queryLines > 0 {
			p.endOffset = p.bytesRead
			p.sendEvent(false, false)
		}

		p.logf("done")
		close(p.parsedEventsCh)
	}()

	br := NewBlockReader(p.r)

	for {
		block, err := br.NextBlock()
		if err != nil {
			p.stopErr = err
			return
		}
		p.currentBlock = block

		for _, line := range block {
			lineLen := uint64(len(line))
			p.bytesRead += lineLen
			p.lineOffset = p.bytesRead - lineLen
			p.logf("+%d line: %s", p.lineOffset, line)

			// Filter out meta lines:
			//   /usr/local/bin/mysqld, Version: 5.6.15-62.0-tokudb-7.1.0-tokudb-log (binary). started with:
			//   Tcp port: 3306  Unix socket: /var/lib/mysql/mysql.sock
			//   Time                 Id Command    Argument
			if lineLen >= 20 && ((line[0] == '/' && line[lineLen-6:lineLen] == "with:\n") ||
				(line[0:5] == "Time ") ||
				(line[0:4] == "Tcp ") ||
				(line[0:4] == "TCP ")) {
				p.logf("meta")
				continue
			}

			// PMM-1834: Filter out empty comments and MariaDB explain:
			if line == "#\n" || strings.HasPrefix(line, "# explain:") {
				continue
			}

			// Remove \n.
			line = line[0 : lineLen-1]

			switch {
			case p.inHeader:
				p.parseHeader(line)
			case p.inQuery:
				p.parseQuery(line)
			case headerRe.MatchString(line):
				p.inHeader = true
				p.inQuery = false
				p.parseHeader(line)
			default:
				p.logf("unhandled line: %q", line)
			}
		}
	}
}

func (p *SlowLogParser) parseHeader(line string) {
	p.logf("header")

	if !headerRe.MatchString(line) {
		p.inHeader = false
		p.inQuery = true
		p.parseQuery(line)
		return
	}

	if p.headerLines == 0 {
		p.currentEvent.Offset = p.lineOffset
	}
	p.headerLines++

	switch {
	case strings.HasPrefix(line, "# Time"):
		p.logf("time")
		m := timeRe.FindStringSubmatch(line)
		if len(m) == 2 {
			p.currentEvent.Ts, _ = time.ParseInLocation("060102 15:04:05", m[1], p.opts.DefaultLocation)
		} else {
			m = timeNewRe.FindStringSubmatch(line)
			if len(m) == 2 {
				p.currentEvent.Ts, _ = time.ParseInLocation(time.RFC3339Nano, m[1], p.opts.DefaultLocation)
			} else {
				return
			}
		}
		if userRe.MatchString(line) {
			p.logf("user (bad format)")
			m := userRe.FindStringSubmatch(line)
			p.currentEvent.User = m[1]
			p.currentEvent.Host = m[2]
		}

	case strings.HasPrefix(line, "# User"):
		p.logf("user")
		m := userRe.FindStringSubmatch(line)
		if len(m) < 3 {
			return
		}
		p.currentEvent.User = m[1]
		p.currentEvent.Host = m[2]

	case strings.HasPrefix(line, "# admin"):
		p.parseAdmin(line)

	default:
		p.logf("metrics")
		submatch := schema.FindStringSubmatch(line)
		if len(submatch) == 2 {
			p.currentEvent.Db = submatch[1]
		}

		m := metricsRe.FindAllStringSubmatch(line, -1)
		for _, smv := range m {
			switch {
			// [String, Metric, Value], e.g. ["Query_time: 2", "Query_time", "2"]
			case strings.HasSuffix(smv[1], "_time") || strings.HasSuffix(smv[1], "_wait"):
				// microsecond value
				val, _ := strconv.ParseFloat(smv[2], 64)
				p.currentEvent.TimeMetrics[smv[1]] = val

			case smv[2] == "Yes" || smv[2] == "No":
				// boolean value
				if smv[2] == "Yes" {
					p.currentEvent.BoolMetrics[smv[1]] = true
				} else {
					p.currentEvent.BoolMetrics[smv[1]] = false
				}

			case smv[1] == "Schema":
				p.currentEvent.Db = smv[2]

			case smv[1] == "Log_slow_rate_type":
				p.currentEvent.RateType = smv[2]

			case smv[1] == "Log_slow_rate_limit":
				val, _ := strconv.ParseUint(smv[2], 10, 64)
				p.currentEvent.RateLimit = uint(val)

			default:
				// integer value
				val, _ := strconv.ParseUint(smv[2], 10, 64)
				p.currentEvent.NumberMetrics[smv[1]] = val
			}
		}
	}
}

func (p *SlowLogParser) parseQuery(line string) {
	p.logf("query")

	if strings.HasPrefix(line, "# admin") {
		p.parseAdmin(line)
		return
	}

	if headerRe.MatchString(line) {
		p.logf("next event")
		p.inHeader = true
		p.inQuery = false
		p.endOffset = p.lineOffset
		p.sendEvent(true, false)
		p.parseHeader(line)
		return
	}

	isUse := useRe.FindString(line)
	switch {
	case p.queryLines == 0 && isUse != "":
		p.logf("use db")
		db := strings.TrimPrefix(line, isUse)
		db = strings.TrimRight(db, ";")
		db = strings.Trim(db, "`")
		p.currentEvent.Db = db
		// Set the 'use' as the query itself.
		// In case we are on a group of lines like in test 23, lines 6~8, the
		// query will be replaced by the real query "select field...."
		// In case we are on a group of lines like in test23, lines 27~28, the
		// query will be "use dbnameb" since the user executed a use command
		p.currentEvent.Query = line

	case setRe.MatchString(line):
		p.logf("set var")
		// @todo ignore or use these lines?

	default:
		p.logf("query")
		if p.queryLines > 0 {
			p.currentEvent.Query += "\n" + line
		} else {
			p.currentEvent.Query = line
		}
		p.queryLines++
	}
}

func (p *SlowLogParser) parseAdmin(line string) {
	p.logf("admin")
	p.currentEvent.Admin = true
	m := adminRe.FindStringSubmatch(line)
	if m != nil {
		p.currentEvent.Query = m[1]
		p.currentEvent.Query = strings.TrimSuffix(p.currentEvent.Query, ";") // makes FilterAdminCommand work
	}

	// admin commands should be the last line of the event.
	if filtered := p.opts.FilterAdminCommand[p.currentEvent.Query]; !filtered {
		p.logf("not filtered")
		p.endOffset = p.bytesRead
		p.sendEvent(false, false)
		return
	}

	p.inHeader = false
	p.inQuery = false
}

func (p *SlowLogParser) sendEvent(inHeader bool, inQuery bool) {
	p.logf("send event")

	p.currentEvent.OffsetEnd = p.endOffset

	// Make a new event and reset our metadata.
	defer func() {
		p.currentBlock = nil
		p.currentEvent = log.NewEvent()
		p.headerLines = 0
		p.queryLines = 0
		p.inHeader = inHeader
		p.inQuery = inQuery
	}()

	if _, ok := p.currentEvent.TimeMetrics["Query_time"]; !ok {
		// Started parsing in header after Query_time.  Throw away event.
		p.logf("No Query_time in event at %d: %#v", p.lineOffset, p.currentEvent)
		return
	}

	// Clean up the event.
	p.currentEvent.Db = strings.TrimSuffix(p.currentEvent.Db, ";\n")
	p.currentEvent.Query = strings.TrimSuffix(p.currentEvent.Query, ";")

	p.parsedEventsCh <- &ParsedEvent{
		Block: p.currentBlock,
		Event: p.currentEvent,
	}
}
