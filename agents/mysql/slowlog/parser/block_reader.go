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
	"io"
	"strings"
)

type BlockReader struct {
	r        Reader
	lastLine string
}

func NewBlockReader(r Reader) *BlockReader {
	return &BlockReader{
		r: r,
	}
}

// NextBlock reads the whole block using reader's NextLine method.
func (br *BlockReader) NextBlock() (block []string, err error) {
	// TODO optimize, cleanup

	defer func() {
		if err == io.EOF && block != nil {
			err = nil
		}
	}()

	var l string

	// skip over startup messages until headers at the beginning of the next block
	for !strings.HasPrefix(l, "#") {
		if l, err = br.nextLine(); err != nil {
			return
		}
	}

	inHeaders := true
	for {
		block = append(block, l)

		if l, err = br.nextLine(); err != nil {
			return
		}

		if strings.HasPrefix(l, "# administrator") {
			block = append(block, l)
			return
		}

		switch {
		case inHeaders && strings.HasPrefix(l, "#"):
			// continue reading headers
			continue

		case inHeaders && !strings.HasPrefix(l, "#"):
			// start reading query
			inHeaders = false
			continue

		case !inHeaders && strings.HasPrefix(l, "#"):
			br.lastLine = l
			return

		case !inHeaders && !strings.HasPrefix(l, "#"):
			continue
		}
	}
}

func (br *BlockReader) nextLine() (string, error) {
	if l := br.lastLine; l != "" {
		br.lastLine = ""
		return l, nil
	}

	return br.r.NextLine()
}
