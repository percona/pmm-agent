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

package actions

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunShellAction(t *testing.T) {
	// setup
	p := NewShellAction("/action_id/6a479303-5081-46d0-baa0-87d6248c987b", "pt-summary", nil)
	_, err := exec.LookPath("pt-summary")
	if err != nil {
		t.Skipf("Test skipped, reason: %s", err)
	}

	// run
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	got, err := p.Run(ctx)

	// check
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	t.Logf("'%d' bytes read", len(got))
}

func TestRunForbiddenShellAction(t *testing.T) {
	// setup
	p := NewShellAction("/action_id/84140ab2-612d-4d93-9360-162a4bd5de14", "rm", nil)
	_, err := exec.LookPath("rm")
	if err != nil {
		t.Skipf("Test skipped, reason: %s", err)
	}

	// run
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	_, err = p.Run(ctx)

	// check
	require.Equal(t, err, errUnknownAction)
}
