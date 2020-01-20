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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	pgquery "github.com/lfittl/pg_query_go"
	pgquerynodes "github.com/lfittl/pg_query_go/nodes"
	"github.com/pkg/errors"
)

// ExtractTables extracts table names from query.
func ExtractTables(query string) (tables []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			// preserve stack
			err = errors.WithStack(fmt.Errorf("%v", r))
		}
	}()

	var jsonTree string
	if jsonTree, err = pgquery.ParseToJSON(query); err != nil {
		err = errors.Wrap(err, "error on parsing sql query")
		return
	}

	var list []json.RawMessage
	err = json.Unmarshal([]byte(jsonTree), &list)
	if err != nil {
		return
	}

	tables = []string{}
	tableNames := make(map[string]bool)
	excludedtableNames := make(map[string]bool)
	foundTables, excludeTables := extractTableNames(list...)
	for _, tableName := range excludeTables {
		if _, ok := excludedtableNames[tableName]; !ok {
			excludedtableNames[tableName] = true
		}
	}
	for _, tableName := range foundTables {
		_, tableAdded := tableNames[tableName]
		_, tableExcluded := excludedtableNames[tableName]
		if !tableAdded && !tableExcluded {
			tables = append(tables, tableName)
			tableNames[tableName] = true
		}
	}

	sort.Strings(tables)

	return
}

func extractTableNames(stmts ...json.RawMessage) ([]string, []string) {
	var tables, excludeTables []string
	for _, input := range stmts {
		if input == nil || string(input) == "null" || !(strings.HasPrefix(string(input), "{") || strings.HasPrefix(string(input), "[")) {
			continue
		}

		if strings.HasPrefix(string(input), "[") {
			var list []json.RawMessage
			err := json.Unmarshal(input, &list)
			if err != nil {
				logrus.Warn(err)
				continue
			}
			foundTables, tmpExcludeTables := extractTableNames(list...)
			tables = append(tables, foundTables...)
			excludeTables = append(excludeTables, tmpExcludeTables...)
			continue
		}

		var nodeMap map[string]json.RawMessage

		err := json.Unmarshal(input, &nodeMap)
		if err != nil {
			logrus.Warnln("couldn't decode json", err)
			continue
		}

		for nodeType, jsonText := range nodeMap {
			var foundTables, tmpExcludeTables []string
			switch nodeType {
			case "RangeVar":
				var outNode pgquerynodes.RangeVar
				err = json.Unmarshal(jsonText, &outNode)
				if err != nil {
					logrus.Warnln("couldn't decode json", err)
					continue
				}
				logrus.Debugln(*outNode.Relname)
				foundTables = []string{*outNode.Relname}
			case "CommonTableExpr":
				var ctesNodeMap map[string]json.RawMessage

				err := json.Unmarshal(jsonText, &ctesNodeMap)
				if err != nil {
					logrus.Warnln("couldn't decode json", err)
					continue
				}
				foundTables, tmpExcludeTables = extractTableNames(ctesNodeMap["ctequery"])
				var cteName string
				err = json.Unmarshal(ctesNodeMap["ctename"], &cteName)
				if err != nil {
					logrus.Warnln("couldn't decode json", err)
					continue
				}
				tmpExcludeTables = append(tmpExcludeTables, cteName)
			default:
				foundTables, tmpExcludeTables = extractTableNames(jsonText)
			}
			tables = append(tables, foundTables...)
			excludeTables = append(excludeTables, tmpExcludeTables...)
		}
	}

	return tables, excludeTables
}
