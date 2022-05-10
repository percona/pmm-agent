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

var (
	pmmAgentProcessID int = 0
)

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

func runPmmAgent(commandLineArgs []string, restartPolicy RestartPolicy, l *logrus.Entry, pmmAgentSidecarSleep int) int {
	pmmAgentFullCommand := "pmm-admin " + strings.Join(commandLineArgs, " ")
	for {
		l.Infof("Starting 'pmm-admin %s'...", strings.Join(commandLineArgs, " "))
		cmd := commandPmmAgent(commandLineArgs)
		if err := cmd.Start(); err != nil {
			l.Errorf("Can't run: '%s', Error: %s", commandLineArgs, err)
			return -1
		}
		var exitCode int
		pmmAgentProcessID = cmd.Process.Pid
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
			l.Infof("Restarting `%s` in %d seconds because PMM_AGENT_SIDECAR is enabled...", pmmAgentFullCommand, pmmAgentSidecarSleep)
			time.Sleep(time.Duration(pmmAgentSidecarSleep) * time.Second)
		} else {
			return exitCode
		}

	}
}

func commandPmmAgent(args []string) *exec.Cmd {
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
		if pmmAgentProcessID != 0 {
			// gracefull shutdown for pmm-agent
			if err := syscall.Kill(pmmAgentProcessID, syscall.SIGTERM); err != nil {
				l.Warn("Failed to send SIGTERM, command must have exited:", err)
			}
		}
		cancel()
		os.Exit(1)
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

	if pmmAgentSetup {
		var agent *exec.Cmd
		restartPolicy := DoNotRestart
		if pmmAgentSidecar {
			restartPolicy = RestartOnFail
			l.Info("Starting pmm-agent for liveness probe...")
			agent = commandPmmAgent([]string{"run"})
			err := agent.Start()
			if err != nil {
				l.Fatalf("Can't run pmm-agent: %s", err)
			}
		}
		statusSetup := runPmmAgent([]string{"setup"}, restartPolicy, l, pmmAgentSidecarSleep)
		if statusSetup != 0 {
			os.Exit(statusSetup)
		}
		if pmmAgentSidecar {
			l.Info("Stopping pmm-agent...")
			if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
				l.Fatal("Failed to kill pmm-agent: ", err)
			}
		}
	}

	status = 0
	if pmmAgentPrerunFile != "" || pmmAgentPrerunScript != "" {
		l.Info("Starting pmm-agent for prerun...")
		agent := commandPmmAgent([]string{"run"})
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
					status = exitError.ExitCode()
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
					status = exitError.ExitCode()
					l.Infof("Prerun shell script exited with %d", exitError.ExitCode())
				}
			}
		}

		l.Info("Stopping pmm-agent...")
		if err := agent.Process.Signal(syscall.SIGTERM); err != nil {
			l.Infof("Failed to term pmm-agent: %s", err)
		}

		// kill pmm-agent process in 10 seconds if SIGTERM doesn't work
		pmmAgentProcessTimeout := 10
		timer := time.AfterFunc(time.Second*time.Duration(pmmAgentProcessTimeout), func() {
			l.Infof("Can't finish pmm-agent process in %d second. Send SIGKILL", pmmAgentProcessTimeout)
			err := agent.Process.Kill()
			if err != nil {
				l.Warnf("Failed to kill pmm-agent: %s", err)
			}
		})

		err = agent.Wait()
		if err != nil {
			exitError, ok := err.(*exec.ExitError)
			if !ok {
				l.Warnf("Can't get exit code for pmm-agent. Error code: %s", err)
			}
			l.Infof("Prerun pmm-agent exited with %d", exitError.ExitCode())
		}
		timer.Stop()

		if status != 0 && !pmmAgentSidecar {
			os.Exit(status)
		}
	}
	restartPolicy := DoNotRestart
	if pmmAgentSidecar {
		restartPolicy = RestartAlways
	}
	runPmmAgent([]string{"run"}, restartPolicy, l, pmmAgentSidecarSleep)
}
