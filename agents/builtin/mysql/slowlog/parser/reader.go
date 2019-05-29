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

type reader interface {
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

type ContinuousReader struct {
	r *bufio.Reader

	m sync.Mutex
	f *os.File
}

func NewContinuousReader(filename string) (*ContinuousReader, error) {
	f, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return nil, err
	}
	return &ContinuousReader{
		r: bufio.NewReader(f),
		f: f,
	}, nil
}

func (r *ContinuousReader) NextLine() (string, error) {
	r.m.Lock()
	l, err := r.r.ReadString('\n')
	r.m.Unlock()
	return l, err
}

func (r *ContinuousReader) Close() error {
	r.m.Lock()
	err := r.f.Close()
	r.m.Unlock()
	return err
}
