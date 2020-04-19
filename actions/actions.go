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

// Package actions provides Actions implementations and runner.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
)

//go-sumtype:decl Action

// Action describes an abstract thing that can be run by a client and return some output.
type Action interface {
	// ID returns an Action ID.
	ID() string
	// Type returns an Action type.
	Type() string
	// Run runs an Action and returns output and error.
	Run(ctx context.Context) ([]byte, error)

	sealed()
}

type readRowsParams struct {
	// do not convert []byte to string
	keepBytes bool
}

// readRows reads and closes given *sql.Rows, returning columns, data rows, and first encountered error.
func readRows(rows *sql.Rows, params *readRowsParams) (columns []string, dataRows [][]interface{}, err error) {
	if params == nil {
		params = new(readRowsParams)
	}

	defer func() {
		// overwrite err with e only if err does not already contains (more interesting) error
		if e := rows.Close(); err == nil {
			err = e
		}
	}()

	columns, err = rows.Columns()
	if err != nil {
		return
	}

	for rows.Next() {
		dest := make([]interface{}, len(columns))
		for i := range dest {
			var ei interface{}
			dest[i] = &ei
		}
		if err = rows.Scan(dest...); err != nil {
			return
		}

		// Each dest element is an *interface{} (&ei above) which always contain some typed data
		// (in particular, it can contain typed nil). Dereference it for easier manipulations by the caller.
		for i, d := range dest {
			ei := *(d.(*interface{}))
			dest[i] = ei

			// As a special case, convert []byte values to strings unless caller requested to keep []byte.
			// QAN Actions prefer strings to avoid base64 output from jsonRows;
			// Query Actions prefer []byte to avoid `proto: field XXX contains invalid UTF-8` error from protobuf.
			if b, ok := (ei).([]byte); ok && !params.keepBytes {
				dest[i] = string(b)
			}
		}

		dataRows = append(dataRows, dest)
	}
	err = rows.Err()
	return //nolint:nakedret
}

func convertRows(columns []string, dataRows [][]interface{}) []map[string]interface{} {
	res := make([]map[string]interface{}, len(dataRows))
	for i, row := range dataRows {
		resRow := make(map[string]interface{}, len(columns))

		for j, column := range columns {
			resRow[column] = row[j]
		}

		res[i] = resRow
	}

	return res
}

// jsonRows converts input to JSON array:
// [
//   ["column 1", "column 2", …],
//   ["value 1", 2, …]
//   …
// ]
func jsonRows(columns []string, dataRows [][]interface{}) ([]byte, error) {
	res := make([][]interface{}, len(dataRows)+1)

	res[0] = make([]interface{}, len(columns))
	for i, col := range columns {
		res[0][i] = col
	}

	for i, row := range dataRows {
		res[i+1] = make([]interface{}, len(columns))
		copy(res[i+1], row)
	}

	return json.Marshal(res)
}
