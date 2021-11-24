package pgstatmonitor

/*
pgDefault
	Rows              int64          `reform:"rows"`
pg 0.8
	BucketStartTime   string         `reform:"bucket_start_time"`
	User              string         `reform:"userid"`
	DatName           string         `reform:"datname"`
pg 0.9
	BucketStartTime   string         `reform:"bucket_start_time"`
	User              string         `reform:"userid"`
	DatName           string         `reform:"datname"`
	QueryID           string         `reform:"queryid"`
	TopQueryID        *string        `reform:"top_queryid"`
	Query             string         `reform:"query"`
	PlanID            *string        `reform:"planid"`
	QueryPlan         *string        `reform:"query_plan"`
	TopQuery          *string        `reform:"top_query"`
	ApplicationName   *string        `reform:"application_name"`
	Relations         pq.StringArray `reform:"relations"`
	CmdType           int32          `reform:"cmd_type"`
	CmdTypeText       string         `reform:"cmd_type_text"`
	Elevel            int32          `reform:"elevel"`
	Sqlcode           *string        `reform:"sqlcode"`
	Message           *string        `reform:"message"`
	MinTime           float64        `reform:"min_time"`
	MaxTime           float64        `reform:"max_time"`
	MeanTime          float64        `reform:"mean_time"`
	StddevTime        float64        `reform:"stddev_time"`
	RowsRetrieved     int64          `reform:"rows_retrieved"`
	PlansCalls        int64          `reform:"plans_calls"`
	PlanTotalTime     float64        `reform:"plan_total_time"`
	PlanMinTime       float64        `reform:"plan_min_time"`
	PlanMaxTime       float64        `reform:"plan_max_time"`
	PlanMeanTime      float64        `reform:"plan_mean_time"`
	WalRecords        int64          `reform:"wal_records"`
	WalFpi            int64          `reform:"wal_fpi"`
	WalBytes          int64          `reform:"wal_bytes"`

	// state_code = 0 state 'PARSING'
	// state_code = 1 state 'PLANNING'
	// state_code = 2 state 'ACTIVE'
	// state_code = 3 state 'FINISHED'
	// state_code = 4 state 'FINISHED WITH ERROR'
	StateCode int64 `reform:"state_code"`

	State string `reform:"state"`
pg 1.0 - 11.0-12.0
removed
	PlanTotalTime     float64        `reform:"plan_total_time"`
	PlanMinTime       float64        `reform:"plan_min_time"`
	PlanMaxTime       float64        `reform:"plan_max_time"`
	PlanMeanTime      float64        `reform:"plan_mean_time"`
pg 1.0 - 13.0

*/

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/lib/pq"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/parse"
)

var (
	v10 = version.Must(version.NewVersion("1.0.0-beta-2"))
	v09 = version.Must(version.NewVersion("0.9"))
	v08 = version.Must(version.NewVersion("0.8"))
)

// pgStatMonitor represents a row in pg_stat_monitor
// view in version lower than 0.8.
type pgStatMonitor struct {
	//common
	Bucket            int64
	BucketStartTime   time.Time
	ClientIP          string
	QueryID           string // we select only non-NULL rows
	Query             string // we select only non-NULL rows
	Relations         pq.StringArray
	Calls             int64
	SharedBlksHit     int64
	SharedBlksRead    int64
	SharedBlksDirtied int64
	SharedBlksWritten int64
	LocalBlksHit      int64
	LocalBlksRead     int64
	LocalBlksDirtied  int64
	LocalBlksWritten  int64
	TempBlksRead      int64
	TempBlksWritten   int64
	BlkReadTime       float64
	BlkWriteTime      float64
	RespCalls         pq.StringArray
	CPUUserTime       float64
	CPUSysTime        float64

	// < pg0.6
	DBID   int64
	UserID int64

	// >= pg0.8
	DatName               string
	UserName              string
	BucketStartTimeString string

	// < pg 0.9
	Rows int64

	// pg 0.9
	//RowsRetrieved   int64
	TopQueryID      *string
	PlanID          *string
	QueryPlan       *string
	TopQuery        *string
	ApplicationName *string
	CmdType         int32
	CmdTypeText     string
	Elevel          int32
	Sqlcode         *string
	Message         *string
	MinTime         float64
	MaxTime         float64
	MeanTime        float64
	StddevTime      float64
	PlansCalls      int64
	PlanTotalTime   float64
	PlanMinTime     float64
	PlanMaxTime     float64
	PlanMeanTime    float64
	WalRecords      int64
	WalFpi          int64
	WalBytes        int64
	// state_code = 0 state 'PARSING'
	// state_code = 1 state 'PLANNING'
	// state_code = 2 state 'ACTIVE'
	// state_code = 3 state 'FINISHED'
	// state_code = 4 state 'FINISHED WITH ERROR'
	StateCode int64
	State     string

	// < pg 1.0
	TotalTime float64

	pointers []interface{}
	view     reform.View
}

