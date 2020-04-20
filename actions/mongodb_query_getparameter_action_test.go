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

package actions

import (
	"context"
	"testing"
	"time"

	"github.com/percona/pmm/api/agentpb"
	"github.com/stretchr/objx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/utils/tests"
)

func TestMongoDBGetparameter(t *testing.T) {
	t.Parallel()

	client := tests.OpenTestMongoDB(t)
	defer client.Disconnect(context.Background()) //nolint:errcheck

	t.Run("Default", func(t *testing.T) {
		params := &agentpb.StartActionRequest_MongoDBQueryGetParameterParams{
			Dsn: tests.GetTestMongoDBDSN(t),
		}
		a := NewMongoDBQueryGetparameterAction("", params)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		b, err := a.Run(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 10518, len(b), 1)

		data, err := agentpb.UnmarshalActionQueryResult(b)
		require.NoError(t, err)
		assert.Len(t, data, 1)
		m := objx.Map(data[0])
		assert.Equal(t, 1.0, m.Get("ok").Data())
		assert.Less(t, int64(1024), m.Get("transactionSizeLimitBytes").Int64())
		assert.Equal(t, []interface{}{"MONGODB-X509", "SCRAM-SHA-1", "SCRAM-SHA-256"}, m.Get("authenticationMechanisms").Data())
		assert.Equal(t, "4.2", m.Get("featureCompatibilityVersion.version").String())
	})
}