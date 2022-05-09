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
	"context"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	reaper "github.com/ramr/go-reaper"
	"github.com/sirupsen/logrus"
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
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000-07:00",
	})
}

func (process processRunner) run(commandLineArgs []string, restartPolicy RestartPolicy, l *logrus.Entry) int {
	pmmAgentFullCommand := "pmm-admin " + strings.Join(commandLineArgs, " ")
	for {
		l.Infof("Starting 'pmm-admin %s'...", strings.Join(commandLineArgs, " "))
		cmd := runPmmAgent(commandLineArgs)
		if err := cmd.Start(); err != nil {
			l.Errorf("Can't run: '%s', Error: %s", commandLineArgs, err)
			return -1
		}
		process.command = cmd
		var exitCode int
		if err := cmd.Wait(); err != nil {
			exitError, ok := err.(*exec.ExitError)
			if !ok {
				l.Errorf("Can't get exit code for '%d'. Error code: %s", pmmAgentFullCommand, err)
				return -1
			}
			exitCode = exitError.ExitCode()
		}
		l.Infof("'%s' exited with %d", pmmAgentFullCommand, exitCode)

		if restartPolicy == RestartAlways || (restartPolicy == RestartOnFail && exitCode != 0) {
			l.Infof("Restarting `%s` in %d seconds because PMM_AGENT_SIDECAR is enabled...", pmmAgentFullCommand, process.pmmAgentSidecarSleep)
			time.Sleep(time.Duration(process.pmmAgentSidecarSleep) * time.Second)
		} else {
			return exitCode
		}

	}
}

func runPmmAgent(args []string) *exec.Cmd {
	const pmmAgentCommandName = "pmm-agent"
	command := exec.Command(pmmAgentCommandName, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command
}

func main() {
	go reaper.Reap()

	var status int

	l := logrus.WithField("component", "entrypoint")

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-signals
		signal.Stop(signals)
		l.Warnf("Got %s, shutting down...", unix.SignalName(s.(unix.Signal)))
		cancel()
	}()

	if len(os.Args) > 1 {
		l.Info(helpText)
		exec.CommandContext(ctx, "pmm-agent", "setup", "--help")
		os.Exit(1)
	}

	pmmAgentSetup, err := strconv.ParseBool(pmmAgentSetupEnv)
	if err != nil {
		l.Fatalf("Can't parse %s as boolean variable", pmmAgentSetupEnv)
	}
	l.Infof("Run setup: %t", pmmAgentSetup)

	pmmAgentSidecar, err := strconv.ParseBool(pmmAgentSidecarEnv)
	if err != nil {
		l.Fatalf("Can't parse %s as boolean variable", pmmAgentSidecarEnv)
	}
	l.Infof("Sidecar mode: %t", pmmAgentSidecar)

	pmmAgentSidecarSleep, err := strconv.Atoi(pmmAgentSidecarSleepEnv)
	if err != nil {
		l.Fatalf("Can't parse %s as int variable", pmmAgentSidecarSleepEnv)
	}

	if pmmAgentPrerunFile != "" && pmmAgentPrerunScript != "" {
		l.Error("Both PMM_AGENT_PRERUN_FILE and PMM_AGENT_PRERUN_SCRIPT cannot be set.")
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
			l.Info("Starting pmm-agent for liveness probe...")
			agent = runPmmAgent(ctx, []string{"run"})
			agent.Stdout = os.Stdout
			agent.Stderr = os.Stderr
			err := agent.Start()
			if err != nil {
				l.Fatalf("Can't run pmm-agent: %s", err)
			}
		}
		status := runner.run([]string{"setup"}, restartPolicy, l)
		if status != 0 {
			os.Exit(status)
		}
		if pmmAgentSidecar {
			l.Info("Stopping pmm-agent...")
			if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
				l.Fatal("Failed to kill pmm-agent: ", err)
			}
		}
	}

	if pmmAgentPrerunFile != "" || pmmAgentPrerunScript != "" {
		l.Info("Starting pmm-agent for prerun...")
		agent := runPmmAgent(ctx, []string{"run"})
		err := agent.Start()
		if err != nil {
			l.Errorf("Failed to run pmm-agent run command: %s")
		}

		if pmmAgentPrerunFile != "" {
			l.Infof("Running prerun file %s...", pmmAgentPrerunFile)
			cmd := exec.CommandContext(ctx, pmmAgentPrerunFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					l.Infof("Prerun file exited with %d", exitError.ExitCode())
				}
			}
		}

		if pmmAgentPrerunScript != "" {
			l.Infof("Running prerun shell script %s...", pmmAgentPrerunScript)
			cmd := exec.CommandContext(ctx, "/bin/sh", pmmAgentPrerunScript)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					l.Infof("Prerun shell script exited with %d", exitError.ExitCode())
				}
			}
		}

		l.Info("Stopping pmm-agent...")
		if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
			l.Fatalf("Failed to kill pmm-agent: %s", err)
		}
		for i := 0; i < 10; i++ {
			if agent.ProcessState.Exited() {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if !agent.ProcessState.Exited() {
			l.Info("Killing pmm-agent...")
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
	runner.run([]string{"run"}, restartPolicy, l)
}
