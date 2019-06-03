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

const readerBufSize = 16 * 1024

type ContinuousFileReader struct {
	filename string
	l        Logger

	// file Read/Close calls must be synchronized
	m      sync.Mutex
	closed bool
	f      *os.File
	r      *bufio.Reader

	sleep time.Duration // for testing only
}

func NewContinuousFileReader(filename string, l Logger) (*ContinuousFileReader, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}

	if _, err = f.Seek(0, io.SeekEnd); err != nil {
		l.Errorf("Failed to seek %q to the end: %s.", err)
	}

	return &ContinuousFileReader{
		filename: filename,
		l:        l,
		f:        f,
		r:        bufio.NewReaderSize(f, readerBufSize),
		sleep:    time.Second,
	}, nil
}

// NextLine implements Reader interface.
func (r *ContinuousFileReader) NextLine() (string, error) {
	r.m.Lock()
	defer r.m.Unlock()

	var line string
	for {
		l, err := r.r.ReadString('\n')
		r.l.Tracef("ReadLine: %q %v", l, err)
		line += l

		switch {
		case err == nil:
			// Full line successfully read - return it.
			return line, nil

		case r.closed:
			// If file is closed, err would be os.PathError{"read", filename, os.ErrClosed}.
			// Return io.EOF instead.
			return line, io.EOF

		case err != io.EOF:
			// Return unexpected error as is.
			return line, err

		default:
			// err is io.EOF, but reader is not closed - reset or sleep.
			needReset := r.needReset()
			if needReset {
				r.reset()
			} else {
				r.m.Unlock()
				time.Sleep(r.sleep)
				r.m.Lock()
			}
		}
	}
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
		r.l.Infof("File renamed, resetting.")
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
	r.r = bufio.NewReaderSize(f, readerBufSize)
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
	if err != nil {
		r.l.Errorf("%s", err)
		return nil
	}
	m.InputSize = fi.Size()

	pos, err := r.f.Seek(0, io.SeekCurrent)
	if err != nil {
		r.l.Errorf("%s", err)
		return nil
	}
	m.InputPos = pos

	return &m
}

// check interfaces
var (
	_ Reader = (*ContinuousFileReader)(nil)
)
