// pmm-managed
// Copyright (C) 2022 Percona LLC
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
package main

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type RestartPolicy int

const (
	DoNotRestart RestartPolicy = iota + 1
	RestartAlways
	RestartOnFail
)

var helpText = `
PMM 2.x Client Docker container.

It runs pmm-agent as a process with PID 1.
It is configured entirely by environment variables. Arguments or flags are not used.

The following environment variables are recognized by the Docker entrypoint:
* PMM_AGENT_SETUP            - if true, 'pmm-agent setup' is called before 'pmm-agent run'.
* PMM_AGENT_PRERUN_FILE      - if non-empty, runs given file with 'pmm-agent run' running in the background.
* PMM_AGENT_PRERUN_SCRIPT    - if non-empty, runs given shell script content with 'pmm-agent run' running in the background.
* PMM_AGENT_SIDECAR          - if true, 'pmm-agent' will be restarted in case of it's failed.
* PMM_AGENT_SIDECAR_SLEEP    - time to wait before restarting pmm-agent if PMM_AGENT_SIDECAR is true. 1 second by default.

Additionally, the many environment variables are recognized by pmm-agent itself.
The following help text shows them as [PMM_AGENT_XXX].
`

var (
	pmmAgentSetup        = strconv.ParseBool(getEnvWithDefault("PMM_AGENT_SETUP", "false"))
	pmmAgentSidecar      = strconv.ParseBool(getEnvWithDefault("PMM_AGENT_SIDECAR", "false"))
	pmmAgentSidecarSleep = strconv.Atoi(getEnvWithDefault("PMM_AGENT_SIDECAR_SLEEP", "1"))
	pmmAgentPrerunFile   = getEnvWithDefault("PMM_AGENT_PRERUN_FILE", "")
	pmmAgentPrerunScript = getEnvWithDefault("PMM_AGENT_PRERUN_SCRIPT", "")
)

func getEnvWithDefault(key, defautlValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defautlValue
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000-07:00",
	})
}

func main() {

	if len(os.Args) > 1 {
		log.Info(helpText)
		exec.Command("pmm-agent", "setup", "--help")
		os.Exit(1)
		// print(__doc__, file=sys.stderr)
		// subprocess.call(['pmm-agent', 'setup', '--help'])
		// sys.exit(1)
	}
	if pmmAgentPrerunFile && pmmAgentPrerunScript

	log.Info(pmmAgentSetup, pmmAgentPrerunFile, pmmAgentPrerunScript, pmmAgentSidecar, pmmAgentSidecarSleep)

}