//
//func (s pgStatMonitor) ToPgStatMonitor() pgStatMonitor {
//	return pgStatMonitor{
//		Bucket:            s.Bucket,
//		BucketStartTime:   s.BucketStartTime,
//		UserID:            s.UserID,
//		DBID:              s.DBID,
//		QueryID:           s.QueryID,
//		Query:             s.Query,
//		Calls:             s.Calls,
//		TotalTime:         s.TotalTime,
//		Rows:              s.RowsRetrieved,
//		SharedBlksHit:     s.SharedBlksHit,
//		SharedBlksRead:    s.SharedBlksRead,
//		SharedBlksDirtied: s.SharedBlksDirtied,
//		SharedBlksWritten: s.SharedBlksWritten,
//		LocalBlksHit:      s.LocalBlksHit,
//		LocalBlksRead:     s.LocalBlksRead,
//		LocalBlksDirtied:  s.LocalBlksDirtied,
//		LocalBlksWritten:  s.LocalBlksWritten,
//		TempBlksRead:      s.TempBlksRead,
//		TempBlksWritten:   s.TempBlksWritten,
//		BlkReadTime:       s.BlkReadTime,
//		BlkWriteTime:      s.BlkWriteTime,
//		ClientIP:          s.ClientIP,
//		RespCalls:         s.RespCalls,
//		CPUUserTime:       s.CPUUserTime,
//		CPUSysTime:        s.CPUSysTime,
//		Relations:         s.Relations,
//		Elevel:            s.Elevel,
//		CmdType:           s.CmdType,
//	}
//}

type Field struct {
	info    parse.FieldInfo
	pointer interface{}
}

