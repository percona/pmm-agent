package parser

import (
	"reflect"

	pgquery "github.com/lfittl/pg_query_go"
	pgquerynodes "github.com/lfittl/pg_query_go/nodes"
)

type Parser struct{}

// ExtractTables extracts table names from query.
func (p *Parser) ExtractTables(query string) (tables []string, err error) {
	tree, err := pgquery.Parse(query)
	if err != nil {
		return nil, err
	}
	tables = []string{}
	tableNames := make(map[string]bool)
	excludedtableNames := make(map[string]bool)
	for _, stmt := range tree.Statements {
		foundTables, excludeTables := p.extractTableNames(stmt)
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
	}

	return
}

func (p *Parser) extractTableNames(stmts ...pgquerynodes.Node) (tables, excludeTables []string) {
	for _, stmt := range stmts {
		if isNilValue(stmt) {
			continue
		}
		var foundTables []string
		var tmpExcludeTables []string
		switch v := stmt.(type) {
		case pgquerynodes.RawStmt:
			return p.extractTableNames(v.Stmt)
		case pgquerynodes.SelectStmt: // Select queries
			foundTables, tmpExcludeTables = p.extractTableNames(v.FromClause, v.WhereClause, v.WithClause, v.Larg, v.Rarg)
		case pgquerynodes.InsertStmt: // Insert queries
			foundTables, tmpExcludeTables = p.extractTableNames(v.Relation, v.SelectStmt, v.WithClause)
		case pgquerynodes.UpdateStmt: // Update queries
			foundTables, tmpExcludeTables = p.extractTableNames(v.Relation, v.FromClause, v.WhereClause, v.WithClause)
		case pgquerynodes.DeleteStmt: // Delete queries
			foundTables, tmpExcludeTables = p.extractTableNames(v.Relation, v.WhereClause, v.WithClause)

		case pgquerynodes.JoinExpr: // Joins
			foundTables, tmpExcludeTables = p.extractTableNames(v.Larg, v.Rarg)

		case pgquerynodes.RangeVar: // Table name
			foundTables = []string{*v.Relname}

		case pgquerynodes.List:
			foundTables, tmpExcludeTables = p.extractTableNames(v.Items...)

		case pgquerynodes.WithClause: // To exclude temporary tables
			foundTables, tmpExcludeTables = p.extractTableNames(v.Ctes)
			for _, item := range v.Ctes.Items {
				switch cte := item.(type) {
				case pgquerynodes.CommonTableExpr:
					tmpExcludeTables = append(tmpExcludeTables, *cte.Ctename)
				}
			}

		case pgquerynodes.A_Expr: // Where a=b
			foundTables, tmpExcludeTables = p.extractTableNames(v.Lexpr, v.Rexpr)

		// Subqueries
		case pgquerynodes.SubLink:
			foundTables, tmpExcludeTables = p.extractTableNames(v.Subselect, v.Xpr, v.Testexpr)
		case pgquerynodes.RangeSubselect:
			foundTables, tmpExcludeTables = p.extractTableNames(v.Subquery)
		case pgquerynodes.CommonTableExpr:
			foundTables, tmpExcludeTables = p.extractTableNames(v.Ctequery)

		default:
			if isPointer(v) { // to avoid duplications in case of pointers
				switch deref := reflect.ValueOf(v).Elem().Interface().(type) {
				case pgquerynodes.Node:
					foundTables, tmpExcludeTables = p.extractTableNames(deref)
				}
			}
		}
		tables = append(tables, foundTables...)
		excludeTables = append(excludeTables, tmpExcludeTables...)
	}

	return
}

func isNilValue(i interface{}) bool {
	return i == nil || (isPointer(i) && reflect.ValueOf(i).IsNil())
}
func isPointer(v interface{}) bool {
	return reflect.ValueOf(v).Kind() == reflect.Ptr
}
