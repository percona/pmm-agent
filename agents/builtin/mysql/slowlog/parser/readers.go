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
	"os"
	"sync"
)

type Reader interface {
	// NextLine reads full lines from the source and returns them (including the last '\n').
	// If the full line can't be read because of EOF, reader implementation may decide to block and
	// wait for new data to arrive. Other errors should be returned without blocking.
	NextLine() (string, error)

	Close() error
}

type SimpleFileReader struct {
	r *bufio.Reader

	m sync.Mutex
	f *os.File
}

func NewSimpleReader(filename string) (*SimpleFileReader, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return &SimpleFileReader{
		r: bufio.NewReader(f),
		f: f,
	}, nil
}

func (r *SimpleFileReader) NextLine() (string, error) {
	r.m.Lock()
	l, err := r.r.ReadString('\n')
	r.m.Unlock()
	return l, err
}

func (r *SimpleFileReader) Close() error {
	r.m.Lock()
	err := r.f.Close()
	r.m.Unlock()
	return err
}

type ContinuousFileReader struct {
	r *bufio.Reader

	m sync.Mutex
	f *os.File
}

func NewContinuousFileReader(filename string) (*ContinuousFileReader, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return &ContinuousFileReader{
		r: bufio.NewReader(f),
		f: f,
	}, nil
}

func (r *ContinuousFileReader) NextLine() (string, error) {
	r.m.Lock()
	l, err := r.r.ReadString('\n')
	r.m.Unlock()
	return l, err
}

func (r *ContinuousFileReader) Close() error {
	r.m.Lock()
	err := r.f.Close()
	r.m.Unlock()
	return err
}
