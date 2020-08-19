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

	"github.com/percona/pmm/api/agentpb"
)

type ptSummaryAction struct {
	id     string
	params *agentpb.StartActionRequest_PtSummaryParams
}

// NewPTSummaryAction creates a MongoDB adminCommand query Action.
func NewPTSummaryAction(id string, params *agentpb.StartActionRequest_PtSummaryParams) Action {
	return &ptSummaryAction{
		id:     id,
		params: params,
	}
}

// ID returns an Action ID.
func (a *ptSummaryAction) ID() string {
	return a.id
}

// Type returns an Action type.
func (a *ptSummaryAction) Type() string {
	return "pt-summary"
}

// Run runs an Action and returns output and error.
func (a *ptSummaryAction) Run(ctx context.Context) ([]byte, error) {
	test := make(map[string]interface{})
	test["test"] = "test string"

	data := []map[string]interface{}{}
	data = append(data, test)
	return agentpb.MarshalActionQueryDocsResult(data)
}

func (a *ptSummaryAction) sealed() {}
