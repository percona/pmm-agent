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

		nodeMap, err := parseNodeMap(input)
		if err != nil {
			continue
		}

		for nodeType, jsonText := range nodeMap {
			if jsonText == nil || string(jsonText) == "null" {
				continue
			}
			var foundTables, tmpExcludeTables []string
			if nodeType == "RangeVar" {
				var outNode pgquerynodes.RangeVar
				err = json.Unmarshal(jsonText, &outNode)
				if err != nil {
					logrus.Warnln("couldn't decode json", err)
					continue
				}
				tables = append(tables, *outNode.Relname)
				continue
			} else if nodeType == "List" {
				foundTables, tmpExcludeTables = extractTableNames(jsonText)
			} else {
				nm, err := parseNodeMap(jsonText)
				if err != nil {
					continue
				}
				switch nodeType {
				case "RangeVar":
				case "CommonTableExpr":
					foundTables, tmpExcludeTables = extractTableNames(nm["ctequery"])
					cteName := string(nm["ctename"])
					cteName = strings.TrimPrefix(cteName, `"`)
					cteName = strings.TrimSuffix(cteName, `"`)
					tmpExcludeTables = append(tmpExcludeTables, cteName)

				case "RawStmt":
					foundTables, tmpExcludeTables = extractTableNames(nm["stmt"])
				case "SelectStmt":
					foundTables, tmpExcludeTables = extractTableNames(nm["fromClause"], nm["whereClause"], nm["withClause"], nm["larg"], nm["rarg"])
				case "InsertStmt":
					foundTables, tmpExcludeTables = extractTableNames(nm["relation"], nm["selectStmt"], nm["withClause"])
				case "UpdateStmt":
					foundTables, tmpExcludeTables = extractTableNames(nm["relation"], nm["fromClause"], nm["whereClause"], nm["withClause"])
				case "DeleteStmt":
					foundTables, tmpExcludeTables = extractTableNames(nm["relation"], nm["whereClause"], nm["withClause"])

				case "JoinExpr":
					foundTables, tmpExcludeTables = extractTableNames(nm["larg"], nm["rarg"])

				case "WithClause":
					foundTables, tmpExcludeTables = extractTableNames(nm["ctes"])
				case "A_Expr":
					foundTables, tmpExcludeTables = extractTableNames(nm["lexpr"], nm["rexpr"])

				//Subqueries
				case "SubLink":
					foundTables, tmpExcludeTables = extractTableNames(nm["subselect"], nm["xpr"], nm["testexpr"])
				case "RangeSubselect":
					foundTables, tmpExcludeTables = extractTableNames(nm["subquery"])
				}
			}
			tables = append(tables, foundTables...)
			excludeTables = append(excludeTables, tmpExcludeTables...)
		}
	}

	return tables, excludeTables
}

func parseNodeMap(jsonText json.RawMessage) (map[string]json.RawMessage, error) {
	var nm map[string]json.RawMessage
	err := json.Unmarshal(jsonText, &nm)
	if err != nil {
		logrus.Warnln("couldn't decode json", err)
		return nil, err
	}
	return nm, err
}
