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

package parser

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"
	"time"
)

type ContinuousFileReader struct {
	filename string
	l        Logger

	// file Read/Close calls must be synchronized
	m      sync.Mutex
	closed bool
	f      *os.File
	r      *bufio.Reader
}

func NewContinuousFileReader(filename string, l Logger) (*ContinuousFileReader, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}

	r := &ContinuousFileReader{
		filename: filename,
		l:        l,
		f:        f,
		r:        bufio.NewReader(f),
	}

	for err == nil {
		_, err = r.readLine()
	}

	return r, nil
}

// NextLine implements Reader interface.
func (r *ContinuousFileReader) NextLine() (string, error) {
	for {
		r.m.Lock()
		l, err := r.readLine()
		r.m.Unlock()

		r.l.Tracef("readLine: %q %v", l, err)
		if l != "" || err != nil {
			return l, err
		}

		r.m.Lock()
		needReset := r.needReset()
		if needReset {
			r.reset()
		}
		r.m.Unlock()

		if !needReset {
			time.Sleep(time.Second)
		}
	}
}

func (r *ContinuousFileReader) readLine() (string, error) {
	// TODO there?
	if r.closed {
		return "", io.EOF
	}

	l, err := r.r.ReadString('\n')
	if err == io.EOF {
		if l != "" {
			// FIXME handle this
			panic("partial read: " + l)
		}
		err = nil
	}
	return l, err
}

func (r *ContinuousFileReader) needReset() bool {
	oldFI, err := r.f.Stat()
	if err != nil {
		r.l.Errorf("%s", err)
	}
	newFI, err := os.Stat(r.filename)
	if err != nil {
		r.l.Errorf("%s", err)
	}
	if !os.SameFile(oldFI, newFI) {
		r.l.Infof("File changed, resetting.")
		return true
	}

	oldPos, err := r.f.Seek(0, io.SeekCurrent)
	r.l.Tracef("Old file pos: %d, %v.", oldPos, err)
	if oldPos > newFI.Size() {
		r.l.Infof("File truncated, resetting.")
		return true
	}

	// TODO handle symlinks

	r.l.Tracef("File not changed.")
	return false
}

func (r *ContinuousFileReader) reset() {
	if err := r.f.Close(); err != nil {
		r.l.Warnf("Failed to close %s: %s.", r.f.Name(), err)
	}

	f, err := os.Open(r.filename)
	if err != nil {
		r.l.Errorf("Failed to open %s: %s. Closing reader.", r.filename, err)
		r.r = bufio.NewReader(bytes.NewReader(nil))
		r.closed = true
		return
	}

	r.f = f
	r.r = bufio.NewReader(f)
}

// Close implements Reader interface.
func (r *ContinuousFileReader) Close() error {
	r.m.Lock()
	defer r.m.Unlock()

	err := r.f.Close()
	r.closed = true
	return err
}

// Metrics implements Reader interface.
func (r *ContinuousFileReader) Metrics() *ReaderMetrics {
	r.m.Lock()
	defer r.m.Unlock()

	var m ReaderMetrics
	fi, err := r.f.Stat()
	if err == nil {
		m.InputSize = fi.Size()
	}
	pos, err := r.f.Seek(0, io.SeekCurrent)
	if err == nil {
		m.InputPos = pos
	}
	return &m
}

// check interfaces
var (
	_ Reader = (*ContinuousFileReader)(nil)
)
