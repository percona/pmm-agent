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

package pgstatstatements

import (
	"fmt"

	"github.com/AlekSi/pointer"
)

//go:generate reform

func trimQuery(query string) string {
	if len(query) <= 50 {
		return query
	}
	return fmt.Sprintf("%s[... %d chars ...]%s", query[:25], len(query)-50, query[len(query)-25:])
}

// pgStatDatabase represents a row in pg_stat_database view.
//reform:pg_catalog.pg_stat_database
type pgStatDatabase struct {
	DatID   int64   `reform:"datid"`
	DatName *string `reform:"datname"`
}

// pgUser represents a row in pg_user view.
//reform:pg_catalog.pg_user
type pgUser struct {
	UserID   int64   `reform:"usesysid"`
	UserName *string `reform:"usename"`
}

// pgStatStatements represents a row in pg_stat_statements view.
//reform:pg_stat_statements
type pgStatStatements struct {
	UserID    int64   `reform:"userid"`
	DBID      int64   `reform:"dbid"`
	QueryID   *int64  `reform:"queryid"`
	Query     *string `reform:"query"`
	Calls     int64   `reform:"calls"`
	TotalTime float64 `reform:"total_time"`
	//MinTime           *float64 `reform:"min_time"`
	//MaxTime           *float64 `reform:"max_time"`
	//MeanTime          *float64 `reform:"mean_time"`
	//StddevTime        *float64 `reform:"stddev_time"`
	Rows              int64   `reform:"rows"`
	SharedBlksHit     int64   `reform:"shared_blks_hit"`
	SharedBlksRead    int64   `reform:"shared_blks_read"`
	SharedBlksDirtied int64   `reform:"shared_blks_dirtied"`
	SharedBlksWritten int64   `reform:"shared_blks_written"`
	LocalBlksHit      int64   `reform:"local_blks_hit"`
	LocalBlksRead     int64   `reform:"local_blks_read"`
	LocalBlksDirtied  int64   `reform:"local_blks_dirtied"`
	LocalBlksWritten  int64   `reform:"local_blks_written"`
	TempBlksRead      int64   `reform:"temp_blks_read"`
	TempBlksWritten   int64   `reform:"temp_blks_written"`
	BlkReadTime       float64 `reform:"blk_read_time"`
	BlkWriteTime      float64 `reform:"blk_write_time"`
}

// pgStatStatementsExtended contains pgStatStatements data and extends it with database, username and tables data.
// It's made for performance reason.
type pgStatStatementsExtended struct {
	pgStatStatements

	// In those fields, nil means "not know yet", non-nil value (even if empty) means extraction was performed.
	Database *string
	Username *string
	Tables   []string
}

func (e *pgStatStatementsExtended) String() string {
	return fmt.Sprintf("%q %q %v: %d %s",
		pointer.GetString(e.Database), pointer.GetString(e.Username), e.Tables,
		pointer.GetInt64(e.pgStatStatements.QueryID), trimQuery(pointer.GetString(e.pgStatStatements.Query)),
	)
}