func NewPgStatMonitorStructs(v pgStatMonitorVersion) (*pgStatMonitor, reform.View) {
	s := &pgStatMonitor{}
	fields := []Field{
		{info: parse.FieldInfo{Name: "Bucket", Type: "int64", Column: "bucket"}, pointer: &s.Bucket},
		{info: parse.FieldInfo{Name: "ClientIP", Type: "string", Column: "client_ip"}, pointer: &s.ClientIP},
		{info: parse.FieldInfo{Name: "QueryID", Type: "string", Column: "queryid"}, pointer: &s.QueryID},
		{info: parse.FieldInfo{Name: "Query", Type: "string", Column: "query"}, pointer: &s.Query},
		{info: parse.FieldInfo{Name: "Relations", Type: "pq.StringArray", Column: "relations"}, pointer: &s.Relations},
		{info: parse.FieldInfo{Name: "Calls", Type: "int64", Column: "calls"}, pointer: &s.Calls},
		{info: parse.FieldInfo{Name: "SharedBlksHit", Type: "int64", Column: "shared_blks_hit"}, pointer: &s.SharedBlksHit},
		{info: parse.FieldInfo{Name: "SharedBlksRead", Type: "int64", Column: "shared_blks_read"}, pointer: &s.SharedBlksRead},
		{info: parse.FieldInfo{Name: "SharedBlksDirtied", Type: "int64", Column: "shared_blks_dirtied"}, pointer: &s.SharedBlksDirtied},
		{info: parse.FieldInfo{Name: "SharedBlksWritten", Type: "int64", Column: "shared_blks_written"}, pointer: &s.SharedBlksWritten},
		{info: parse.FieldInfo{Name: "LocalBlksHit", Type: "int64", Column: "local_blks_hit"}, pointer: &s.LocalBlksHit},
		{info: parse.FieldInfo{Name: "LocalBlksRead", Type: "int64", Column: "local_blks_read"}, pointer: &s.LocalBlksRead},
		{info: parse.FieldInfo{Name: "LocalBlksDirtied", Type: "int64", Column: "local_blks_dirtied"}, pointer: &s.LocalBlksDirtied},
		{info: parse.FieldInfo{Name: "LocalBlksWritten", Type: "int64", Column: "local_blks_written"}, pointer: &s.LocalBlksWritten},
		{info: parse.FieldInfo{Name: "TempBlksRead", Type: "int64", Column: "temp_blks_read"}, pointer: &s.TempBlksRead},
		{info: parse.FieldInfo{Name: "TempBlksWritten", Type: "int64", Column: "temp_blks_written"}, pointer: &s.TempBlksWritten},
		{info: parse.FieldInfo{Name: "BlkReadTime", Type: "float64", Column: "blk_read_time"}, pointer: &s.BlkReadTime},
		{info: parse.FieldInfo{Name: "BlkWriteTime", Type: "float64", Column: "blk_write_time"}, pointer: &s.BlkWriteTime},
		{info: parse.FieldInfo{Name: "RespCalls", Type: "pq.StringArray", Column: "resp_calls"}, pointer: &s.RespCalls},
		{info: parse.FieldInfo{Name: "CPUUserTime", Type: "float64", Column: "cpu_user_time"}, pointer: &s.CPUUserTime},
		{info: parse.FieldInfo{Name: "CPUSysTime", Type: "float64", Column: "cpu_sys_time"}, pointer: &s.CPUSysTime},
	}

	if v == pgStatMonitorVersion06 {
		// versions older than 0.8
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "DBID", Type: "int64", Column: "dbid"}, pointer: &s.DBID},
			Field{info: parse.FieldInfo{Name: "UserID", Type: "int64", Column: "userid"}, pointer: &s.UserID},
			Field{info: parse.FieldInfo{Name: "Rows", Type: "int64", Column: "rows"}, pointer: &s.Rows},
			Field{info: parse.FieldInfo{Name: "BucketStartTime", Type: "time.Time", Column: "bucket_start_time"}, pointer: &s.BucketStartTime},
		)
	}
	if v == pgStatMonitorVersion08 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "Rows", Type: "int64", Column: "rows"}, pointer: &s.Rows},
		)
	}
	if v >= pgStatMonitorVersion08 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "DatName", Type: "string", Column: "datname"}, pointer: &s.DatName},
			Field{info: parse.FieldInfo{Name: "UserName", Type: "string", Column: "userid"}, pointer: &s.UserName},
			Field{info: parse.FieldInfo{Name: "BucketStartTimeString", Type: "string", Column: "bucket_start_time"}, pointer: &s.BucketStartTimeString},
		)
	}
	if v == pgStatMonitorVersion09 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "MinTime", Type: "float64", Column: "min_time"}, pointer: &s.MinTime},
			Field{info: parse.FieldInfo{Name: "MaxTime", Type: "float64", Column: "max_time"}, pointer: &s.MaxTime},
			Field{info: parse.FieldInfo{Name: "MeanTime", Type: "float64", Column: "mean_time"}, pointer: &s.MeanTime},
			Field{info: parse.FieldInfo{Name: "StddevTime", Type: "float64", Column: "stddev_time"}, pointer: &s.StddevTime},
			Field{info: parse.FieldInfo{Name: "PlanTotalTime", Type: "float64", Column: "plan_total_time"}, pointer: &s.PlanTotalTime},
			Field{info: parse.FieldInfo{Name: "PlanMinTime", Type: "float64", Column: "plan_min_time"}, pointer: &s.PlanMinTime},
			Field{info: parse.FieldInfo{Name: "PlanMaxTime", Type: "float64", Column: "plan_max_time"}, pointer: &s.PlanMaxTime},
			Field{info: parse.FieldInfo{Name: "PlanMeanTime", Type: "float64", Column: "plan_mean_time"}, pointer: &s.PlanMeanTime},
			Field{info: parse.FieldInfo{Name: "PlansCalls", Type: "int64", Column: "plans_calls"}, pointer: &s.PlansCalls},
		)
	}
	if v >= pgStatMonitorVersion09 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "Rows", Type: "int64", Column: "rows_retrieved"}, pointer: &s.Rows},
			Field{info: parse.FieldInfo{Name: "TopQueryID", Type: "*string", Column: "top_queryid"}, pointer: &s.TopQueryID},
			Field{info: parse.FieldInfo{Name: "PlanID", Type: "*string", Column: "planid"}, pointer: &s.PlanID},
			Field{info: parse.FieldInfo{Name: "QueryPlan", Type: "*string", Column: "query_plan"}, pointer: &s.QueryPlan},
			Field{info: parse.FieldInfo{Name: "TopQuery", Type: "*string", Column: "top_query"}, pointer: &s.TopQuery},
			Field{info: parse.FieldInfo{Name: "ApplicationName", Type: "*string", Column: "application_name"}, pointer: &s.ApplicationName},
			Field{info: parse.FieldInfo{Name: "CmdType", Type: "int32", Column: "cmd_type"}, pointer: &s.CmdType},
			Field{info: parse.FieldInfo{Name: "CmdTypeText", Type: "string", Column: "cmd_type_text"}, pointer: &s.CmdTypeText},
			Field{info: parse.FieldInfo{Name: "Elevel", Type: "int32", Column: "elevel"}, pointer: &s.Elevel},
			Field{info: parse.FieldInfo{Name: "Sqlcode", Type: "*string", Column: "sqlcode"}, pointer: &s.Sqlcode},
			Field{info: parse.FieldInfo{Name: "Message", Type: "*string", Column: "message"}, pointer: &s.Message},
			Field{info: parse.FieldInfo{Name: "WalRecords", Type: "int64", Column: "wal_records"}, pointer: &s.WalRecords},
			Field{info: parse.FieldInfo{Name: "WalFpi", Type: "int64", Column: "wal_fpi"}, pointer: &s.WalFpi},
			Field{info: parse.FieldInfo{Name: "WalBytes", Type: "int64", Column: "wal_bytes"}, pointer: &s.WalBytes},
			Field{info: parse.FieldInfo{Name: "StateCode", Type: "int64", Column: "state_code"}, pointer: &s.StateCode},
			Field{info: parse.FieldInfo{Name: "State", Type: "string", Column: "state"}, pointer: &s.State},
		)
	}

	if v <= pgStatMonitorVersion10PG12 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "TotalTime", Type: "float64", Column: "total_time"}, pointer: &s.TotalTime},
			Field{info: parse.FieldInfo{Name: "MinTime", Type: "float64", Column: "min_time"}, pointer: &s.MinTime},
			Field{info: parse.FieldInfo{Name: "MaxTime", Type: "float64", Column: "max_time"}, pointer: &s.MaxTime},
			Field{info: parse.FieldInfo{Name: "MeanTime", Type: "float64", Column: "mean_time"}, pointer: &s.MeanTime},
			Field{info: parse.FieldInfo{Name: "StddevTime", Type: "float64", Column: "stddev_time"}, pointer: &s.StddevTime},
		)
	}
	if v >= pgStatMonitorVersion10PG13 {
		fields = append(fields,
			Field{info: parse.FieldInfo{Name: "TotalTime", Type: "float64", Column: "total_exec_time"}, pointer: &s.TotalTime},
			Field{info: parse.FieldInfo{Name: "MinTime", Type: "float64", Column: "min_exec_time"}, pointer: &s.MinTime},
			Field{info: parse.FieldInfo{Name: "MaxTime", Type: "float64", Column: "max_exec_time"}, pointer: &s.MaxTime},
			Field{info: parse.FieldInfo{Name: "MeanTime", Type: "float64", Column: "mean_exec_time"}, pointer: &s.MeanTime},
			Field{info: parse.FieldInfo{Name: "StddevTime", Type: "float64", Column: "stddev_exec_time"}, pointer: &s.StddevTime},
			Field{info: parse.FieldInfo{Name: "PlansCalls", Type: "int64", Column: "plans_calls"}, pointer: &s.PlansCalls},
			Field{info: parse.FieldInfo{Name: "PlanTotalTime", Type: "float64", Column: "total_plan_time"}, pointer: &s.PlanTotalTime},
			Field{info: parse.FieldInfo{Name: "PlanMinTime", Type: "float64", Column: "min_plan_time"}, pointer: &s.PlanMinTime},
			Field{info: parse.FieldInfo{Name: "PlanMaxTime", Type: "float64", Column: "max_plan_time"}, pointer: &s.PlanMaxTime},
			Field{info: parse.FieldInfo{Name: "PlanMeanTime", Type: "float64", Column: "mean_plan_time"}, pointer: &s.PlanMeanTime},
		)
	}
	//if v >= pgStatMonitorVersion10PG14 {
	//	fields = append(fields,
	//		Field{info: parse.FieldInfo{Name: "PlanTotalTime", Type: "float64", Column: "total_plan_time"}, pointer: &s.PlanTotalTime},
	//		Field{info: parse.FieldInfo{Name: "PlanMinTime", Type: "float64", Column: "min_plan_time"}, pointer: &s.PlanMinTime},
	//		Field{info: parse.FieldInfo{Name: "PlanMaxTime", Type: "float64", Column: "max_plan_time"}, pointer: &s.PlanMaxTime},
	//		Field{info: parse.FieldInfo{Name: "PlanMeanTime", Type: "float64", Column: "mean_plan_time"}, pointer: &s.PlanMeanTime},
	//	)
	//}

	s.pointers = make([]interface{}, len(fields))
	var pgStatMonitorDefaultView = &pgStatMonitorAllViewType{
		s: parse.StructInfo{
			Type:         "pgStatMonitor",
			SQLName:      "pg_stat_monitor",
			Fields:       make([]parse.FieldInfo, len(fields)),
			PKFieldIndex: -1,
		},
		c: make([]string, len(fields)),
		v: v,
	}
	for i, field := range fields {
		pgStatMonitorDefaultView.s.Fields[i] = field.info
		pgStatMonitorDefaultView.c[i] = field.info.Column
		s.pointers[i] = field.pointer
	}
	s.view = pgStatMonitorDefaultView
	pgStatMonitorDefaultView.z = s.Values()
	return s, pgStatMonitorDefaultView
}

