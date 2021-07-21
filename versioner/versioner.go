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
