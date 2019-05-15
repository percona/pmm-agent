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
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConcurrentRunnerRun(t *testing.T) {
	cr := NewConcurrentRunner(logrus.WithField("component", "runner"), 0)

	a1 := NewShellAction("/action_id/6a479303-5081-46d0-baa0-87d6248c987b", "echo", []string{"test"})
	a2 := NewShellAction("/action_id/84140ab2-612d-4d93-9360-162a4bd5de14", "echo", []string{"test2"})

	cr.Run(a1)
	cr.Run(a2)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(t *testing.T, ready <-chan ActionResult) {
		defer wg.Done()
		expected := []string{"test\n", "test2\n"}
		for i := 0; i < 2; i++ {
			a := <-ready
			assert.Contains(t, expected, string(a.CombinedOutput))
		}
	}(t, cr.ActionReady())
	wg.Wait()
}

func TestConcurrentRunnerTimeout(t *testing.T) {
	cr := NewConcurrentRunner(logrus.WithField("component", "runner"), time.Second)
	a1 := NewShellAction("/action_id/6a479303-5081-46d0-baa0-87d6248c987b", "sleep", []string{"20"})
	a2 := NewShellAction("/action_id/84140ab2-612d-4d93-9360-162a4bd5de14", "sleep", []string{"30"})

	cr.Run(a1)
	cr.Run(a2)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(t *testing.T, ready <-chan ActionResult) {
		defer wg.Done()

		// check action returns proper errors and output.
		expected := []string{"signal: killed", "signal: killed"}
		expectedOut := []string{"", ""}
		for i := 0; i < 2; i++ {
			a := <-ready
			assert.Contains(t, expected, a.Error.Error())
			assert.Contains(t, expectedOut, string(a.CombinedOutput))
		}

		// check action was deleted from actions map.
		_, ok := cr.runningActions.Load(a1.ID())
		_, ok2 := cr.runningActions.Load(a2.ID())
		assert.False(t, ok)
		assert.False(t, ok2)

	}(t, cr.ActionReady())
	wg.Wait()
}
