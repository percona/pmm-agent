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

	"github.com/stretchr/testify/require"

	"github.com/percona/pmm/api/agentpb"
)

func TestPTSummaryAction(t *testing.T) {
	params := &agentpb.StartActionRequest_PTSummaryParams{
		PmmAgentId: "pmm",
		NodeId:     "node",
	}
	a := NewPTSummaryAction("", params)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := a.Run(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, res)
}
