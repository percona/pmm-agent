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

package config

const (
	DefaultInterval        uint  = 60         // 1 minute
	DefaultMaxSlowLogSize  int64 = 1073741824 // 1G
	DefaultSlowLogRotation       = true       // whether to rotate slow logs
	DefaultRetainSlowLogs        = 1          // how many slow logs to keep on filesystem
	DefaultExampleQueries        = true
	// internal
	DefaultReportLimit uint = 200
)

type QAN struct {
	UUID           string // of MySQL instance
	CollectFrom    string `json:",omitempty"` // "slowlog" or "perfschema"
	Interval       uint   `json:",omitempty"` // seconds, 0 = DEFAULT_INTERVAL
	ExampleQueries *bool  `json:",omitempty"` // send real example of each query
	// "slowlog" specific options.
	MaxSlowLogSize  int64 `json:"-"`          // bytes, 0 = DEFAULT_MAX_SLOW_LOG_SIZE. Don't write it to the config
	SlowLogRotation *bool `json:",omitempty"` // Enable slow logs rotation.
	RetainSlowLogs  *int  `json:",omitempty"` // Number of slow logs to keep.
	// internal
	Start       []string `json:",omitempty"` // queries to configure MySQL (enable slow log, etc.)
	Stop        []string `json:",omitempty"` // queries to un-configure MySQL (disable slow log, etc.)
	ReportLimit uint     `json:",omitempty"` // top N queries, 0 = DEFAULT_REPORT_LIMIT
}

func NewQAN() QAN {
	return QAN{
		Interval:       DefaultInterval,
		ExampleQueries: boolPointer(DefaultExampleQueries),
		// "slowlog" specific options.
		MaxSlowLogSize:  DefaultMaxSlowLogSize,
		SlowLogRotation: boolPointer(DefaultSlowLogRotation),
		RetainSlowLogs:  intPointer(DefaultRetainSlowLogs),
		// internal
		ReportLimit: DefaultReportLimit,
	}
}

// boolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func boolPointer(v bool) *bool {
	return &v
}

// boolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func intPointer(v int) *int {
	return &v
}
