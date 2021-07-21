package versioner

import (
	"os/exec"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedExec struct {
	Output []byte
}

func (m *mockedExec) CombinedOutput() ([]byte, error) {
	return m.Output, nil
}

func TestVersioner(t *testing.T) {
	execMock := &MockExecFunctions{}
	versioner := New(execMock)

	t.Run("not found", func(t *testing.T) {
		execMock.On("LookPath", mysqldBin).Return("", &exec.Error{Err: exec.ErrNotFound}).Once()

		version, err := versioner.MySQLdVersion()
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Equal(t, "", version)
	})
	t.Run("mysqld", func(t *testing.T) {
		mysqldVersionOutput := []byte(`/usr/sbin/mysqld  Ver 8.0.22-13 for Linux on x86_64 (Percona Server (GPL), Release '13', Revision '6f7822f')
`)
		execMock.On("LookPath", mysqldBin).Return("", nil).Once()
		execMock.On("CommandContext", mock.Anything, mysqldBin, "--version").
			Return(&mockedExec{Output: mysqldVersionOutput}).Once()
		version, err := versioner.MySQLdVersion()
		assert.NoError(t, err)
		assert.Equal(t, "8.0.22-13", version)
	})
	t.Run("xtrabackup", func(t *testing.T) {
		mysqldVersionOutput := []byte(`xtrabackup version 8.0.23-16 based on MySQL server 8.0.23 Linux (x86_64) (revision id: 934bc8f)
`)
		execMock.On("LookPath", xtrabackupBin).Return("", nil).Once()
		execMock.On("CommandContext", mock.Anything, xtrabackupBin, "--version").
			Return(&mockedExec{Output: mysqldVersionOutput}).Once()
		version, err := versioner.XtrabackupVersion()
		assert.NoError(t, err)
		assert.Equal(t, "8.0.23-16", version)
	})
	t.Run("xbcloud", func(t *testing.T) {
		mysqldVersionOutput := []byte(`xbcloud  Ver 8.0.23-16 for Linux (x86_64) (revision id: 934bc8f)
`)
		execMock.On("LookPath", xbcloudBin).Return("", nil).Once()
		execMock.On("CommandContext", mock.Anything, xbcloudBin, "--version").
			Return(&mockedExec{Output: mysqldVersionOutput}).Once()
		version, err := versioner.XbcloudVersion()
		assert.NoError(t, err)
		assert.Equal(t, "8.0.23-16", version)
	})
	t.Run("qpress", func(t *testing.T) {
		mysqldVersionOutput := []byte(`qpress 1.1 - Copyright 2006-2010 Lasse Reinhold - www.quicklz.com
Using QuickLZ 1.4.1 compression library
Compiled for: Windows [*nix]    [x86/x64] RISC    32-bit [64-bit]
...
`)
		execMock.On("LookPath", qpressBin).Return("", nil).Once()
		execMock.On("CommandContext", mock.Anything, qpressBin).
			Return(&mockedExec{Output: mysqldVersionOutput}).Once()
		version, err := versioner.Qpress()
		assert.NoError(t, err)
		assert.Equal(t, "1.1", version)
	})

	mock.AssertExpectationsForObjects(t, execMock)
}
