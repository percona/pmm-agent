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
)

var (
	mysqldVersionRegexp     = regexp.MustCompile("^.*Ver ([!-~]*).*")
	xtrabackupVersionRegexp = regexp.MustCompile("^xtrabackup version ([!-~]*).*")
)

type SoftwareVersioner struct {

}

func NewSoftwareVersion() *SoftwareVersioner {
	return &SoftwareVersioner{}
}

func (*SoftwareVersioner) MySQLServerVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	if _, err := exec.LookPath(mysqldBin); err != nil {
		return "", errors.Wrapf(err, "lookpath: %s", mysqldBin)
	}

	versionBytes, err := exec.CommandContext(ctx, mysqldBin, "--version").CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	matches := mysqldVersionRegexp.FindStringSubmatch(string(versionBytes))
	if len(matches) != 2 {
		return "", errors.Errorf("cannot match version from output %q", string(versionBytes))
	}

	return matches[1], nil
}

func (*SoftwareVersioner) XtrabackupVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	if _, err := exec.LookPath(xtrabackupBin); err != nil {
		return "", errors.Wrapf(err, "lookpath: %s", xtrabackupBin)
	}

	versionBytes, err := exec.CommandContext(ctx, xtrabackupBin, "--version").CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	matches := xtrabackupVersionRegexp.FindStringSubmatch(string(versionBytes))
	if len(matches) != 2 {
		return "", errors.Errorf("cannot match version from output %q", string(versionBytes))
	}

	return matches[1], nil
}
