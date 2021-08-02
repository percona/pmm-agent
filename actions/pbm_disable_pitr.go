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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"

	"github.com/percona/pmm-agent/utils/templates"
)

const pbmBin = "pbm"

type pbmSwitchPITRAction struct {
	id      string
	params  *agentpb.StartActionRequest_PBMSwitchPITRParams
	tempDir string
}

// NewPBMSwitchPITRAction creates a PBM swithc PITR Action.
func NewPBMSwitchPITRAction(id string, params *agentpb.StartActionRequest_PBMSwitchPITRParams, tempDir string) Action {
	return &pbmSwitchPITRAction{
		id:      id,
		params:  params,
		tempDir: tempDir,
	}
}

func (a pbmSwitchPITRAction) ID() string {
	return a.id
}

func (a pbmSwitchPITRAction) Type() string {
	return "pbm-switch-pitr"
}

func (a pbmSwitchPITRAction) Run(ctx context.Context) ([]byte, error) {
	if _, err := exec.LookPath(pbmBin); err != nil {
		return nil, errors.Wrapf(err, "lookpath: %s", pbmBin)
	}

	dsn, err := templates.RenderDSN(a.params.Dsn, a.params.TextFiles, filepath.Join(a.tempDir, strings.ToLower(a.Type()), a.id))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	output, err := exec.CommandContext(
		ctx,
		pbmBin,
		"config",
		"--set pitr.enabled="+strconv.FormatBool(a.params.Enabled),
		"--mongodb-uri="+dsn).
		CombinedOutput() // #nosec G204
	if err != nil {
		return nil, errors.Wrapf(err, "pbm config error: %s", string(output))
	}

	return output, nil
}

func (a pbmSwitchPITRAction) sealed() {}
