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

// Package versioner implements version retrieving functions for different software.
package versioner

import (
	"context"
	"os/exec"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

const (
	versionCheckTimeout = 5 * time.Second
	mysqldBin           = "mysqld"
	xtrabackupBin       = "xtrabackup"
	xbcloudBin          = "xbcloud"
	qpressBin           = "qpress"
)

var (
	mysqldVersionRegexp     = regexp.MustCompile("^.*Ver ([!-~]*).*")
	xtrabackupVersionRegexp = regexp.MustCompile("^xtrabackup version ([!-~]*).*")
	xbcloudVersionRegexp    = regexp.MustCompile("^xbcloud[ ][ ]Ver ([!-~]*).*")
	qpressRegexp            = regexp.MustCompile("^qpress[ ]([!-~]*).*")

	ErrNotFound = errors.New("not found")
)

type CombinedOutputer interface {
	CombinedOutput() ([]byte, error)
}

//go:generate mockery -name=ExecFunctions -case=snake -inpkg -testonly
type ExecFunctions interface {
	LookPath(file string) (string, error)
	CommandContext(ctx context.Context, name string, arg ...string) CombinedOutputer
}

type RealExecFunctions struct{}

func (RealExecFunctions) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (RealExecFunctions) CommandContext(ctx context.Context, name string, arg ...string) CombinedOutputer {
	return exec.CommandContext(ctx, name, arg...)
}

type Versioner struct {
	ef ExecFunctions
}

func New(ef ExecFunctions) *Versioner {
	return &Versioner{
		ef: ef,
	}
}

func (v *Versioner) binaryVersion(binaryName string, versionRegexp *regexp.Regexp, arg ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	if _, err := v.ef.LookPath(binaryName); err != nil {
		if err.(*exec.Error).Err == exec.ErrNotFound {
			return "", ErrNotFound
		}

		return "", errors.Wrapf(err, "lookpath: %s", binaryName)
	}

	versionBytes, err := v.ef.CommandContext(ctx, binaryName, arg...).CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	matches := versionRegexp.FindStringSubmatch(string(versionBytes))
	if len(matches) != 2 {
		return "", errors.Errorf("cannot match version from output %q", string(versionBytes))
	}

	return matches[1], nil
}

func (v *Versioner) MySQLdVersion() (string, error) {
	return v.binaryVersion(mysqldBin, mysqldVersionRegexp, "--version")
}

func (v *Versioner) XtrabackupVersion() (string, error) {
	return v.binaryVersion(xtrabackupBin, xtrabackupVersionRegexp, "--version")
}

func (v *Versioner) XbcloudVersion() (string, error) {
	return v.binaryVersion(xbcloudBin, xbcloudVersionRegexp, "--version")
}

func (v *Versioner) Qpress() (string, error) {
	return v.binaryVersion(qpressBin, qpressRegexp)
}
