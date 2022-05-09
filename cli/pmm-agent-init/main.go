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
	"os/exec"
	"strconv"
	"syscall"
	"time"

	reaper "github.com/ramr/go-reaper"
	log "github.com/sirupsen/logrus"
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

type RestartPolicy int

const (
	DoNotRestart RestartPolicy = iota + 1
	RestartAlways
	RestartOnFail
)

var (
	pmmAgentSetupEnv        = getEnvWithDefault("PMM_AGENT_SETUP", "false")
	pmmAgentSidecarEnv      = getEnvWithDefault("PMM_AGENT_SIDECAR", "false")
	pmmAgentSidecarSleepEnv = getEnvWithDefault("PMM_AGENT_SIDECAR_SLEEP", "1")
	pmmAgentPrerunFile      = getEnvWithDefault("PMM_AGENT_PRERUN_FILE", "")
	pmmAgentPrerunScript    = getEnvWithDefault("PMM_AGENT_PRERUN_SCRIPT", "")
)

type processRunner struct {
	command              *exec.Cmd
	pmmAgentSidecarSleep int
}

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

func (process processRunner) run(commandLine string, restartPolicy RestartPolicy) int {
	for {
		log.Infof("Starting '%s' ...", commandLine)
		cmd := exec.Command(commandLine)
		if err := cmd.Run(); err != nil {
			log.Errorf("Can't run: '%s', Error: %s", commandLine, err)
			return -1
		}
		process.command = cmd
		if err := cmd.Wait(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if restartPolicy == RestartAlways || (restartPolicy == RestartOnFail && exitError.ExitCode() != 0) {
					log.Infof("Restarting %s in %s seconds because PMM_AGENT_SIDECAR is enabled ...", commandLine, process.pmmAgentSidecarSleep)
					time.Sleep(time.Duration(process.pmmAgentSidecarSleep) * time.Second)
				} else {
					return exitError.ExitCode()
				}
			} else {
				log.Errorf("Can't get exit code for %s.", commandLine)
				return -1
			}
		}
	}
}

func main() {
	go reaper.Reap()

	var status int

	if len(os.Args) > 1 {
		log.Info(helpText)
		exec.Command("pmm-agent", "setup", "--help")
		os.Exit(1)
	}

	pmmAgentSetup, err := strconv.ParseBool(pmmAgentSetupEnv)
	if err != nil {
		log.Fatalf("Can't parse %s as boolean variable", pmmAgentSetupEnv)
	}

	pmmAgentSidecar, err := strconv.ParseBool(pmmAgentSidecarEnv)
	if err != nil {
		log.Fatalf("Can't parse %s as boolean variable", pmmAgentSidecarEnv)
	}

	pmmAgentSidecarSleep, err := strconv.Atoi(pmmAgentSidecarSleepEnv)
	if err != nil {
		log.Fatalf("Can't parse %s as int variable", pmmAgentSidecarSleepEnv)
	}

	if pmmAgentPrerunFile != "" && pmmAgentPrerunScript != "" {
		log.Error("Both PMM_AGENT_PRERUN_FILE and PMM_AGENT_PRERUN_SCRIPT cannot be set.")
		os.Exit(1)
	}

	runner := processRunner{
		pmmAgentSidecarSleep: pmmAgentSidecarSleep,
	}

	if pmmAgentSetup {
		var agent *exec.Cmd
		restartPolicy := DoNotRestart
		if pmmAgentSidecar {
			restartPolicy = RestartOnFail
			log.Info("Starting pmm-agent for liveness probe...")
			agent = exec.Command("pmm-agent run")
			err := agent.Start()
			if err != nil {
				log.Fatalf("Can't run pmm-agent: %s", err)
			}
		}
		status := runner.run("pmm-agent setup", restartPolicy)
		if status != 0 {
			os.Exit(status)
		}
		if pmmAgentSidecar {
			log.Info("Stopping pmm-agent...")
			if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
				log.Fatal("Failed to kill pmm-agent: ", err)
			}
		}
	}

	if pmmAgentPrerunFile != "" || pmmAgentPrerunScript != "" {
		log.Info("Starting pmm-agent for prerun ...")
		agent := exec.Command("pmm-agent run")
		err := agent.Start()
		if err != nil {

		}

		if pmmAgentPrerunFile != "" {
			log.Info("Running prerun file %s...", pmmAgentPrerunFile)
			cmd := exec.Command(pmmAgentPrerunFile)
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					log.Info("Prerun file exited with %s", exitError.ExitCode())
				}
			}
		}

		if pmmAgentPrerunScript != "" {
			log.Info("Running prerun shell script %s...", pmmAgentPrerunScript)
			cmd := exec.Command("/bin/sh " + pmmAgentPrerunScript)
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					log.Info("Prerun shell script exited with %s", exitError.ExitCode())
				}
			}
		}

		log.Info("Stopping pmm-agent...")
		if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
			log.Fatalf("Failed to kill pmm-agent: %s", err)
		}
		for i := 0; i < 10; i++ {
			if agent.ProcessState.Exited() {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if !agent.ProcessState.Exited() {
			log.Info("Killing pmm-agent...")
			agent.Process.Kill()
		}
		agent.Wait()

		if status != 0 && !pmmAgentSidecar {
			os.Exit(status)
		}
	}
	restartPolicy := DoNotRestart
	if pmmAgentSidecar {
		restartPolicy = RestartAlways
	}
	runner.run("pmm-agent run", restartPolicy)
}