//
type pgStatMonitorAllViewType struct {
	s parse.StructInfo
	z []interface{}
	c []string
	v pgStatMonitorVersion
}

// Schema returns a schema name in SQL database ("").
func (v *pgStatMonitorAllViewType) Schema() string {
	return v.s.SQLSchema
}

// Name returns a view or table name in SQL database ("pg_stat_monitor").
func (v *pgStatMonitorAllViewType) Name() string {
	return v.s.SQLName
}

// Columns returns a new slice of column names for that view or table in SQL database.
func (v *pgStatMonitorAllViewType) Columns() []string {
	return v.c
}

// NewStruct makes a new struct for that view or table.
func (v *pgStatMonitorAllViewType) NewStruct() reform.Struct {
	str, _ := NewPgStatMonitorStructs(v.v)
	return str
}

// String returns a string representation of this struct or record.
func (s pgStatMonitor) String() string {
	res := make([]string, 51)
	res[0] = "Bucket: " + reform.Inspect(s.Bucket, true)
	res[1] = "BucketStartTime: " + reform.Inspect(s.BucketStartTime, true)
	res[2] = "UserID: " + reform.Inspect(s.UserID, true)
	res[3] = "ClientIP: " + reform.Inspect(s.ClientIP, true)
	res[4] = "QueryID: " + reform.Inspect(s.QueryID, true)
	res[5] = "Query: " + reform.Inspect(s.Query, true)
	res[6] = "Relations: " + reform.Inspect(s.Relations, true)
	res[7] = "Calls: " + reform.Inspect(s.Calls, true)
	res[8] = "TotalTime: " + reform.Inspect(s.TotalTime, true)
	res[9] = "SharedBlksHit: " + reform.Inspect(s.SharedBlksHit, true)
	res[10] = "SharedBlksRead: " + reform.Inspect(s.SharedBlksRead, true)
	res[11] = "SharedBlksDirtied: " + reform.Inspect(s.SharedBlksDirtied, true)
	res[12] = "SharedBlksWritten: " + reform.Inspect(s.SharedBlksWritten, true)
	res[13] = "LocalBlksHit: " + reform.Inspect(s.LocalBlksHit, true)
	res[14] = "LocalBlksRead: " + reform.Inspect(s.LocalBlksRead, true)
	res[15] = "LocalBlksDirtied: " + reform.Inspect(s.LocalBlksDirtied, true)
	res[16] = "LocalBlksWritten: " + reform.Inspect(s.LocalBlksWritten, true)
	res[17] = "TempBlksRead: " + reform.Inspect(s.TempBlksRead, true)
	res[18] = "TempBlksWritten: " + reform.Inspect(s.TempBlksWritten, true)
	res[19] = "BlkReadTime: " + reform.Inspect(s.BlkReadTime, true)
	res[20] = "BlkWriteTime: " + reform.Inspect(s.BlkWriteTime, true)
	res[21] = "RespCalls: " + reform.Inspect(s.RespCalls, true)
	res[22] = "CPUUserTime: " + reform.Inspect(s.CPUUserTime, true)
	res[23] = "CPUSysTime: " + reform.Inspect(s.CPUSysTime, true)
	res[24] = "DBID: " + reform.Inspect(s.DBID, true)
	res[25] = "DatName: " + reform.Inspect(s.DatName, true)
	res[26] = "Rows: " + reform.Inspect(s.Rows, true)
	res[27] = "TopQueryID: " + reform.Inspect(s.TopQueryID, true)
	res[28] = "PlanID: " + reform.Inspect(s.PlanID, true)
	res[29] = "QueryPlan: " + reform.Inspect(s.QueryPlan, true)
	res[30] = "TopQuery: " + reform.Inspect(s.TopQuery, true)
	res[31] = "ApplicationName: " + reform.Inspect(s.ApplicationName, true)
	res[32] = "CmdType: " + reform.Inspect(s.CmdType, true)
	res[33] = "CmdTypeText: " + reform.Inspect(s.CmdTypeText, true)
	res[34] = "Elevel: " + reform.Inspect(s.Elevel, true)
	res[35] = "Sqlcode: " + reform.Inspect(s.Sqlcode, true)
	res[36] = "Message: " + reform.Inspect(s.Message, true)
	res[37] = "MinTime: " + reform.Inspect(s.MinTime, true)
	res[38] = "MaxTime: " + reform.Inspect(s.MaxTime, true)
	res[39] = "MeanTime: " + reform.Inspect(s.MeanTime, true)
	res[40] = "StddevTime: " + reform.Inspect(s.StddevTime, true)
	res[41] = "PlansCalls: " + reform.Inspect(s.PlansCalls, true)
	res[42] = "PlanTotalTime: " + reform.Inspect(s.PlanTotalTime, true)
	res[43] = "PlanMinTime: " + reform.Inspect(s.PlanMinTime, true)
	res[44] = "PlanMaxTime: " + reform.Inspect(s.PlanMaxTime, true)
	res[45] = "PlanMeanTime: " + reform.Inspect(s.PlanMeanTime, true)
	res[46] = "WalRecords: " + reform.Inspect(s.WalRecords, true)
	res[47] = "WalFpi: " + reform.Inspect(s.WalFpi, true)
	res[48] = "WalBytes: " + reform.Inspect(s.WalBytes, true)
	res[49] = "StateCode: " + reform.Inspect(s.StateCode, true)
	res[50] = "State: " + reform.Inspect(s.State, true)
	return strings.Join(res, ", ")
}

// Values returns a slice of struct or record field values.
// Returned interface{} values are never untyped nils.
func (s *pgStatMonitor) Values() []interface{} {
	values := make([]interface{}, len(s.pointers))
	for i, pointer := range s.pointers {
		values[i] = reflect.ValueOf(pointer).Interface()
	}
	return values
}

// Pointers returns a slice of pointers to struct or record fields.
// Returned interface{} values are never untyped nils.
func (s *pgStatMonitor) Pointers() []interface{} {
	return s.pointers
}

// View returns View object for that struct.
func (s *pgStatMonitor) View() reform.View {
	return s.view
}

// check interfaces
var (
	_ reform.Struct = (*pgStatMonitor)(nil)
	_ fmt.Stringer  = (*pgStatMonitor)(nil)
)
