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
	"bytes"
	"context"
	"os/exec"

	"github.com/percona/pmm-agent/config"
	"github.com/sirupsen/logrus"
)

type ptSummaryAction struct {
	id string
}

// NewPTSummaryAction creates a PT summary Action.
func NewPTSummaryAction(id string) Action {
	return &ptSummaryAction{
		id: id,
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
	l := logrus.WithField("component", "pt-summary")
	cfg, _, err := config.Get(l)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	cmd := exec.CommandContext(ctx, "pt-summary")
	cmd.Dir = cfg.Paths.PTSummary
	cmd.Stdout = buf
	cmd.Stderr = new(bytes.Buffer)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *ptSummaryAction) sealed() {}
