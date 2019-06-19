package pgstatsstatements

//go:generate reform

// pgStatDatabase represents a row in pg_stat_database view.
//reform:pg_catalog.pg_stat_database
type pgStatDatabase struct {
	Datid   []byte  `reform:"datid"` // FIXME unhandled database type "oid"
	Datname *string `reform:"datname"`
}

// pgUser represents a row in pg_user view.
//reform:pg_catalog.pg_user
type pgUser struct {
	UserId   []byte  `reform:"usesysid"` // FIXME unhandled database type "oid"
	Username *string `reform:"usename"`
}

// pgStatStatements represents a row in pg_stat_statements view.
//reform:pg_stat_statements
type pgStatStatements struct {
	Userid    []byte   `reform:"userid"` // FIXME unhandled database type "oid"
	Dbid      []byte   `reform:"dbid"`   // FIXME unhandled database type "oid"
	Queryid   *int64   `reform:"queryid"`
	Query     *string  `reform:"query"`
	Calls     *int64   `reform:"calls"`
	TotalTime *float64 `reform:"total_time"`
	//MinTime           *float64 `reform:"min_time"`
	//MaxTime           *float64 `reform:"max_time"`
	//MeanTime          *float64 `reform:"mean_time"`
	//StddevTime        *float64 `reform:"stddev_time"`
	Rows              *int64   `reform:"rows"`
	SharedBlksHit     *int64   `reform:"shared_blks_hit"`
	SharedBlksRead    *int64   `reform:"shared_blks_read"`
	SharedBlksDirtied *int64   `reform:"shared_blks_dirtied"`
	SharedBlksWritten *int64   `reform:"shared_blks_written"`
	LocalBlksHit      *int64   `reform:"local_blks_hit"`
	LocalBlksRead     *int64   `reform:"local_blks_read"`
	LocalBlksDirtied  *int64   `reform:"local_blks_dirtied"`
	LocalBlksWritten  *int64   `reform:"local_blks_written"`
	TempBlksRead      *int64   `reform:"temp_blks_read"`
	TempBlksWritten   *int64   `reform:"temp_blks_written"`
	BlkReadTime       *float64 `reform:"blk_read_time"`
	BlkWriteTime      *float64 `reform:"blk_write_time"`
}
