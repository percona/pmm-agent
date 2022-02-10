// Code generated by gopkg.in/reform.v1. DO NOT EDIT.

package pgstatstatements

import (
	"fmt"
	"strings"

	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

type pgStatDatabaseViewType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("pg_catalog").
func (v *pgStatDatabaseViewType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("pg_stat_database").
func (v *pgStatDatabaseViewType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *pgStatDatabaseViewType) Columns() []string {
	return []string{
		"datid",
		"datname",
	}
}

// NewStruct makes a new struct for that view or table.
func (v *pgStatDatabaseViewType) NewStruct() reform.Struct {
	return &pgStatDatabase{}
}

// pgStatDatabaseView represents pg_stat_database view or table in SQL database.
var pgStatDatabaseView = &pgStatDatabaseViewType{
	s: parse.StructInfo{
		Type:      "pgStatDatabase",
		SQLSchema: "pg_catalog",
		SQLName:   "pg_stat_database",
		Fields: []parse.FieldInfo{
			{Name: "DatID", Type: "int64", Column: "datid"},
			{Name: "DatName", Type: "*string", Column: "datname"},
		},
		PKFieldIndex: -1,
	},
	z: (&pgStatDatabase{}).Values(),
}

// String returns a string representation of this struct or record.
func (s pgStatDatabase) String() string {
	res := make([]string, 2)
	res[0] = "DatID: " + reform.Inspect(s.DatID, true)
	res[1] = "DatName: " + reform.Inspect(s.DatName, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *pgStatDatabase) Values() []interface{} {
	return []interface{}{
		s.DatID,
		s.DatName,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *pgStatDatabase) Pointers() []interface{} {
	return []interface{}{
		&s.DatID,
		&s.DatName,
	}
}

// View returns View object for that struct.
func (s *pgStatDatabase) View() reform.View {
	return pgStatDatabaseView
}

// check interfaces
var (
	_ reform.View   = pgStatDatabaseView
	_ reform.Struct = (*pgStatDatabase)(nil)
	_ fmt.Stringer  = (*pgStatDatabase)(nil)
)

type pgUserViewType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("pg_catalog").
func (v *pgUserViewType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("pg_user").
func (v *pgUserViewType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *pgUserViewType) Columns() []string {
	return []string{
		"usesysid",
		"usename",
	}
}

// NewStruct makes a new struct for that view or table.
func (v *pgUserViewType) NewStruct() reform.Struct {
	return &pgUser{}
}

// pgUserView represents pg_user view or table in SQL database.
var pgUserView = &pgUserViewType{
	s: parse.StructInfo{
		Type:      "pgUser",
		SQLSchema: "pg_catalog",
		SQLName:   "pg_user",
		Fields: []parse.FieldInfo{
			{Name: "UserID", Type: "int64", Column: "usesysid"},
			{Name: "UserName", Type: "*string", Column: "usename"},
		},
		PKFieldIndex: -1,
	},
	z: (&pgUser{}).Values(),
}

// String returns a string representation of this struct or record.
func (s pgUser) String() string {
	res := make([]string, 2)
	res[0] = "UserID: " + reform.Inspect(s.UserID, true)
	res[1] = "UserName: " + reform.Inspect(s.UserName, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *pgUser) Values() []interface{} {
	return []interface{}{
		s.UserID,
		s.UserName,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *pgUser) Pointers() []interface{} {
	return []interface{}{
		&s.UserID,
		&s.UserName,
	}
}

// View returns View object for that struct.
func (s *pgUser) View() reform.View {
	return pgUserView
}

// check interfaces
var (
	_ reform.View   = pgUserView
	_ reform.Struct = (*pgUser)(nil)
	_ fmt.Stringer  = (*pgUser)(nil)
)

type pgStatStatementsViewType struct {
	s parse.StructInfo
	z []interface{}
}

// Schema returns a schema name in SQL database ("").
func (v *pgStatStatementsViewType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("pg_stat_statements").
func (v *pgStatStatementsViewType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *pgStatStatementsViewType) Columns() []string {
	return []string{
		"userid",
		"dbid",
		"queryid",
		"query",
		"calls",
		"total_time",
		"rows",
		"shared_blks_hit",
		"shared_blks_read",
		"shared_blks_dirtied",
		"shared_blks_written",
		"local_blks_hit",
		"local_blks_read",
		"local_blks_dirtied",
		"local_blks_written",
		"temp_blks_read",
		"temp_blks_written",
		"blk_read_time",
		"blk_write_time",
	}
}

// NewStruct makes a new struct for that view or table.
func (v *pgStatStatementsViewType) NewStruct() reform.Struct {
	return &pgStatStatements{}
}

// pgStatStatementsView represents pg_stat_statements view or table in SQL database.
var pgStatStatementsView = &pgStatStatementsViewType{
	s: parse.StructInfo{
		Type:    "pgStatStatements",
		SQLName: "pg_stat_statements",
		Fields: []parse.FieldInfo{
			{Name: "UserID", Type: "int64", Column: "userid"},
			{Name: "DBID", Type: "int64", Column: "dbid"},
			{Name: "QueryID", Type: "int64", Column: "queryid"},
			{Name: "Query", Type: "string", Column: "query"},
			{Name: "Calls", Type: "int64", Column: "calls"},
			{Name: "TotalTime", Type: "float64", Column: "total_time"},
			{Name: "Rows", Type: "int64", Column: "rows"},
			{Name: "SharedBlksHit", Type: "int64", Column: "shared_blks_hit"},
			{Name: "SharedBlksRead", Type: "int64", Column: "shared_blks_read"},
			{Name: "SharedBlksDirtied", Type: "int64", Column: "shared_blks_dirtied"},
			{Name: "SharedBlksWritten", Type: "int64", Column: "shared_blks_written"},
			{Name: "LocalBlksHit", Type: "int64", Column: "local_blks_hit"},
			{Name: "LocalBlksRead", Type: "int64", Column: "local_blks_read"},
			{Name: "LocalBlksDirtied", Type: "int64", Column: "local_blks_dirtied"},
			{Name: "LocalBlksWritten", Type: "int64", Column: "local_blks_written"},
			{Name: "TempBlksRead", Type: "int64", Column: "temp_blks_read"},
			{Name: "TempBlksWritten", Type: "int64", Column: "temp_blks_written"},
			{Name: "BlkReadTime", Type: "float64", Column: "blk_read_time"},
			{Name: "BlkWriteTime", Type: "float64", Column: "blk_write_time"},
		},
		PKFieldIndex: -1,
	},
	z: (&pgStatStatements{}).Values(),
}

// String returns a string representation of this struct or record.
func (s pgStatStatements) String() string {
	res := make([]string, 19)
	res[0] = "UserID: " + reform.Inspect(s.UserID, true)
	res[1] = "DBID: " + reform.Inspect(s.DBID, true)
	res[2] = "QueryID: " + reform.Inspect(s.QueryID, true)
	res[3] = "Query: " + reform.Inspect(s.Query, true)
	res[4] = "Calls: " + reform.Inspect(s.Calls, true)
	res[5] = "TotalTime: " + reform.Inspect(s.TotalTime, true)
	res[6] = "Rows: " + reform.Inspect(s.Rows, true)
	res[7] = "SharedBlksHit: " + reform.Inspect(s.SharedBlksHit, true)
	res[8] = "SharedBlksRead: " + reform.Inspect(s.SharedBlksRead, true)
	res[9] = "SharedBlksDirtied: " + reform.Inspect(s.SharedBlksDirtied, true)
	res[10] = "SharedBlksWritten: " + reform.Inspect(s.SharedBlksWritten, true)
	res[11] = "LocalBlksHit: " + reform.Inspect(s.LocalBlksHit, true)
	res[12] = "LocalBlksRead: " + reform.Inspect(s.LocalBlksRead, true)
	res[13] = "LocalBlksDirtied: " + reform.Inspect(s.LocalBlksDirtied, true)
	res[14] = "LocalBlksWritten: " + reform.Inspect(s.LocalBlksWritten, true)
	res[15] = "TempBlksRead: " + reform.Inspect(s.TempBlksRead, true)
	res[16] = "TempBlksWritten: " + reform.Inspect(s.TempBlksWritten, true)
	res[17] = "BlkReadTime: " + reform.Inspect(s.BlkReadTime, true)
	res[18] = "BlkWriteTime: " + reform.Inspect(s.BlkWriteTime, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *pgStatStatements) Values() []interface{} {
	return []interface{}{
		s.UserID,
		s.DBID,
		s.QueryID,
		s.Query,
		s.Calls,
		s.TotalTime,
		s.Rows,
		s.SharedBlksHit,
		s.SharedBlksRead,
		s.SharedBlksDirtied,
		s.SharedBlksWritten,
		s.LocalBlksHit,
		s.LocalBlksRead,
		s.LocalBlksDirtied,
		s.LocalBlksWritten,
		s.TempBlksRead,
		s.TempBlksWritten,
		s.BlkReadTime,
		s.BlkWriteTime,
	}
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *pgStatStatements) Pointers() []interface{} {
	return []interface{}{
		&s.UserID,
		&s.DBID,
		&s.QueryID,
		&s.Query,
		&s.Calls,
		&s.TotalTime,
		&s.Rows,
		&s.SharedBlksHit,
		&s.SharedBlksRead,
		&s.SharedBlksDirtied,
		&s.SharedBlksWritten,
		&s.LocalBlksHit,
		&s.LocalBlksRead,
		&s.LocalBlksDirtied,
		&s.LocalBlksWritten,
		&s.TempBlksRead,
		&s.TempBlksWritten,
		&s.BlkReadTime,
		&s.BlkWriteTime,
	}
}

// View returns View object for that struct.
func (s *pgStatStatements) View() reform.View {
	return pgStatStatementsView
}

// check interfaces
var (
	_ reform.View   = pgStatStatementsView
	_ reform.Struct = (*pgStatStatements)(nil)
	_ fmt.Stringer  = (*pgStatStatements)(nil)
)

func init() {
	parse.AssertUpToDate(&pgStatDatabaseView.s, &pgStatDatabase{})
	parse.AssertUpToDate(&pgUserView.s, &pgUser{})
	parse.AssertUpToDate(&pgStatStatementsView.s, &pgStatStatements{})
}
