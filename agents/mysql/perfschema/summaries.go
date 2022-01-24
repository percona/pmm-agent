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

package perfschema

import (
	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
	"strings"
)

func getSummaries(q *reform.Querier) (summaryMap, error) {
	rows, err := q.SelectRows(eventsStatementsSummaryByDigestView, "WHERE DIGEST IS NOT NULL AND DIGEST_TEXT IS NOT NULL")
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events_statements_summary_by_digest")
	}
	defer rows.Close() //nolint:errcheck

	res := make(summaryMap)
	for {
		var ess eventsStatementsSummaryByDigest
		if err = q.NextRow(&ess, rows); err != nil {
			break
		}

		// From https://dev.mysql.com/doc/relnotes/mysql/8.0/en/news-8-0-11.html:
		// > The Performance Schema could produce DIGEST_TEXT values with a trailing space. […] (Bug #26908015)
		*ess.DigestText = strings.TrimSpace(*ess.DigestText)

		res[*ess.Digest] = &ess
	}
	if err != reform.ErrNoRows {
		return nil, errors.Wrap(err, "failed to fetch events_statements_summary_by_digest")
	}
	return res, nil
}
